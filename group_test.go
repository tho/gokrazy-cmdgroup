package main_test

import (
	"bytes"
	"context"
	"log/slog"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmdgroup "github.com/tho/gokrazy-cmdgroup"
)

// TestNew tests creating a new Group with various options.
func TestNew(t *testing.T) {
	t.Parallel()

	cmdName := "echo"
	cmdPath, err := exec.LookPath(cmdName)
	require.NoError(t, err)
	discardLogger := slog.New(slog.DiscardHandler)

	tests := map[string]struct {
		cmdName       string
		options       []cmdgroup.Option
		wantInstances []*cmdgroup.Instance
		wantErr       assert.ErrorAssertionFunc
	}{
		"empty command name": {
			cmdName: "",
			wantErr: assert.Error,
		},
		"invalid watch value": {
			cmdName: cmdName,
			options: []cmdgroup.Option{cmdgroup.WithWatch("a")},
			wantErr: assert.Error,
		},
		"nil logger": {
			cmdName:       cmdName,
			options:       []cmdgroup.Option{cmdgroup.WithLogger(nil)},
			wantInstances: nil,
			wantErr:       assert.Error,
		},
		"nil options": {
			cmdName: cmdName,
			options: nil,
			wantInstances: []*cmdgroup.Instance{
				{Name: cmdPath, Args: nil, Logger: discardLogger},
			},
			wantErr: assert.NoError,
		},
		"single instance with args": {
			cmdName: cmdName,
			options: []cmdgroup.Option{cmdgroup.WithArgs([]string{"arg1", "-flag1"})},
			wantInstances: []*cmdgroup.Instance{
				{Name: cmdPath, Args: []string{"arg1", "-flag1"}, Logger: discardLogger},
			},
			wantErr: assert.NoError,
		},
		"two instances no global args": {
			cmdName: cmdName,
			options: []cmdgroup.Option{
				cmdgroup.WithArgs([]string{"--", "arg1", "-flag1", "--", "arg2", "-flag2"}),
			},
			wantInstances: []*cmdgroup.Instance{
				{Name: cmdPath, Args: []string{"arg1", "-flag1"}, Logger: discardLogger},
				{Name: cmdPath, Args: []string{"arg2", "-flag2"}, Logger: discardLogger},
			},
			wantErr: assert.NoError,
		},
		"two instances with global args": {
			cmdName: cmdName,
			options: []cmdgroup.Option{
				cmdgroup.WithArgs([]string{"-v", "--", "arg1", "--", "arg2"}),
				cmdgroup.WithWatch("all"),
			},
			wantInstances: []*cmdgroup.Instance{
				{Name: cmdPath, Args: []string{"-v", "arg1"}, Watch: true, Logger: discardLogger},
				{Name: cmdPath, Args: []string{"-v", "arg2"}, Watch: true, Logger: discardLogger},
			},
			wantErr: assert.NoError,
		},
		"watch selective": {
			cmdName: cmdName,
			options: []cmdgroup.Option{
				cmdgroup.WithArgs([]string{"--", "arg1", "-flag1", "--", "arg2", "-flag2"}),
				cmdgroup.WithWatch("1"),
			},
			wantInstances: []*cmdgroup.Instance{
				{Name: cmdPath, Args: []string{"arg1", "-flag1"}, Watch: false, Logger: discardLogger},
				{Name: cmdPath, Args: []string{"arg2", "-flag2"}, Watch: true, Logger: discardLogger},
			},
			wantErr: assert.NoError,
		},
		"watch index out of range": {
			cmdName: cmdName,
			options: []cmdgroup.Option{
				cmdgroup.WithArgs([]string{"--", "arg1", "--", "arg2"}),
				cmdgroup.WithWatch("2"),
			},
			wantErr: assert.Error,
		},
		"watch negative index": {
			cmdName: cmdName,
			options: []cmdgroup.Option{
				cmdgroup.WithArgs([]string{"--", "arg1", "--", "arg2"}),
				cmdgroup.WithWatch("-1"),
			},
			wantErr: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			group, err := cmdgroup.New(tt.cmdName, tt.options...)
			tt.wantErr(t, err)
			if group != nil {
				assert.Equal(t, tt.wantInstances, group.Instances)
			}
		})
	}
}

