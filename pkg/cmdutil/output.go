package cmdutil

import (
	"fmt"
	"io"
	"strings"

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

// StructuredOutputRequested reports whether the user selected a structured
// output mode (--json, --yaml, or --template) for this command.
func StructuredOutputRequested(cmd *cobra.Command) bool {
	settings, err := ResolveOutputSettings(cmd)
	if err != nil {
		return false
	}
	return settings.Format != "" || settings.Template != ""
}

// ValidateFields rejects --fields with table output: the server omits the
// unrequested fields, so the table would render them as fabricated zero
// values (price 0.00, status "disabled").
func ValidateFields(cmd *cobra.Command, fields string) error {
	if fields != "" && !StructuredOutputRequested(cmd) {
		return fmt.Errorf("--fields requires --json, --yaml, or --template output")
	}
	return nil
}

// ValidateListFields additionally rejects a pre-wrapped items[...] projection
// on list commands. The search builder wraps the field list in
// items[...],total_count itself; a double-wrapped items[items[...]] makes
// Magento silently return empty objects instead of an error (found live
// against Magento 2.4.7).
func ValidateListFields(cmd *cobra.Command, fields string) error {
	if err := ValidateFields(cmd, fields); err != nil {
		return err
	}
	if strings.HasPrefix(strings.TrimSpace(fields), "items[") {
		return fmt.Errorf(`--fields is wrapped in items[...] automatically; pass the bare field list (e.g. "sku,name,price")`)
	}
	return nil
}

func WriteOutput(cmd *cobra.Command, w io.Writer, data any, fallback func() error) error {
	settings, err := ResolveOutputSettings(cmd)
	if err != nil {
		return err
	}
	opts := format.Options{Format: settings.Format, JQ: settings.JQ, Template: settings.Template}
	return format.Write(w, opts, data, fallback)
}
