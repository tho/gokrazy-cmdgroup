package main

import (
	"context"
	"errors"
	"os/exec"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheckErr tests error classification for process exit errors.
func TestCheckErr(t *testing.T) {
	t.Parallel()

	// signalErr returns an error from a process killed by the given signal.
	signalErr := func(sig syscall.Signal) func(*testing.T) error {
		return func(t *testing.T) error {
			t.Helper()
			cmd := exec.CommandContext(t.Context(), "sleep", "60")
			if err := cmd.Start(); err != nil {
				t.Fatalf("start: %v", err)
			}
			if err := cmd.Process.Signal(sig); err != nil {
				t.Fatalf("signal: %v", err)
			}

			return cmd.Wait()
		}
	}

	tests := map[string]struct {
		err     func(*testing.T) error
		wantErr assert.ErrorAssertionFunc
	}{
		"nil": {
			err:     func(*testing.T) error { return nil },
			wantErr: assert.NoError,
		},
		"context canceled": {
			err:     func(*testing.T) error { return context.Canceled },
			wantErr: assert.NoError,
		},
		"non-exit error": {
			err:     func(*testing.T) error { return errors.New("some error") },
			wantErr: assert.Error,
		},
		"exit code": {
			err: func(t *testing.T) error {
				t.Helper()
				return exec.CommandContext(t.Context(), "false").Run()
			},
			wantErr: assert.Error,
		},
		"sigterm": {
			err:     signalErr(syscall.SIGTERM),
			wantErr: assert.NoError,
		},
		"sigkill": {
			err:     signalErr(syscall.SIGKILL),
			wantErr: assert.Error,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tt.wantErr(t, checkErr(tt.err(t)))
		})
	}
}
