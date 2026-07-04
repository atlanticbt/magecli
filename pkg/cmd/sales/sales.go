package sales

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/magento"
)

func NewCmdSales(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sales",
		Short: "View orders, invoices, shipments, credit memos, and sales totals",
		Long: `Query Magento sales documents and revenue totals.

All sales commands require the Integration token to have Sales resource
access (Magento Admin > System > Integrations); without it Magento
returns 403.

Examples:
  magecli sales order list --filter "status eq processing"
  magecli sales order view 000000042
  magecli sales totals --from 2026-06-01
  magecli sales invoice list --filter "order_id eq 42" --json`,
	}
	cmd.AddCommand(newOrderCmd(f))
	cmd.AddCommand(newDocumentCmd(f, documentSpec{
		use:     "invoice",
		short:   "List invoices (billing documents for orders)",
		long:    "List invoices with optional filtering and sorting.\n\nFilter by order_id to find a specific order's invoices.",
		example: `  magecli sales invoice list\n  magecli sales invoice list --filter "order_id eq 42" --json`,
		fields:  magento.InvoiceListFields,
		list: func(ios listContext) (*magento.GenericResult, error) {
			return ios.client.ListInvoices(ios.ctx, ios.search)
		},
		empty: "No invoices found.",
		title: "Invoices",
	}))
	cmd.AddCommand(newDocumentCmd(f, documentSpec{
		use:     "shipment",
		short:   "List shipments including tracking numbers",
		long:    "List shipments with optional filtering and sorting.\n\nFilter by order_id to find a specific order's shipments. Tracking\nnumbers and carriers are included by default.",
		example: `  magecli sales shipment list\n  magecli sales shipment list --filter "order_id eq 42" --json`,
		fields:  magento.ShipmentListFields,
		list: func(ios listContext) (*magento.GenericResult, error) {
			return ios.client.ListShipments(ios.ctx, ios.search)
		},
		empty: "No shipments found.",
		title: "Shipments",
	}))
	cmd.AddCommand(newDocumentCmd(f, documentSpec{
		use:     "creditmemo",
		short:   "List credit memos (refund documents)",
		long:    "List credit memos with optional filtering and sorting.\n\nFilter by order_id to find a specific order's refunds.",
		example: `  magecli sales creditmemo list\n  magecli sales creditmemo list --filter "order_id eq 42" --json`,
		fields:  magento.CreditmemoListFields,
		list: func(ios listContext) (*magento.GenericResult, error) {
			return ios.client.ListCreditmemos(ios.ctx, ios.search)
		},
		empty: "No credit memos found.",
		title: "Credit Memos",
	}))
	cmd.AddCommand(newTotalsCmd(f))
	return cmd
}

// --- Orders ---

func newOrderCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order",
		Short: "List and view orders",
	}
	cmd.AddCommand(newOrderListCmd(f))
	cmd.AddCommand(newOrderViewCmd(f))
	return cmd
}

