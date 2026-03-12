package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/magento"
)

func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Query Magento system configuration values",
		Long: `Query and inspect Magento system configuration values.

Useful for auditing settings, comparing environments (staging vs prod),
and verifying that configuration values are set correctly.

Examples:
  magecli config list
  magecli config list --filter web/secure
  magecli config get general/locale/code
  magecli config dump --json`,
	}
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newGetCmd(f))
	cmd.AddCommand(newDumpCmd(f))
	return cmd
}

// --- list subcommand ---

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	var filter string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configuration values as path=value pairs",
		Long: `List all queryable Magento configuration values in a flat path=value format.

Use --filter to narrow results by config path prefix or keyword.

This output is designed for easy comparison between environments:
  diff <(magecli -c staging config list) <(magecli -c prod config list)

Examples:
  magecli config list
  magecli config list --filter web/secure
  magecli config list --filter currency
  magecli config list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, f, filter)
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Filter paths by prefix or keyword")
	return cmd
}

func runList(cmd *cobra.Command, f *cmdutil.Factory, filter string) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	_, ctx, host, err := cmdutil.ResolveContext(f, cmd, cmdutil.FlagValue(cmd, "context"))
	if err != nil {
		return err
	}

	client, err := cmdutil.NewMagentoClient(host, ctx.StoreCode)
	if err != nil {
		return err
	}

	entries, err := client.GetConfigEntries(cmd.Context(), "")
	if err != nil {
		return err
	}

	entries = magento.FilterConfigEntries(entries, filter)

	return cmdutil.WriteOutput(cmd, ios.Out, entries, func() error {
		if len(entries) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No configuration values found.")
			return nil
		}

		// Group by scope for readable output
		currentScope := ""
		for _, e := range entries {
			if e.Scope != currentScope {
				if currentScope != "" {
					_, _ = fmt.Fprintln(ios.Out)
				}
				_, _ = fmt.Fprintf(ios.Out, "[%s]\n", e.Scope)
				currentScope = e.Scope
			}
			_, _ = fmt.Fprintf(ios.Out, "  %-45s = %s\n", e.Path, e.Value)
		}
		return nil
	})
}

// --- get subcommand ---

func newGetCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <path>",
		Short: "Get a specific configuration value by path",
		Long: `Get the value of a specific Magento configuration path.

Supports exact path matches and prefix matches. When multiple stores are
configured, shows the value for each store scope.

Examples:
  magecli config get general/locale/code
  magecli config get web/secure/base_url
  magecli config get currency`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
}

func runGet(cmd *cobra.Command, f *cmdutil.Factory, path string) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	_, ctx, host, err := cmdutil.ResolveContext(f, cmd, cmdutil.FlagValue(cmd, "context"))
	if err != nil {
		return err
	}

	client, err := cmdutil.NewMagentoClient(host, ctx.StoreCode)
	if err != nil {
		return err
	}

	entries, err := client.GetConfigEntries(cmd.Context(), "")
	if err != nil {
		return err
	}

	// Try exact match first, then prefix/contains match
	var matched []magento.ConfigEntry
	for _, e := range entries {
		if strings.EqualFold(e.Path, path) {
			matched = append(matched, e)
		}
	}
	if len(matched) == 0 {
		matched = magento.FilterConfigEntries(entries, path)
	}

	if len(matched) == 0 {
		return fmt.Errorf("no configuration found matching %q\n\nUse 'magecli config list' to see all available paths", path)
	}

	return cmdutil.WriteOutput(cmd, ios.Out, matched, func() error {
		for _, e := range matched {
			_, _ = fmt.Fprintf(ios.Out, "[%s] %s = %s\n", e.Scope, e.Path, e.Value)
		}
		return nil
	})
}

// --- dump subcommand ---

func newDumpCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "dump",
		Short: "Dump all configuration as structured data",
		Long: `Dump all queryable Magento configuration in structured format.

Best used with --json or --yaml for environment comparison:
  diff <(magecli -c staging config dump --json) <(magecli -c prod config dump --json)
  magecli config dump --json --jq '.[].locale'

Examples:
  magecli config dump --json
  magecli config dump --yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDump(cmd, f)
		},
	}
}

func runDump(cmd *cobra.Command, f *cmdutil.Factory) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	_, ctx, host, err := cmdutil.ResolveContext(f, cmd, cmdutil.FlagValue(cmd, "context"))
	if err != nil {
		return err
	}

	client, err := cmdutil.NewMagentoClient(host, ctx.StoreCode)
	if err != nil {
		return err
	}

	raw, err := client.GetStoreConfigRaw(cmd.Context())
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, raw, func() error {
		// For plain text, delegate to list format
		entries, err := client.GetConfigEntries(cmd.Context(), "")
		if err != nil {
			return err
		}
		currentScope := ""
		for _, e := range entries {
			if e.Scope != currentScope {
				if currentScope != "" {
					_, _ = fmt.Fprintln(ios.Out)
				}
				_, _ = fmt.Fprintf(ios.Out, "[%s]\n", e.Scope)
				currentScope = e.Scope
			}
			_, _ = fmt.Fprintf(ios.Out, "  %-45s = %s\n", e.Path, e.Value)
		}
		return nil
	})
}
