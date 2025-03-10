package cmdgroup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"sync"
	"syscall"
	"time"
)

// cmdWaitDelay is how long to wait for an instance to exit after SIGTERM.
// After this time, the process is killed.
const cmdWaitDelay = 10 * time.Second

// Group manages multiple command instances.
type Group struct {
	Instances []*Instance

	logger *slog.Logger
}

// Instance represents a single command execution with its configuration.
type Instance struct {
	Name  string
	Args  []string
	Watch bool

	logger *slog.Logger
}

// Options holds configuration for creating a new group.
type Options struct {
	args   []string
	watch  string
	logger *slog.Logger
}

// Option is a functional option for configuring a group.
type Option func(*Options)

// WithArgs sets the command arguments for the group.
func WithArgs(args []string) Option {
	return func(o *Options) {
		o.args = args
	}
}

// WithWatch sets which command instances should be monitored and restarted.
func WithWatch(watch string) Option {
	return func(o *Options) {
		o.watch = watch
	}
}

// WithLogger sets the logger for the group and its instances.
func WithLogger(logger *slog.Logger) Option {
	return func(o *Options) {
		o.logger = logger
	}
}

// New creates a command group for the specified command name and options.
// By default, no instances are watched and no logging is performed.
func New(name string, options ...Option) (*Group, error) {
	opts := &Options{
		watch:  "none",
		logger: slog.New(slog.DiscardHandler),
	}
	for _, option := range options {
		option(opts)
	}

	path, err := exec.LookPath(name)
	if err != nil {
		return nil, fmt.Errorf("look path: %w", err)
	}

	instances := parseArgs(opts.args)

	watchIndexes, err := parseInts(opts.watch, len(instances))
	if err != nil {
		return nil, fmt.Errorf("parse watch: %w", err)
	}

	for index, instance := range instances {
		instance.Name = path
		instance.logger = opts.logger

		if slices.Contains(watchIndexes, index) {
			instances[index].Watch = true
		}
	}

	return &Group{Instances: instances, logger: opts.logger}, nil
}

// Run executes all command instances in parallel and waits for them to complete.
func (g *Group) Run(ctx context.Context) error {
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs error
	)

	for _, instance := range g.Instances {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := instance.Run(ctx)
			mu.Lock()
			errs = errors.Join(errs, checkErr(err))
			mu.Unlock()
		}()
	}

	wg.Wait()

	return errs
}

// checkErr filters out expected termination errors (context cancel, SIGTERM).
func checkErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.Canceled) {
		return nil
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return err
	}

	waitStatus, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		return err
	}

	if !waitStatus.Signaled() || waitStatus.Signal() != syscall.SIGTERM {
		return err
	}

	return nil
}

// Run executes this command instance, potentially restarting it if configured to watch.
func (i *Instance) Run(ctx context.Context) error {
	var err error

	for {
		cmd := i.newCmd(ctx)
		cmdLogger := i.logger.With("cmd", cmd.String())

		if err = cmd.Start(); err != nil {
			i.logger.Error("start command", "error", err)
		} else {
			cmdLogger = cmdLogger.With("pid", cmd.Process.Pid)
			cmdLogger.Info("started")

			if err = cmd.Wait(); err != nil {
				cmdLogger.Info("exited", "reason", err)
			} else {
				cmdLogger.Info("exited")
			}
		}

		if !i.Watch {
			return err
		}

		select {
		case <-ctx.Done():
			cmdLogger.Info("not restarting", "reason", ctx.Err())
			return ctx.Err()
		case <-time.After(time.Second):
			cmdLogger.Info("restarting")
		}
	}
}

// newCmd creates a new exec.Cmd with process group handling for clean termination.
func (i *Instance) newCmd(ctx context.Context) *exec.Cmd {
	// #nosec G204 -- user/caller is responsible for name and args
	cmd := exec.CommandContext(ctx, i.Name, i.Args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error {
		// Signal entire process group on termination.
		if pgid, err := syscall.Getpgid(cmd.Process.Pid); err == nil {
			return syscall.Kill(-pgid, syscall.SIGTERM)
		}
		// Fallback to single process termination.
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	cmd.WaitDelay = cmdWaitDelay
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group.
	}

	return cmd
}
