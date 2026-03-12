package category

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/magento"
)

func NewCmdCategory(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "category",
		Short: "Browse Magento categories",
	}
	cmd.AddCommand(newTreeCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newProductsCmd(f))
	return cmd
}

func newTreeCmd(f *cmdutil.Factory) *cobra.Command {
	var rootID int
	var depth int

	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Display category tree",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTree(cmd, f, rootID, depth)
		},
	}
	cmd.Flags().IntVar(&rootID, "root", 0, "Root category ID")
	cmd.Flags().IntVar(&depth, "depth", 0, "Maximum depth to display")
	cmd.Flags().String("store-code", "", "Override store code")
	return cmd
}

func runTree(cmd *cobra.Command, f *cmdutil.Factory, rootID, depth int) error {
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

	tree, err := client.GetCategoryTree(cmd.Context(), rootID, depth)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, tree, func() error {
		printTree(ios.Out, tree, 0)
		return nil
	})
}

func printTree(w interface{ Write([]byte) (int, error) }, cat *magento.Category, indent int) {
	if cat == nil {
		return
	}
	prefix := strings.Repeat("  ", indent)
	active := ""
	if !cat.IsActive {
		active = " [inactive]"
	}
	_, _ = fmt.Fprintf(w, "%s[%d] %s (products: %d)%s\n", prefix, cat.ID, cat.Name, cat.ProductCount, active)
	for i := range cat.ChildrenData {
		printTree(w, &cat.ChildrenData[i], indent+1)
	}
}

func newViewCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a category by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid category ID: %w", err)
			}
			return runView(cmd, f, id)
		},
	}
	cmd.Flags().String("store-code", "", "Override store code")
	return cmd
}

func runView(cmd *cobra.Command, f *cmdutil.Factory, id int) error {
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

	cat, err := client.GetCategory(cmd.Context(), id)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, cat, func() error {
		_, _ = fmt.Fprintf(ios.Out, "ID:        %d\n", cat.ID)
		_, _ = fmt.Fprintf(ios.Out, "Name:      %s\n", cat.Name)
		_, _ = fmt.Fprintf(ios.Out, "Parent ID: %d\n", cat.ParentID)
		_, _ = fmt.Fprintf(ios.Out, "Level:     %d\n", cat.Level)
		_, _ = fmt.Fprintf(ios.Out, "Position:  %d\n", cat.Position)
		_, _ = fmt.Fprintf(ios.Out, "Active:    %v\n", cat.IsActive)
		_, _ = fmt.Fprintf(ios.Out, "Products:  %d\n", cat.ProductCount)
		return nil
	})
}

func newProductsCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "products <id>",
		Short: "List products in a category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid category ID: %w", err)
			}
			return runProducts(cmd, f, id)
		},
	}
	cmd.Flags().String("store-code", "", "Override store code")
	return cmd
}

func runProducts(cmd *cobra.Command, f *cmdutil.Factory, id int) error {
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

	products, err := client.GetCategoryProducts(cmd.Context(), id)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, products, func() error {
		if len(products) == 0 {
			_, _ = fmt.Fprintf(ios.Out, "No products in category %d\n", id)
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Products in category %d:\n\n", id)
		for _, p := range products {
			_, _ = fmt.Fprintf(ios.Out, "  %-25s  position: %d\n", p.SKU, p.Position)
		}
		return nil
	})
}
