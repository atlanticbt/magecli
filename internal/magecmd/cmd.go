package magecmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/atlanticbt/magecli/internal/build"
	"github.com/atlanticbt/magecli/pkg/cmd/factory"
	"github.com/atlanticbt/magecli/pkg/cmd/root"
	"github.com/atlanticbt/magecli/pkg/cmdutil"
)

func Main() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	f, err := factory.New(build.Version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialise factory: %v\n", err)
		return 1
	}

	ios, err := f.Streams()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to configure IO: %v\n", err)
		return 1
	}

	rootCmd, err := root.NewCmdRoot(f)
	if err != nil {
		_, _ = fmt.Fprintf(ios.ErrOut, "failed to create root command: %v\n", err)
		return 1
	}
	rootCmd.SetContext(ctx)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		var exitErr *cmdutil.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Msg != "" {
				_, _ = fmt.Fprintln(ios.ErrOut, exitErr.Msg)
			}
			return exitErr.Code
		}
		if errors.Is(err, cmdutil.ErrPending) {
			return 8
		}
		if errors.Is(err, cmdutil.ErrSilent) {
			return 1
		}
		_, _ = fmt.Fprintf(ios.ErrOut, "Error: %v\n", err)
		return 1
	}

	return 0
}
