package cms

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/magento"
)

func NewCmdCMS(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cms",
		Short: "Browse CMS pages and blocks",
	}
	cmd.AddCommand(newPageCmd(f))
	cmd.AddCommand(newBlockCmd(f))
	return cmd
}

// --- Pages ---

func newPageCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "page",
		Short: "Manage CMS pages",
	}
	cmd.AddCommand(newPageListCmd(f))
	cmd.AddCommand(newPageViewCmd(f))
	return cmd
}

func newPageListCmd(f *cmdutil.Factory) *cobra.Command {
	var filters []string
	var limit, page int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List CMS pages",
		Long: `List CMS pages with optional filtering.

Page HTML content is omitted from list output; use 'cms page view <id> --content'
to retrieve a single page's body.

Examples:
  magecli cms page list
  magecli cms page list --filter "identifier like %home%"
  magecli cms page list --filter "content like %promo%"   # search page bodies
  magecli cms page list --json --jq '.items[] | {id, identifier, title}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPageList(cmd, f, filters, limit, page)
		},
	}
	cmd.Flags().StringArrayVar(&filters, "filter", nil, `Filter expression (e.g. "title like %about%")`)
	cmd.Flags().IntVar(&limit, "limit", 50, "Number of results per page (1-10000)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	return cmd
}

func runPageList(cmd *cobra.Command, f *cmdutil.Factory, filters []string, limit, page int) error {
	if err := cmdutil.ValidateLimit(limit); err != nil {
		return err
	}

	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
	if err != nil {
		return err
	}

	search := magento.NewSearch()
	search.SetPageSize(limit)
	search.SetCurrentPage(page)
	// Exclude bulky HTML bodies server-side (retrieve via `cms page view`);
	// content filters still work, fields= only shapes the response.
	search.SetFields(magento.CMSPageMetadataFields)
	for _, expr := range filters {
		if err := search.AddFilter(expr); err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}
	}

	result, err := client.ListCMSPages(cmd.Context(), search)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
		if len(result.Items) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No CMS pages found.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "CMS Pages (%d total):\n\n", result.TotalCount)
		_, _ = fmt.Fprintf(ios.Out, "%-5s  %-25s  %-40s  %s\n", "ID", "IDENTIFIER", "TITLE", "ACTIVE")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 80))
		for _, p := range result.Items {
			active := "yes"
			if !p.Active {
				active = "no"
			}
			_, _ = fmt.Fprintf(ios.Out, "%-5d  %-25s  %-40s  %s\n",
				p.ID, cmdutil.Truncate(p.Identifier, 25), cmdutil.Truncate(p.Title, 40), active)
		}
		return nil
	})
}

func newPageViewCmd(f *cmdutil.Factory) *cobra.Command {
	var showContent bool

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a CMS page by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid page ID: %w", err)
			}
			return runPageView(cmd, f, id, showContent)
		},
	}
	cmd.Flags().BoolVar(&showContent, "content", false, "Include page HTML content in output")
	return cmd
}

func runPageView(cmd *cobra.Command, f *cmdutil.Factory, id int, showContent bool) error {
	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
	if err != nil {
		return err
	}

	// Omit the HTML body unless explicitly requested — server-side, so it is
	// never transferred, and absent from JSON/YAML output too.
	fields := magento.CMSPageMetadataFields
	if showContent {
		fields = ""
	}
	page, err := client.GetCMSPage(cmd.Context(), id, fields)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, page, func() error {
		_, _ = fmt.Fprintf(ios.Out, "ID:         %d\n", page.ID)
		_, _ = fmt.Fprintf(ios.Out, "Identifier: %s\n", page.Identifier)
		_, _ = fmt.Fprintf(ios.Out, "Title:      %s\n", page.Title)
		_, _ = fmt.Fprintf(ios.Out, "Active:     %v\n", page.Active)
		_, _ = fmt.Fprintf(ios.Out, "Created:    %s\n", page.CreationTime)
		_, _ = fmt.Fprintf(ios.Out, "Updated:    %s\n", page.UpdateTime)
		if page.MetaTitle != "" {
			_, _ = fmt.Fprintf(ios.Out, "Meta Title: %s\n", page.MetaTitle)
		}
		if page.MetaDescription != "" {
			_, _ = fmt.Fprintf(ios.Out, "Meta Desc:  %s\n", cmdutil.Truncate(page.MetaDescription, 80))
		}
		if showContent && page.Content != "" {
			_, _ = fmt.Fprintf(ios.Out, "\nContent:\n%s\n", page.Content)
		}
		return nil
	})
}

// --- Blocks ---

func newBlockCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "Manage CMS blocks",
	}
	cmd.AddCommand(newBlockListCmd(f))
	cmd.AddCommand(newBlockViewCmd(f))
	return cmd
}

func newBlockListCmd(f *cmdutil.Factory) *cobra.Command {
	var filters []string
	var limit, page int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List CMS blocks",
		Long: `List CMS blocks with optional filtering.

Block HTML content is omitted from list output; use 'cms block view <id> --content'
to retrieve a single block's body.

Examples:
  magecli cms block list
  magecli cms block list --filter "identifier like %footer%"
  magecli cms block list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlockList(cmd, f, filters, limit, page)
		},
	}
	cmd.Flags().StringArrayVar(&filters, "filter", nil, `Filter expression (e.g. "identifier like %footer%")`)
	cmd.Flags().IntVar(&limit, "limit", 50, "Number of results per page (1-10000)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	return cmd
}

