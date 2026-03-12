package factory

import (
	"github.com/atlanticbt/magecli/internal/config"
	"github.com/atlanticbt/magecli/pkg/browser"
	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/iostreams"
	"github.com/atlanticbt/magecli/pkg/pager"
	"github.com/atlanticbt/magecli/pkg/progress"
	"github.com/atlanticbt/magecli/pkg/prompter"
)

func New(appVersion string) (*cmdutil.Factory, error) {
	ios := iostreams.System()

	f := &cmdutil.Factory{
		AppVersion:     appVersion,
		ExecutableName: "magecli",
		IOStreams:      ios,
	}

	f.Browser = browser.NewSystem()
	f.Pager = pager.NewSystem(ios)
	f.Prompter = prompter.New(ios)
	f.Spinner = progress.NewSpinner(ios)

	f.Config = config.Load

	return f, nil
}
