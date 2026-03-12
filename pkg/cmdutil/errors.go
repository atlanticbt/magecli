package cmdutil

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	ErrSilent  = errors.New("silent")
	ErrPending = errors.New("pending")
)

type ExitError struct {
	Code int
	Msg  string
}

func (e *ExitError) Error() string { return e.Msg }

func NotImplemented(cmd *cobra.Command) error {
	return fmt.Errorf("%s not yet implemented", cmd.CommandPath())
}
