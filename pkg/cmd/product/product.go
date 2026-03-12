package product

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/magento"
)

func NewCmdProduct(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "product",
		Short: "Manage Magento catalog products",
	}
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newSearchCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newMediaCmd(f))
	cmd.AddCommand(newChildrenCmd(f))
	cmd.AddCommand(newOptionsCmd(f))
	cmd.AddCommand(newURLCmd(f))
	return cmd
}

type listOptions struct {
	Filters   []string
	Sort      []string
	Limit     int
	Page      int
	StoreCode string
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List products matching filters",
		Long: `Search and list Magento products with filtering, sorting, and pagination.

Examples:
  magecli product list --filter "name like %shirt%"
  magecli product list --filter "price gt 50" --sort "price:ASC" --limit 10
  magecli product list --filter "status eq 1" --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, f, opts)
		},
	}
	cmd.Flags().StringArrayVar(&opts.Filters, "filter", nil, `Filter expression (e.g. "name like %shirt%")`)
	cmd.Flags().StringArrayVar(&opts.Sort, "sort", nil, `Sort expression (e.g. "price:ASC")`)
	cmd.Flags().IntVar(&opts.Limit, "limit", 20, "Number of results per page")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number")
	cmd.Flags().StringVar(&opts.StoreCode, "store-code", "", "Override store code")
	return cmd
}

func runList(cmd *cobra.Command, f *cmdutil.Factory, opts *listOptions) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	_, ctx, host, err := cmdutil.ResolveContext(f, cmd, cmdutil.FlagValue(cmd, "context"))
	if err != nil {
		return err
	}

	storeCode := cmdutil.FlagValue(cmd, "store-code")
	if storeCode == "" {
		storeCode = ctx.StoreCode
	}

	client, err := cmdutil.NewMagentoClient(host, storeCode)
	if err != nil {
		return err
	}

	search := magento.NewSearch()
	search.SetPageSize(opts.Limit)
	search.SetCurrentPage(opts.Page)

	for _, expr := range opts.Filters {
		if err := search.AddFilter(expr); err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}
	}
	for _, expr := range opts.Sort {
		if err := search.AddSort(expr); err != nil {
			return fmt.Errorf("invalid sort: %w", err)
		}
	}

	result, err := client.ListProducts(cmd.Context(), search)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
		if len(result.Items) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No products found.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Products (%d total, page %d of %d):\n\n",
			result.TotalCount, opts.Page, (result.TotalCount+opts.Limit-1)/opts.Limit)

		_, _ = fmt.Fprintf(ios.Out, "%-20s  %-40s  %10s  %-14s  %s\n",
			"SKU", "NAME", "PRICE", "TYPE", "STATUS")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 95))

		for _, p := range result.Items {
			status := "enabled"
			if p.Status != 1 {
				status = "disabled"
			}
			_, _ = fmt.Fprintf(ios.Out, "%-20s  %-40s  %10.2f  %-14s  %s\n",
				truncate(p.SKU, 20), truncate(p.Name, 40), p.Price, p.TypeID, status)
		}
		return nil
	})
}

func newViewCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <sku>",
		Short: "View a product by SKU",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(cmd, f, args[0])
		},
	}
	cmd.Flags().String("store-code", "", "Override store code")
	return cmd
}

func runView(cmd *cobra.Command, f *cmdutil.Factory, sku string) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	_, ctx, host, err := cmdutil.ResolveContext(f, cmd, cmdutil.FlagValue(cmd, "context"))
	if err != nil {
		return err
	}

	storeCode := cmdutil.FlagValue(cmd, "store-code")
	if storeCode == "" {
		storeCode = ctx.StoreCode
	}

	client, err := cmdutil.NewMagentoClient(host, storeCode)
	if err != nil {
		return err
	}

	product, err := client.GetProduct(cmd.Context(), sku)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, product, func() error {
		_, _ = fmt.Fprintf(ios.Out, "Name:       %s\n", product.Name)
		_, _ = fmt.Fprintf(ios.Out, "SKU:        %s\n", product.SKU)
		_, _ = fmt.Fprintf(ios.Out, "Type:       %s\n", product.TypeID)
		_, _ = fmt.Fprintf(ios.Out, "Price:      $%.2f\n", product.Price)
		status := "Enabled"
		if product.Status != 1 {
			status = "Disabled"
		}
		_, _ = fmt.Fprintf(ios.Out, "Status:     %s\n", status)
		_, _ = fmt.Fprintf(ios.Out, "Visibility: %d\n", product.Visibility)
		if product.Weight > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Weight:     %.2f\n", product.Weight)
		}
		_, _ = fmt.Fprintf(ios.Out, "Created:    %s\n", product.CreatedAt)
		_, _ = fmt.Fprintf(ios.Out, "Updated:    %s\n", product.UpdatedAt)

		if len(product.CustomAttributes) > 0 {
			_, _ = fmt.Fprintln(ios.Out, "\nCustom Attributes:")
			for _, attr := range product.CustomAttributes {
				_, _ = fmt.Fprintf(ios.Out, "  %s: %v\n", attr.AttributeCode, attr.Value)
			}
		}
		return nil
	})
}

func newMediaCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media <sku>",
		Short: "List media gallery entries for a product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMedia(cmd, f, args[0])
		},
	}
	cmd.Flags().String("store-code", "", "Override store code")
	return cmd
}

func runMedia(cmd *cobra.Command, f *cmdutil.Factory, sku string) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	_, ctx, host, err := cmdutil.ResolveContext(f, cmd, cmdutil.FlagValue(cmd, "context"))
	if err != nil {
		return err
	}

	storeCode := cmdutil.FlagValue(cmd, "store-code")
	if storeCode == "" {
		storeCode = ctx.StoreCode
	}

	client, err := cmdutil.NewMagentoClient(host, storeCode)
	if err != nil {
		return err
	}

	media, err := client.GetProductMedia(cmd.Context(), sku)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, media, func() error {
		if len(media) == 0 {
			_, _ = fmt.Fprintf(ios.Out, "No media entries for %s\n", sku)
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Media for %s:\n\n", sku)
		for _, m := range media {
			_, _ = fmt.Fprintf(ios.Out, "  [%d] %s  %s  types=%s\n", m.Position, m.File, m.MediaType, strings.Join(m.Types, ","))
			if m.Label != "" {
				_, _ = fmt.Fprintf(ios.Out, "       label: %s\n", m.Label)
			}
		}
		return nil
	})
}

func newChildrenCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "children <sku>",
		Short: "List configurable product variants",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChildren(cmd, f, args[0])
		},
	}
	cmd.Flags().String("store-code", "", "Override store code")
	return cmd
}

func runChildren(cmd *cobra.Command, f *cmdutil.Factory, sku string) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	_, ctx, host, err := cmdutil.ResolveContext(f, cmd, cmdutil.FlagValue(cmd, "context"))
	if err != nil {
		return err
	}

	storeCode := cmdutil.FlagValue(cmd, "store-code")
	if storeCode == "" {
		storeCode = ctx.StoreCode
	}

	client, err := cmdutil.NewMagentoClient(host, storeCode)
	if err != nil {
		return err
	}

	children, err := client.GetConfigurableChildren(cmd.Context(), sku)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, children, func() error {
		if len(children) == 0 {
			_, _ = fmt.Fprintf(ios.Out, "No children for %s (may not be configurable)\n", sku)
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Children of %s:\n\n", sku)
		for _, c := range children {
			_, _ = fmt.Fprintf(ios.Out, "  %-25s  $%.2f  %s\n", c.SKU, c.Price, c.Name)
		}
		return nil
	})
}

func newOptionsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "options <sku>",
		Short: "List configurable product options",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOptions(cmd, f, args[0])
		},
	}
	cmd.Flags().String("store-code", "", "Override store code")
	return cmd
}

func runOptions(cmd *cobra.Command, f *cmdutil.Factory, sku string) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	_, ctx, host, err := cmdutil.ResolveContext(f, cmd, cmdutil.FlagValue(cmd, "context"))
	if err != nil {
		return err
	}

	storeCode := cmdutil.FlagValue(cmd, "store-code")
	if storeCode == "" {
		storeCode = ctx.StoreCode
	}

	client, err := cmdutil.NewMagentoClient(host, storeCode)
	if err != nil {
		return err
	}

	options, err := client.GetConfigurableOptions(cmd.Context(), sku)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, options, func() error {
		if len(options) == 0 {
			_, _ = fmt.Fprintf(ios.Out, "No configurable options for %s\n", sku)
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Configurable options for %s:\n\n", sku)
		for _, o := range options {
			_, _ = fmt.Fprintf(ios.Out, "  %s (attribute_id: %s, position: %d)\n", o.Label, o.AttributeID, o.Position)
			for _, v := range o.Values {
				_, _ = fmt.Fprintf(ios.Out, "    value_index: %d\n", v.ValueIndex)
			}
		}
		return nil
	})
}

func newSearchCmd(f *cmdutil.Factory) *cobra.Command {
	var limit int
	var sort []string

	cmd := &cobra.Command{
		Use:   "search <term>",
		Short: "Quick product name search",
		Long: `Search products by name. Shortcut for: product list --filter "name like %term%"

Examples:
  magecli product search glock
  magecli product search "body armor" --limit 10
  magecli product search holster --json --jq '.items[].sku'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(cmd, f, args[0], limit, sort)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Number of results")
	cmd.Flags().StringArrayVar(&sort, "sort", nil, `Sort expression (e.g. "price:ASC")`)
	cmd.Flags().String("store-code", "", "Override store code")
	return cmd
}