func newOrderListCmd(f *cmdutil.Factory) *cobra.Command {
	var filters, sorts []string
	var limit, page int
	var fields string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List orders",
		Long: `List orders with optional filtering and sorting.

By default only a curated set of order fields is returned (an untrimmed
order payload runs 20-60KB); override with --fields.

Examples:
  magecli sales order list
  magecli sales order list --filter "status eq processing" --sort "created_at:DESC"
  magecli sales order list --filter "created_at from 2026-06-01" --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOrderList(cmd, f, filters, sorts, limit, page, fields)
		},
	}
	cmd.Flags().StringArrayVar(&filters, "filter", nil, `Filter expression (e.g. "status eq processing")`)
	cmd.Flags().StringArrayVar(&sorts, "sort", nil, `Sort expression (e.g. "created_at:DESC")`)
	cmd.Flags().IntVar(&limit, "limit", 20, "Number of results per page (1-10000)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().StringVar(&fields, "fields", "", `Comma-separated item fields to return (overrides the default projection)`)
	return cmd
}

func runOrderList(cmd *cobra.Command, f *cmdutil.Factory, filters, sorts []string, limit, page int, fields string) error {
	if err := cmdutil.ValidateLimit(limit); err != nil {
		return err
	}
	if err := cmdutil.ValidateListFields(cmd, fields); err != nil {
		return err
	}

	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
	if err != nil {
		return err
	}

	search := magento.NewSearch()
	search.SetPageSize(limit)
	search.SetCurrentPage(page)
	if fields == "" {
		fields = magento.OrderListFields
	}
	search.SetFields(fields)
	for _, expr := range filters {
		if err := search.AddFilter(expr); err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}
	}
	for _, expr := range sorts {
		if err := search.AddSort(expr); err != nil {
			return fmt.Errorf("invalid sort: %w", err)
		}
	}

	result, err := client.ListOrders(cmd.Context(), search)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
		if len(result.Items) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No orders found.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Orders (%d total):\n\n", result.TotalCount)
		_, _ = fmt.Fprintf(ios.Out, "%-12s  %-20s  %-12s  %12s  %-5s  %s\n",
			"ORDER #", "CREATED", "STATUS", "TOTAL", "CUR", "CUSTOMER")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 95))
		for _, o := range result.Items {
			customer := mapStr(o, "customer_email")
			_, _ = fmt.Fprintf(ios.Out, "%-12s  %-20s  %-12s  %12.2f  %-5s  %s\n",
				mapStr(o, "increment_id"), mapStr(o, "created_at"),
				cmdutil.Truncate(mapStr(o, "status"), 12), mapNum(o, "grand_total"),
				mapStr(o, "order_currency_code"), cmdutil.Truncate(customer, 30))
		}
		return nil
	})
}

func newOrderViewCmd(f *cmdutil.Factory) *cobra.Command {
	var entityID int
	var fields string

	cmd := &cobra.Command{
		Use:   "view [<order number>]",
		Short: "View a single order",
		Long: `View a single order by its human-facing order number (increment_id),
or by internal entity ID via --id.

The default projection includes line items and totals; the billing address
is limited to city/region/postcode — request street or telephone via
--fields only if truly needed.

Examples:
  magecli sales order view 000000042
  magecli sales order view --id 42 --json
  magecli sales order view 000000042 --fields "increment_id,items[sku,qty_ordered]" --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			incrementID := ""
			if len(args) == 1 {
				incrementID = args[0]
			}
			if (incrementID == "") == (entityID == 0) {
				return fmt.Errorf("provide an order number, or --id for the internal entity ID (not both)")
			}
			return runOrderView(cmd, f, incrementID, entityID, fields)
		},
	}
	cmd.Flags().IntVar(&entityID, "id", 0, "Look up by internal order entity ID instead of order number")
	cmd.Flags().StringVar(&fields, "fields", "", `Comma-separated fields to return (overrides the default projection)`)
	return cmd
}

