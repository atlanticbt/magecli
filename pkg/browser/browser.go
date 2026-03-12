package browser

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
)

type Browser interface {
	Open(url string) error
}

type system struct{}

func NewSystem() Browser { return &system{} }

func (s *system) Open(url string) error {
	if url == "" {
		return errors.New("url is required")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch browser: %w", err)
	}
	return cmd.Wait()
}
