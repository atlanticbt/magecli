package inventory

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmdutil"
)

func NewCmdInventory(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "Check product inventory and stock status",
	}
	cmd.AddCommand(newStatusCmd(f))
	return cmd
}

func newStatusCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <sku>",
		Short: "Check stock status for a product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, f, args[0])
		},
	}
	cmd.Flags().String("store-code", "", "Override store code")
	return cmd
}

func runStatus(cmd *cobra.Command, f *cmdutil.Factory, sku string) error {
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

	stock, err := client.GetStockStatus(cmd.Context(), sku)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, stock, func() error {
		_, _ = fmt.Fprintf(ios.Out, "SKU:         %s\n", sku)
		inStock := "Yes"
		if !stock.IsInStock {
			inStock = "No"
		}
		_, _ = fmt.Fprintf(ios.Out, "In Stock:    %s\n", inStock)
		_, _ = fmt.Fprintf(ios.Out, "Quantity:    %.0f\n", stock.Qty)
		_, _ = fmt.Fprintf(ios.Out, "Min Qty:     %.0f\n", stock.MinQty)
		_, _ = fmt.Fprintf(ios.Out, "Min Sale:    %.0f\n", stock.MinSaleQty)
		_, _ = fmt.Fprintf(ios.Out, "Max Sale:    %.0f\n", stock.MaxSaleQty)
		return nil
	})
}
