package browser

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
)

type Browser interface {
	Open(url string) error
}

type system struct{}

func NewSystem() Browser { return &system{} }

func (s *system) Open(rawURL string) error {
	if rawURL == "" {
		return errors.New("url is required")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsafe URL scheme %q; only http and https are allowed", u.Scheme)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch browser: %w", err)
	}
	return cmd.Wait()
}
