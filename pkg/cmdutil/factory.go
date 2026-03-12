package cmdutil

import (
	"sync"

	"github.com/atlanticbt/magecli/internal/config"
	"github.com/atlanticbt/magecli/pkg/browser"
	"github.com/atlanticbt/magecli/pkg/iostreams"
	"github.com/atlanticbt/magecli/pkg/pager"
	"github.com/atlanticbt/magecli/pkg/progress"
	"github.com/atlanticbt/magecli/pkg/prompter"
)

type Factory struct {
	AppVersion     string
	ExecutableName string

	IOStreams *iostreams.IOStreams

	Config func() (*config.Config, error)

	Browser  browser.Browser
	Pager    pager.Manager
	Prompter prompter.Interface
	Spinner  progress.Spinner

	once struct {
		cfg sync.Once
	}
	cfg    *config.Config
	cfgErr error
	ioOnce sync.Once
	ios    *iostreams.IOStreams
}

func (f *Factory) ResolveConfig() (*config.Config, error) {
	f.once.cfg.Do(func() {
		if f.Config == nil {
			f.cfg, f.cfgErr = config.Load()
			return
		}
		f.cfg, f.cfgErr = f.Config()
	})
	return f.cfg, f.cfgErr
}

func (f *Factory) Streams() (*iostreams.IOStreams, error) {
	f.ioOnce.Do(func() {
		if f.IOStreams != nil {
			f.ios = f.IOStreams
			return
		}
		f.ios = iostreams.System()
	})
	return f.ios, nil
}

func (f *Factory) BrowserOpener() browser.Browser {
	if f.Browser == nil {
		f.Browser = browser.NewSystem()
	}
	return f.Browser
}

func (f *Factory) PagerManager() pager.Manager {
	if f.Pager == nil {
		ios, _ := f.Streams()
		f.Pager = pager.NewSystem(ios)
	}
	return f.Pager
}

func (f *Factory) Prompt() prompter.Interface {
	if f.Prompter == nil {
		ios, _ := f.Streams()
		f.Prompter = prompter.New(ios)
	}
	return f.Prompter
}

func (f *Factory) ProgressSpinner() progress.Spinner {
	if f.Spinner == nil {
		ios, _ := f.Streams()
		f.Spinner = progress.NewSpinner(ios)
	}
	return f.Spinner
}