// TestGroupRun tests running a Group of instances.
func TestGroupRun(t *testing.T) {
	t.Parallel()

	truePath, err := exec.LookPath("true")
	require.NoError(t, err)
	falsePath, err := exec.LookPath("false")
	require.NoError(t, err)
	sleepPath, err := exec.LookPath("sleep")
	require.NoError(t, err)

	logger := slog.New(slog.DiscardHandler)

	tests := map[string]struct {
		instances []*cmdgroup.Instance
		wantErr   assert.ErrorAssertionFunc
	}{
		"all succeed": {
			instances: []*cmdgroup.Instance{
				{Name: truePath, Logger: logger},
				{Name: truePath, Logger: logger},
			},
			wantErr: assert.NoError,
		},
		"one fails": {
			instances: []*cmdgroup.Instance{
				{Name: truePath, Logger: logger},
				{Name: falsePath, Logger: logger},
			},
			wantErr: assert.Error,
		},
		"unwatched failure cancels watched instance": {
			instances: []*cmdgroup.Instance{
				{Name: falsePath, Logger: logger},
				{Name: sleepPath, Args: []string{"60"}, Watch: true, Logger: logger},
			},
			wantErr: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			group := &cmdgroup.Group{Instances: tt.instances}
			err := group.Run(t.Context())
			tt.wantErr(t, err)
		})
	}
}

// TestInstanceRun tests running a single Instance.
func TestInstanceRun(t *testing.T) {
	t.Parallel()

	truePath, err := exec.LookPath("true")
	require.NoError(t, err)
	falsePath, err := exec.LookPath("false")
	require.NoError(t, err)
	sleepPath, err := exec.LookPath("sleep")
	require.NoError(t, err)

	tests := map[string]struct {
		cmdPath         string
		args            []string
		watch           bool
		ctx             func(*testing.T) context.Context
		wantErr         require.ErrorAssertionFunc
		wantErrIs       error
		wantErrContains string
		wantLog         string
	}{
		"unwatched success": {
			cmdPath: truePath,
			wantErr: require.NoError,
		},
		"unwatched failure": {
			cmdPath: falsePath,
			wantErr: require.Error,
		},
		"start error": {
			cmdPath:         "/nonexistent/binary",
			wantErr:         require.Error,
			wantErrContains: "start command",
		},
		"watched restarts": {
			cmdPath: truePath,
			watch:   true,
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 1500*time.Millisecond)
				t.Cleanup(cancel)

				return ctx
			},
			wantErr:   require.Error,
			wantErrIs: context.DeadlineExceeded,
			wantLog:   "msg=restarting",
		},
		"watched context cancel stops restart": {
			cmdPath: sleepPath,
			args:    []string{"60"},
			watch:   true,
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithCancel(t.Context())
				go func() {
					time.Sleep(100 * time.Millisecond)
					cancel()
				}()

				return ctx
			},
			wantErr:   require.Error,
			wantErrIs: context.Canceled,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := slog.New(slog.DiscardHandler)
			if tt.wantLog != "" {
				logger = slog.New(slog.NewTextHandler(&buf, nil))
			}

			instance := &cmdgroup.Instance{
				Name:   tt.cmdPath,
				Args:   tt.args,
				Watch:  tt.watch,
				Logger: logger,
			}

			ctx := t.Context()
			if tt.ctx != nil {
				ctx = tt.ctx(t)
			}

			err := instance.Run(ctx)
			tt.wantErr(t, err)

			if tt.wantErrIs != nil {
				require.ErrorIs(t, err, tt.wantErrIs)
			}
			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			}
			if tt.wantLog != "" {
				require.Contains(t, buf.String(), tt.wantLog)
			}
		})
	}
}
