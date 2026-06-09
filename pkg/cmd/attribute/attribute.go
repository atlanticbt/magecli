package attribute

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/magento"
)

func NewCmdAttribute(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attribute",
		Short: "Manage product attributes",
	}
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newOptionsCmd(f))
	cmd.AddCommand(newSetsCmd(f))
	return cmd
}

func newViewCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <code>",
		Short: "View an attribute definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(cmd, f, args[0])
		},
	}
	return cmd
}

func runView(cmd *cobra.Command, f *cmdutil.Factory, code string) error {
	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
	if err != nil {
		return err
	}

	attr, err := client.GetAttribute(cmd.Context(), code)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, attr, func() error {
		_, _ = fmt.Fprintf(ios.Out, "Code:          %s\n", attr.AttributeCode)
		_, _ = fmt.Fprintf(ios.Out, "Label:         %s\n", attr.FrontendLabel)
		_, _ = fmt.Fprintf(ios.Out, "Input:         %s\n", attr.FrontendInput)
		_, _ = fmt.Fprintf(ios.Out, "Required:      %v\n", attr.IsRequired)
		_, _ = fmt.Fprintf(ios.Out, "User-defined:  %v\n", attr.IsUserDefined)
		if len(attr.Options) > 0 {
			_, _ = fmt.Fprintln(ios.Out, "\nOptions:")
			for _, o := range attr.Options {
				if o.Value == "" {
					continue
				}
				_, _ = fmt.Fprintf(ios.Out, "  [%s] %s\n", o.Value, o.Label)
			}
		}
		return nil
	})
}

func newOptionsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "options <code>",
		Short: "List option values for an attribute",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOptions(cmd, f, args[0])
		},
	}
	return cmd
}

func runOptions(cmd *cobra.Command, f *cmdutil.Factory, code string) error {
	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
	if err != nil {
		return err
	}

	options, err := client.GetAttributeOptions(cmd.Context(), code)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, options, func() error {
		if len(options) == 0 {
			_, _ = fmt.Fprintf(ios.Out, "No options for attribute %q\n", code)
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Options for %q:\n\n", code)
		for _, o := range options {
			if o.Value == "" {
				continue
			}
			_, _ = fmt.Fprintf(ios.Out, "  [%s] %s\n", o.Value, o.Label)
		}
		return nil
	})
}

func newSetsCmd(f *cmdutil.Factory) *cobra.Command {
	var filters []string
	var limit, page int

	cmd := &cobra.Command{
		Use:   "sets",
		Short: "List attribute sets",
		Long: `List attribute sets with optional filtering and pagination.

Examples:
  magecli attribute sets
  magecli attribute sets --filter "attribute_set_name like %Default%" --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSets(cmd, f, filters, limit, page)
		},
	}
	cmd.Flags().StringArrayVar(&filters, "filter", nil, `Filter expression (e.g. "attribute_set_name like %Default%")`)
	cmd.Flags().IntVar(&limit, "limit", 100, "Number of results per page (1-10000)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	return cmd
}

func runSets(cmd *cobra.Command, f *cmdutil.Factory, filters []string, limit, page int) error {
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
	for _, expr := range filters {
		if err := search.AddFilter(expr); err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}
	}

	result, err := client.ListAttributeSets(cmd.Context(), search)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
		if len(result.Items) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No attribute sets found.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Attribute Sets (%d):\n\n", result.TotalCount)
		for _, s := range result.Items {
			_, _ = fmt.Fprintf(ios.Out, "  [%d] %s\n", s.AttributeSetID, s.AttributeSetName)
		}
		return nil
	})
}
