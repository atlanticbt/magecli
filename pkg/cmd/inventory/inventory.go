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
	return cmd
}

func runStatus(cmd *cobra.Command, f *cmdutil.Factory, sku string) error {
	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
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
