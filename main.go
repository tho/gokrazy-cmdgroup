// Package main implements cmdgroup, which runs multiple instances of the same
// command with different arguments within the gokrazy ecosystem.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const gokrazyDoNotSuperviseExitCode = 125

func main() {
	ctx := context.Background()
	os.Exit(run(ctx, os.Args))
}

func run(ctx context.Context, args []string) int {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	flagSet := flag.NewFlagSet("cmdgroup", flag.ContinueOnError)
	watch := flagSet.String("watch", "none", "watch none, all, or 0,1,... instances")
	if err := flagSet.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		logger.ErrorContext(ctx, "parsing flags", "error", err)

		return gokrazyDoNotSuperviseExitCode
	}

	positional := flagSet.Args()
	if len(positional) == 0 {
		logger.ErrorContext(ctx, "no command specified")
		return gokrazyDoNotSuperviseExitCode
	}

	group, err := New(
		positional[0],
		WithArgs(positional[1:]),
		WithWatch(*watch),
		WithLogger(logger),
	)
	if err != nil {
		logger.ErrorContext(ctx, "creating new command group", "error", err)
		return gokrazyDoNotSuperviseExitCode
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := group.Run(ctx); err != nil {
		logger.ErrorContext(ctx, "running command group", "error", err)
		return 1
	}

	return 0
}
