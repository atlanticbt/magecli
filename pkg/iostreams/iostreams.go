package iostreams

import (
	"io"
	"os"
	"sync"

	"golang.org/x/term"
)

type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	isStdinTTY  bool
	isStdoutTTY bool
	isStderrTTY bool

	colorEnabled bool
	once         sync.Once
}

func System() *IOStreams {
	isTTY := func(f *os.File) bool {
		if f == nil {
			return false
		}
		return term.IsTerminal(int(f.Fd()))
	}

	return &IOStreams{
		In:          os.Stdin,
		Out:         os.Stdout,
		ErrOut:      os.Stderr,
		isStdinTTY:  isTTY(os.Stdin),
		isStdoutTTY: isTTY(os.Stdout),
		isStderrTTY: isTTY(os.Stderr),
	}
}

func (s *IOStreams) CanPrompt() bool {
	return s != nil && s.isStdinTTY
}

func (s *IOStreams) ColorEnabled() bool {
	if s == nil {
		return false
	}
	s.once.Do(func() {
		s.colorEnabled = s.isStdoutTTY
	})
	return s.colorEnabled
}

func (s *IOStreams) IsStdoutTTY() bool {
	return s != nil && s.isStdoutTTY
}

func (s *IOStreams) IsStderrTTY() bool {
	return s != nil && s.isStderrTTY
}
