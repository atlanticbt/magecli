package magecmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/atlanticbt/magecli/internal/build"
	"github.com/atlanticbt/magecli/pkg/cmd/factory"
	"github.com/atlanticbt/magecli/pkg/cmd/root"
	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/httpx"
)

// Exit codes. Agents can branch on these instead of parsing error text.
const (
	exitOK       = 0
	exitGeneric  = 1 // usage errors, bad input, config problems
	exitNetwork  = 2 // connection refused, DNS, timeout
	exitNotFound = 3 // HTTP 404
	exitAuth     = 4 // HTTP 401/403
	exitHTTP     = 5 // any other non-2xx HTTP response
	exitPending  = 8 // ErrPending sentinel
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
			return exitPending
		}
		if errors.Is(err, cmdutil.ErrSilent) {
			return exitGeneric
		}

		_, _ = fmt.Fprintf(ios.ErrOut, "Error: %v\n", err)

		var httpErr *httpx.HTTPError
		if errors.As(err, &httpErr) {
			switch {
			case httpErr.StatusCode == 401 || httpErr.StatusCode == 403:
				_, _ = fmt.Fprintf(ios.ErrOut,
					"hint: authentication failed — run `%s auth login`, or verify the Integration token's resource access in Magento Admin > System > Integrations.\n",
					f.ExecutableName)
				return exitAuth
			case httpErr.StatusCode == 404:
				return exitNotFound
			default:
				return exitHTTP
			}
		}

		// User-initiated cancellation (Ctrl-C/SIGTERM) is not a network failure.
		if errors.Is(err, context.Canceled) {
			return exitGeneric
		}

		// Network-level failure (connection refused, DNS, timeout) — distinct
		// from an HTTP error response. Match concrete network error types
		// rather than *url.Error, which also wraps URL parse failures of bad
		// user input that belong at exitGeneric.
		var opErr *net.OpError
		var dnsErr *net.DNSError
		if errors.As(err, &opErr) || errors.As(err, &dnsErr) {
			return exitNetwork
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return exitNetwork
		}

		return exitGeneric
	}

	return exitOK
}
