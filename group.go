package main

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

type (
	// Group manages multiple command instances.
	Group struct {
		Instances []*Instance
	}

	// Instance represents a single command execution with its configuration.
	Instance struct {
		Name   string
		Args   []string
		Watch  bool
		Logger *slog.Logger
	}

	// Options holds configuration for creating a new group.
	Options struct {
		args   []string
		watch  string
		logger *slog.Logger
	}

	// Option is a functional option for configuring a group.
	Option func(*Options)
)

const (
	// cmdRestartDelay is how long to wait before restarting a watched
	// instance.
	cmdRestartDelay = time.Second

	// cmdWaitDelay is how long to wait for an instance to exit after
	// SIGTERM.  After this time, the process is killed.
	cmdWaitDelay = 10 * time.Second
)

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

// WithLogger sets the logger for all instances in the group.
func WithLogger(logger *slog.Logger) Option {
	return func(o *Options) {
		o.logger = logger
	}
}

// New creates a command group for the specified command name and options.
// Arguments before the first "--" separator are global args prepended to every
// instance. Each "--"-delimited section after that defines a separate instance.
// If no "--" separators are present, a single instance receives all arguments.
// By default, no instances are watched and no logging is performed.
func New(name string, options ...Option) (*Group, error) {
	opts := &Options{
		args:   nil,
		watch:  "none",
		logger: slog.New(slog.DiscardHandler),
	}
	for _, option := range options {
		option(opts)
	}
	if opts.logger == nil {
		return nil, errors.New("nil logger")
	}

	path, err := exec.LookPath(name)
	if err != nil {
		return nil, fmt.Errorf("look path: %w", err)
	}

	var (
		instances  []*Instance
		args       = parseArgs(opts.args)
		globalArgs = args[0] // parseArgs always returns at least one element
	)
	for _, args := range args[1:] {
		instances = append(instances, &Instance{
			Name:   path,
			Args:   slices.Concat(globalArgs, args),
			Watch:  false,
			Logger: opts.logger,
		})
	}
	if len(instances) == 0 {
		instances = append(instances, &Instance{
			Name:   path,
			Args:   globalArgs,
			Watch:  false,
			Logger: opts.logger,
		})
	}

	if err := applyWatch(instances, opts.watch); err != nil {
		return nil, err
	}

	return &Group{Instances: instances}, nil
}

// applyWatch configures which instances should be monitored and restarted.
func applyWatch(instances []*Instance, watch string) error {
	switch watch {
	case "", "none":
		return nil
	case "all":
		for i := range instances {
			instances[i].Watch = true
		}

		return nil
	default:
		indexes, err := parseInts(watch)
		if err != nil {
			return fmt.Errorf("parse watch: %w", err)
		}

		for _, index := range indexes {
			if index < 0 || index >= len(instances) {
				return fmt.Errorf("parse watch: index out of range: %d", index)
			}

			instances[index].Watch = true
		}

		return nil
	}
}

// Run executes all command instances in parallel and waits for them to complete.
func (g *Group) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	var wg sync.WaitGroup

	errs := make([]error, len(g.Instances))
	for idx, instance := range g.Instances {
		wg.Go(func() {
			errs[idx] = checkErr(instance.Run(ctx))
			if errs[idx] != nil && !instance.Watch {
				cancel(errs[idx])
			}
		})
	}

	wg.Wait()

	return errors.Join(errs...)
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

// Run executes this command instance, potentially restarting it if configured
// to watch.
func (i *Instance) Run(ctx context.Context) error {
	var err error

	logger := i.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	for {
		cmd := i.newCmd(ctx)
		cmdLogger := logger.With("cmd", cmd.String())

		if err = cmd.Start(); err != nil {
			return fmt.Errorf("start command: %w", err)
		}

		cmdLogger = cmdLogger.With("pid", cmd.Process.Pid)
		cmdLogger.InfoContext(ctx, "started")

		if err = cmd.Wait(); err != nil {
			cmdLogger.ErrorContext(ctx, "exited", "reason", err)
		} else {
			cmdLogger.InfoContext(ctx, "exited")
		}

		if !i.Watch {
			return err
		}

		select {
		case <-ctx.Done():
			cmdLogger.InfoContext(ctx, "not restarting", "reason", ctx.Err())
			return ctx.Err()
		case <-time.After(cmdRestartDelay):
			cmdLogger.InfoContext(ctx, "restarting")
		}
	}
}

// newCmd creates a new [exec.Cmd] with process group handling for clean termination.
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