func runOrderView(cmd *cobra.Command, f *cmdutil.Factory, incrementID string, entityID int, fields string) error {
	if err := cmdutil.ValidateFields(cmd, fields); err != nil {
		return err
	}

	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
	if err != nil {
		return err
	}

	if fields == "" {
		fields = magento.OrderViewFields
	}

	var order map[string]any
	if entityID != 0 {
		order, err = client.GetOrder(cmd.Context(), entityID, fields)
	} else {
		order, err = client.GetOrderByIncrementID(cmd.Context(), incrementID, fields)
		if err == nil && order == nil {
			return fmt.Errorf("no order was found with order number %q", incrementID)
		}
	}
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, order, func() error {
		_, _ = fmt.Fprintf(ios.Out, "Order #:    %s\n", mapStr(order, "increment_id"))
		_, _ = fmt.Fprintf(ios.Out, "Status:     %s\n", mapStr(order, "status"))
		_, _ = fmt.Fprintf(ios.Out, "Created:    %s\n", mapStr(order, "created_at"))
		cur := mapStr(order, "order_currency_code")
		_, _ = fmt.Fprintf(ios.Out, "Total:      %.2f %s\n", mapNum(order, "grand_total"), cur)
		if refunded := mapNum(order, "total_refunded"); refunded > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Refunded:   %.2f %s\n", refunded, cur)
		}
		name := strings.TrimSpace(mapStr(order, "customer_firstname") + " " + mapStr(order, "customer_lastname"))
		if name != "" || mapStr(order, "customer_email") != "" {
			_, _ = fmt.Fprintf(ios.Out, "Customer:   %s <%s>\n", name, mapStr(order, "customer_email"))
		}
		if shipping := mapStr(order, "shipping_description"); shipping != "" {
			_, _ = fmt.Fprintf(ios.Out, "Shipping:   %s\n", shipping)
		}
		if items, ok := order["items"].([]any); ok && len(items) > 0 {
			_, _ = fmt.Fprintf(ios.Out, "\nItems:\n")
			for _, it := range items {
				item, ok := it.(map[string]any)
				if !ok {
					continue
				}
				_, _ = fmt.Fprintf(ios.Out, "  %-24s  qty %.0f  @ %.2f  = %.2f\n",
					cmdutil.Truncate(mapStr(item, "sku"), 24), mapNum(item, "qty_ordered"),
					mapNum(item, "price"), mapNum(item, "row_total"))
			}
		}
		return nil
	})
}

// --- Invoices / shipments / credit memos share one list implementation ---

type listContext struct {
	ctx    context.Context
	client *magento.Client
	search *magento.SearchCriteria
}

type documentSpec struct {
	use     string
	short   string
	long    string
	example string
	fields  string
	list    func(listContext) (*magento.GenericResult, error)
	empty   string
	title   string
}

func newDocumentCmd(f *cmdutil.Factory, spec documentSpec) *cobra.Command {
	parent := &cobra.Command{
		Use:   spec.use,
		Short: spec.short,
	}
	var filters, sorts []string
	var limit, page int
	var fields string

	cmd := &cobra.Command{
		Use:   "list",
		Short: spec.short,
		Long:  spec.long + "\n\nExamples:\n" + strings.ReplaceAll(spec.example, `\n`, "\n"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocumentList(cmd, f, spec, filters, sorts, limit, page, fields)
		},
	}
	cmd.Flags().StringArrayVar(&filters, "filter", nil, `Filter expression (e.g. "order_id eq 42")`)
	cmd.Flags().StringArrayVar(&sorts, "sort", nil, `Sort expression (e.g. "created_at:DESC")`)
	cmd.Flags().IntVar(&limit, "limit", 20, "Number of results per page (1-10000)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().StringVar(&fields, "fields", "", `Comma-separated item fields to return (overrides the default projection)`)
	parent.AddCommand(cmd)
	return parent
}

func runDocumentList(cmd *cobra.Command, f *cmdutil.Factory, spec documentSpec, filters, sorts []string, limit, page int, fields string) error {
	if err := cmdutil.ValidateLimit(limit); err != nil {
		return err
	}
	if err := cmdutil.ValidateListFields(cmd, fields); err != nil {
		return err
	}

	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
	if err != nil {
		return err
	}

	search := magento.NewSearch()
	search.SetPageSize(limit)
	search.SetCurrentPage(page)
	if fields == "" {
		fields = spec.fields
	}
	search.SetFields(fields)
	for _, expr := range filters {
		if err := search.AddFilter(expr); err != nil {
			return fmt.Errorf("invalid filter: %w", err)
		}
	}
	for _, expr := range sorts {
		if err := search.AddSort(expr); err != nil {
			return fmt.Errorf("invalid sort: %w", err)
		}
	}

	result, err := spec.list(listContext{ctx: cmd.Context(), client: client, search: search})
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
		if len(result.Items) == 0 {
			_, _ = fmt.Fprintln(ios.Out, spec.empty)
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "%s (%d total):\n\n", spec.title, result.TotalCount)
		_, _ = fmt.Fprintf(ios.Out, "%-12s  %-10s  %-20s  %12s  %s\n",
			"DOCUMENT #", "ORDER ID", "CREATED", "TOTAL", "EXTRA")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 80))
		for _, d := range result.Items {
			extra := ""
			if tracks, ok := d["tracks"].([]any); ok {
				var nums []string
				for _, tr := range tracks {
					if track, ok := tr.(map[string]any); ok {
						nums = append(nums, mapStr(track, "track_number"))
					}
				}
				extra = strings.Join(nums, ", ")
			} else if qty := mapNum(d, "total_qty"); qty > 0 {
				extra = fmt.Sprintf("qty %.0f", qty)
			}
			total := mapNum(d, "grand_total")
			_, _ = fmt.Fprintf(ios.Out, "%-12s  %-10.0f  %-20s  %12.2f  %s\n",
				mapStr(d, "increment_id"), mapNum(d, "order_id"),
				mapStr(d, "created_at"), total, extra)
		}
		return nil
	})
}

