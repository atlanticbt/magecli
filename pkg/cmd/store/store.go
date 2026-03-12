package store

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/magento"
)

func NewCmdStore(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store",
		Short: "View store configuration, views, and websites",
	}
	cmd.AddCommand(newViewsCmd(f))
	cmd.AddCommand(newConfigCmd(f))
	cmd.AddCommand(newGroupsCmd(f))
	cmd.AddCommand(newWebsitesCmd(f))
	return cmd
}

func newViewsCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "views",
		Short: "List all store views",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runViews(cmd, f)
		},
	}
}

func runViews(cmd *cobra.Command, f *cmdutil.Factory) error {
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

	views, err := client.ListStoreViews(cmd.Context())
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, views, func() error {
		if len(views) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No store views found.")
			return nil
		}

		// Table header
		_, _ = fmt.Fprintf(ios.Out, "%-4s  %-15s  %-30s  %-10s  %-8s  %s\n",
			"ID", "CODE", "NAME", "WEBSITE", "GROUP", "ACTIVE")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 85))

		for _, v := range views {
			active := "yes"
			if v.IsActive != 1 {
				active = "no"
			}
			_, _ = fmt.Fprintf(ios.Out, "%-4d  %-15s  %-30s  %-10d  %-8d  %s\n",
				v.ID, v.Code, truncate(v.Name, 30), v.WebsiteID, v.StoreGroupID, active)
		}
		return nil
	})
}

func newConfigCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "config [store-code]",
		Short: "Show store configuration (locale, currency, URLs)",
		Long: `Display store configuration including locale, currency, timezone, and base URLs.

Without arguments, shows all store configs. With a store code argument, filters to that store.

Examples:
  magecli store config
  magecli store config default
  magecli store config --json --jq '.[0].base_url'`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := ""
			if len(args) > 0 {
				filter = args[0]
			}
			return runConfig(cmd, f, filter)
		},
	}
}

func runConfig(cmd *cobra.Command, f *cmdutil.Factory, filterCode string) error {
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

	configs, err := client.ListStoreConfigs(cmd.Context())
	if err != nil {
		return err
	}

	// Filter to specific store code if provided
	if filterCode != "" {
		var filtered []magento.StoreConfig
		for _, c := range configs {
			if c.Code == filterCode {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("store config for code %q not found", filterCode)
		}
		configs = filtered
	}

	return cmdutil.WriteOutput(cmd, ios.Out, configs, func() error {
		for i := range configs {
			if i > 0 {
				_, _ = fmt.Fprintln(ios.Out)
			}
			c := &configs[i]
			_, _ = fmt.Fprintf(ios.Out, "Store:    %s (ID: %d)\n", c.Code, c.ID)
			_, _ = fmt.Fprintf(ios.Out, "Locale:   %s\n", c.Locale)
			_, _ = fmt.Fprintf(ios.Out, "Currency: %s (display: %s)\n", c.BaseCurrencyCode, c.DefaultDisplayCurrency)
			_, _ = fmt.Fprintf(ios.Out, "Timezone: %s\n", c.Timezone)
			_, _ = fmt.Fprintf(ios.Out, "Weight:   %s\n", c.WeightUnit)
			if c.BaseURL != "" {
				_, _ = fmt.Fprintf(ios.Out, "Base URL: %s\n", c.BaseURL)
			}
			if c.SecureBaseURL != "" {
				_, _ = fmt.Fprintf(ios.Out, "SSL URL:  %s\n", c.SecureBaseURL)
			}
			if c.BaseMediaURL != "" {
				_, _ = fmt.Fprintf(ios.Out, "Media:    %s\n", c.BaseMediaURL)
			}
		}
		return nil
	})
}

func newGroupsCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "groups",
		Short: "List all store groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGroups(cmd, f)
		},
	}
}

func runGroups(cmd *cobra.Command, f *cmdutil.Factory) error {
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

	groups, err := client.ListStoreGroups(cmd.Context())
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, groups, func() error {
		if len(groups) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No store groups found.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "%-4s  %-15s  %-30s  %-10s  %-12s\n",
			"ID", "CODE", "NAME", "WEBSITE", "ROOT CAT ID")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 75))
		for _, g := range groups {
			_, _ = fmt.Fprintf(ios.Out, "%-4d  %-15s  %-30s  %-10d  %-12d\n",
				g.ID, g.Code, truncate(g.Name, 30), g.WebsiteID, g.RootCategoryID)
		}
		return nil
	})
}

func newWebsitesCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "websites",
		Short: "List all websites",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWebsites(cmd, f)
		},
	}
}

func runWebsites(cmd *cobra.Command, f *cmdutil.Factory) error {
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

	websites, err := client.ListWebsites(cmd.Context())
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, websites, func() error {
		if len(websites) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No websites found.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "%-4s  %-15s  %-30s  %s\n",
			"ID", "CODE", "NAME", "DEFAULT GROUP")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 65))
		for _, w := range websites {
			_, _ = fmt.Fprintf(ios.Out, "%-4d  %-15s  %-30s  %d\n",
				w.ID, w.Code, truncate(w.Name, 30), w.DefaultGroupID)
		}
		return nil
	})
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