func runBlockList(cmd *cobra.Command, f *cmdutil.Factory, filters []string, limit, page int) error {
	if err := cmdutil.ValidateLimit(limit); err != nil {
		return err
	}

	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
	if err != nil {
		return err
	}

	search := magento.NewSearch()
	search.SetPageSize(limit)
	search.SetCurrentPage(page)
	// Exclude bulky HTML bodies server-side (retrieve via `cms block view`);
	// content filters still work, fields= only shapes the response.
	search.SetFields(magento.CMSBlockMetadataFields)
	for _, expr := range filters {
		if err := search.AddFilter(expr); err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}
	}

	result, err := client.ListCMSBlocks(cmd.Context(), search)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
		if len(result.Items) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No CMS blocks found.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "CMS Blocks (%d total):\n\n", result.TotalCount)
		_, _ = fmt.Fprintf(ios.Out, "%-5s  %-30s  %-40s  %s\n", "ID", "IDENTIFIER", "TITLE", "ACTIVE")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 85))
		for _, b := range result.Items {
			active := "yes"
			if !b.Active {
				active = "no"
			}
			_, _ = fmt.Fprintf(ios.Out, "%-5d  %-30s  %-40s  %s\n",
				b.ID, cmdutil.Truncate(b.Identifier, 30), cmdutil.Truncate(b.Title, 40), active)
		}
		return nil
	})
}

func newBlockViewCmd(f *cmdutil.Factory) *cobra.Command {
	var showContent bool

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a CMS block by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid block ID: %w", err)
			}
			return runBlockView(cmd, f, id, showContent)
		},
	}
	cmd.Flags().BoolVar(&showContent, "content", false, "Include block HTML content in output")
	return cmd
}

func runBlockView(cmd *cobra.Command, f *cmdutil.Factory, id int, showContent bool) error {
	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
	if err != nil {
		return err
	}

	// Omit the HTML body unless explicitly requested — server-side, so it is
	// never transferred, and absent from JSON/YAML output too.
	fields := magento.CMSBlockMetadataFields
	if showContent {
		fields = ""
	}
	block, err := client.GetCMSBlock(cmd.Context(), id, fields)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, block, func() error {
		_, _ = fmt.Fprintf(ios.Out, "ID:         %d\n", block.ID)
		_, _ = fmt.Fprintf(ios.Out, "Identifier: %s\n", block.Identifier)
		_, _ = fmt.Fprintf(ios.Out, "Title:      %s\n", block.Title)
		_, _ = fmt.Fprintf(ios.Out, "Active:     %v\n", block.Active)
		_, _ = fmt.Fprintf(ios.Out, "Created:    %s\n", block.CreationTime)
		_, _ = fmt.Fprintf(ios.Out, "Updated:    %s\n", block.UpdateTime)
		if showContent && block.Content != "" {
			_, _ = fmt.Fprintf(ios.Out, "\nContent:\n%s\n", block.Content)
		}
		return nil
	})
}