// --- Totals ---

func newTotalsCmd(f *cmdutil.Factory) *cobra.Command {
	var from, to, status string

	cmd := &cobra.Command{
		Use:   "totals",
		Short: "Sum order totals over a date range, grouped by currency",
		Long: `Sum order grand totals over a date range, grouped by currency.

Note: gross is grand_total (includes tax and shipping), not net revenue.
Scans at most 10,000 orders; the result says if the range was too large
to complete.

Examples:
  magecli sales totals --from 2026-06-01
  magecli sales totals --from 2026-06-01 --to 2026-07-01 --status complete --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(from) == "" {
				return fmt.Errorf("--from is required (e.g. --from 2026-06-01)")
			}
			return runTotals(cmd, f, from, to, status)
		},
	}
	cmd.Flags().StringVar(&from, "from", "", `Start of the period (e.g. "2026-06-01" or "2026-06-01 00:00:00")`)
	cmd.Flags().StringVar(&to, "to", "", "End of the period (defaults to now)")
	cmd.Flags().StringVar(&status, "status", "", `Only count orders with this status (e.g. "complete")`)
	return cmd
}

func runTotals(cmd *cobra.Command, f *cmdutil.Factory, from, to, status string) error {
	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
	if err != nil {
		return err
	}

	totals, err := client.SumSalesTotals(cmd.Context(), from, to, status)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, totals, func() error {
		_, _ = fmt.Fprintf(ios.Out, "Sales totals %s to %s", totals.From, totals.To)
		if status != "" {
			_, _ = fmt.Fprintf(ios.Out, " (status %s)", status)
		}
		_, _ = fmt.Fprintf(ios.Out, ":\n\nOrders: %d\n\n", totals.OrderCount)
		if len(totals.Totals) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No orders in this range.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "%-8s  %14s  %14s\n", "CURRENCY", "GROSS", "REFUNDED")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 42))
		var currencies []string
		for cur := range totals.Totals {
			currencies = append(currencies, cur)
		}
		sort.Strings(currencies)
		for _, cur := range currencies {
			t := totals.Totals[cur]
			_, _ = fmt.Fprintf(ios.Out, "%-8s  %14.2f  %14.2f\n", cur, t.Gross, t.Refunded)
		}
		_, _ = fmt.Fprintln(ios.Out, "\nGross is grand_total (includes tax and shipping), not net revenue.")
		if totals.Note != "" {
			_, _ = fmt.Fprintf(ios.Out, "Note: %s\n", totals.Note)
		}
		return nil
	})
}

// --- Helpers for dynamically-typed payloads ---

func mapStr(m map[string]any, key string) string {
	if s, ok := m[key].(string); ok {
		return s
	}
	return ""
}

func mapNum(m map[string]any, key string) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		n, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return n
		}
	}
	return 0
}
