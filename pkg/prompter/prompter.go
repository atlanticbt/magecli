package prompter

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/atlanticbt/magecli/pkg/iostreams"
	"golang.org/x/term"
)

type Interface interface {
	Input(prompt, defaultValue string) (string, error)
	Password(prompt string) (string, error)
	Confirm(prompt string, defaultYes bool) (bool, error)
}

type system struct {
	ios *iostreams.IOStreams
}

func New(ios *iostreams.IOStreams) Interface {
	return &system{ios: ios}
}

func (p *system) reader() (*bufio.Reader, error) {
	if p.ios == nil || !p.ios.CanPrompt() {
		return nil, errors.New("interactive prompts require a TTY")
	}
	return bufio.NewReader(p.ios.In), nil
}

func (p *system) Input(prompt, defaultValue string) (string, error) {
	r, err := p.reader()
	if err != nil {
		return "", err
	}
	question := prompt
	if defaultValue != "" {
		question = fmt.Sprintf("%s [%s]", prompt, defaultValue)
	}
	if _, err := fmt.Fprint(p.ios.Out, question+": "); err != nil {
		return "", err
	}
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue, nil
	}
	return line, nil
}

func (p *system) Password(prompt string) (string, error) {
	if p.ios == nil || !p.ios.CanPrompt() {
		return "", errors.New("interactive prompts require a TTY")
	}
	if _, err := fmt.Fprint(p.ios.Out, prompt+": "); err != nil {
		return "", err
	}
	stdin, ok := p.ios.In.(*os.File)
	if !ok {
		return "", errors.New("password input requires a terminal")
	}
	password, err := term.ReadPassword(int(stdin.Fd()))
	_, _ = fmt.Fprintln(p.ios.Out)
	if err != nil {
		return "", err
	}
	return string(password), nil
}

func (p *system) Confirm(prompt string, defaultYes bool) (bool, error) {
	r, err := p.reader()
	if err != nil {
		return false, err
	}
	var suffix string
	if defaultYes {
		suffix = "[Y/n]"
	} else {
		suffix = "[y/N]"
	}
	for {
		if _, err := fmt.Fprintf(p.ios.Out, "%s %s: ", prompt, suffix); err != nil {
			return false, err
		}
		line, err := r.ReadString('\n')
		if err != nil {
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		case "":
			return defaultYes, nil
		}
	}
}
