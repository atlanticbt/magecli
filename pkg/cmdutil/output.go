package cmdutil

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/format"
)

type OutputSettings struct {
	Format   string
	JQ       string
	Template string
}

func ResolveOutputSettings(cmd *cobra.Command) (OutputSettings, error) {
	root := cmd.Root()
	lookup := func(name string) string {
		flag := root.PersistentFlags().Lookup(name)
		if flag == nil {
			return ""
		}
		return flag.Value.String()
	}

	jsonEnabled := lookup("json") == "true"
	yamlEnabled := lookup("yaml") == "true"
	jqExpr := lookup("jq")
	tmpl := lookup("template")

	if jsonEnabled && yamlEnabled {
		return OutputSettings{}, fmt.Errorf("cannot use --json and --yaml simultaneously")
	}
	if jqExpr != "" && tmpl != "" {
		return OutputSettings{}, fmt.Errorf("cannot use --jq and --template simultaneously")
	}
	if jqExpr != "" && !jsonEnabled {
		return OutputSettings{}, fmt.Errorf("--jq requires --json")
	}

	f := ""
	if jsonEnabled {
		f = "json"
	} else if yamlEnabled {
		f = "yaml"
	}

	return OutputSettings{Format: f, JQ: jqExpr, Template: tmpl}, nil
}

func WriteOutput(cmd *cobra.Command, w io.Writer, data any, fallback func() error) error {
	settings, err := ResolveOutputSettings(cmd)
	if err != nil {
		return err
	}
	opts := format.Options{Format: settings.Format, JQ: settings.JQ, Template: settings.Template}
	return format.Write(w, opts, data, fallback)
}
