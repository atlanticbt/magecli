package pager

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/atlanticbt/magecli/pkg/iostreams"
)

type Manager interface {
	Enabled() bool
	Start() (io.WriteCloser, error)
	Stop() error
}

type systemPager struct {
	ios    *iostreams.IOStreams
	cmd    *exec.Cmd
	writer io.WriteCloser
}

func NewSystem(ios *iostreams.IOStreams) Manager {
	if ios == nil || !ios.IsStdoutTTY() {
		return noop{}
	}
	return &systemPager{ios: ios}
}

func (p *systemPager) Enabled() bool { return true }

func (p *systemPager) Start() (io.WriteCloser, error) {
	if p.writer != nil {
		return p.writer, nil
	}
	pagerCmd := strings.Fields(resolvePager())
	cmd := exec.Command(pagerCmd[0], pagerCmd[1:]...)
	cmd.Stdout = p.ios.Out
	cmd.Stderr = p.ios.ErrOut
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		_ = in.Close()
		return nil, err
	}
	p.cmd = cmd
	p.writer = in
	return in, nil
}

func (p *systemPager) Stop() error {
	if p.writer != nil {
		_ = p.writer.Close()
		p.writer = nil
	}
	if p.cmd != nil {
		err := p.cmd.Wait()
		p.cmd = nil
		return err
	}
	return nil
}

type noop struct{}

func (noop) Enabled() bool                        { return false }
func (noop) Start() (io.WriteCloser, error)        { return nopWriteCloser{Writer: os.Stdout}, nil }
func (noop) Stop() error                           { return nil }

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

func resolvePager() string {
	if cmd := os.Getenv("MAGECLI_PAGER"); cmd != "" {
		return cmd
	}
	if cmd := os.Getenv("PAGER"); cmd != "" {
		return cmd
	}
	return "less -R"
}