func runSearch(cmd *cobra.Command, f *cmdutil.Factory, term string, limit int, sortExprs []string) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	_, ctx, host, err := cmdutil.ResolveContext(f, cmd, cmdutil.FlagValue(cmd, "context"))
	if err != nil {
		return err
	}

	storeCode := cmdutil.FlagValue(cmd, "store-code")
	if storeCode == "" {
		storeCode = ctx.StoreCode
	}

	client, err := cmdutil.NewMagentoClient(host, storeCode)
	if err != nil {
		return err
	}

	search := magento.NewSearch()
	search.SetPageSize(limit)
	if err := search.AddFilter(fmt.Sprintf("name like %%%s%%", term)); err != nil {
		return err
	}
	for _, expr := range sortExprs {
		if err := search.AddSort(expr); err != nil {
			return fmt.Errorf("invalid sort: %w", err)
		}
	}

	result, err := client.ListProducts(cmd.Context(), search)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
		if len(result.Items) == 0 {
			_, _ = fmt.Fprintf(ios.Out, "No products matching %q\n", term)
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Search %q (%d results):\n\n", term, result.TotalCount)

		_, _ = fmt.Fprintf(ios.Out, "%-20s  %-40s  %10s  %-14s  %s\n",
			"SKU", "NAME", "PRICE", "TYPE", "STATUS")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 95))

		for _, p := range result.Items {
			status := "enabled"
			if p.Status != 1 {
				status = "disabled"
			}
			_, _ = fmt.Fprintf(ios.Out, "%-20s  %-40s  %10.2f  %-14s  %s\n",
				truncate(p.SKU, 20), truncate(p.Name, 40), p.Price, p.TypeID, status)
		}
		return nil
	})
}

func newURLCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "url <url-key>",
		Short: "Find a product by URL key",
		Long: `Look up which product has a given URL key.

Examples:
  magecli product url glock-17-gen5
  magecli product url "hatch-g3-giant-swat-bag" --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runURL(cmd, f, args[0])
		},
	}
	cmd.Flags().String("store-code", "", "Override store code")
	return cmd
}

func runURL(cmd *cobra.Command, f *cmdutil.Factory, urlKey string) error {
	ios, err := f.Streams()
	if err != nil {
		return err
	}

	_, ctx, host, err := cmdutil.ResolveContext(f, cmd, cmdutil.FlagValue(cmd, "context"))
	if err != nil {
		return err
	}

	storeCode := cmdutil.FlagValue(cmd, "store-code")
	if storeCode == "" {
		storeCode = ctx.StoreCode
	}

	client, err := cmdutil.NewMagentoClient(host, storeCode)
	if err != nil {
		return err
	}

	search := magento.NewSearch()
	search.SetPageSize(5)
	if err := search.AddFilter(fmt.Sprintf("url_key eq %s", urlKey)); err != nil {
		return err
	}

	result, err := client.ListProducts(cmd.Context(), search)
	if err != nil {
		return err
	}

	if len(result.Items) == 0 {
		return fmt.Errorf("no product found with url_key %q", urlKey)
	}

	// Return first match (url_key should be unique)
	product := &result.Items[0]

	return cmdutil.WriteOutput(cmd, ios.Out, product, func() error {
		_, _ = fmt.Fprintf(ios.Out, "Name:     %s\n", product.Name)
		_, _ = fmt.Fprintf(ios.Out, "SKU:      %s\n", product.SKU)
		_, _ = fmt.Fprintf(ios.Out, "Type:     %s\n", product.TypeID)
		_, _ = fmt.Fprintf(ios.Out, "Price:    $%.2f\n", product.Price)
		status := "Enabled"
		if product.Status != 1 {
			status = "Disabled"
		}
		_, _ = fmt.Fprintf(ios.Out, "Status:   %s\n", status)
		_, _ = fmt.Fprintf(ios.Out, "URL Key:  %s\n", urlKey)
		return nil
	})
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
