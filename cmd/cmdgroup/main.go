package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	cmdgroup "github.com/tho/gokrazy-cmdgroup"
)

const gokrazyDoNotSuperviseExitCode = 125

func main() {
	ctx := context.Background()
	os.Exit(run(ctx, os.Args))
}

func run(ctx context.Context, args []string) int {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	flagSet := flag.NewFlagSet("cmdgroup", flag.ExitOnError)
	var (
		name  = flagSet.String("name", "echo", "name of the command")
		watch = flagSet.String("watch", "none", "watch none, all, or 0,1,2,... instances")
	)
	flagSet.Parse(args[1:]) //nolint:errcheck // using ExitOnError

	group, err := cmdgroup.New(
		*name,
		cmdgroup.WithArgs(flagSet.Args()),
		cmdgroup.WithWatch(*watch),
		cmdgroup.WithLogger(logger),
	)
	if err != nil {
		logger.Error("crating new command group", "error", err)
		return gokrazyDoNotSuperviseExitCode
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := group.Run(ctx); err != nil {
		logger.Error("running command group", "error", err)
		return 1
	}

	return 0
}
