package promo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/magento"
)

func NewCmdPromo(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "promo",
		Aliases: []string{"promotion"},
		Short:   "View catalog price rules, cart price rules, and coupons",
		Long: `Query Magento promotion rules and coupons.

Includes catalog price rules (applied before cart), cart price rules
(applied at checkout), and coupon codes.

Examples:
  magecli promo catalog-rule list
  magecli promo cart-rule list --filter "is_active eq 1"
  magecli promo coupon list --json`,
	}
	cmd.AddCommand(newCatalogRuleCmd(f))
	cmd.AddCommand(newCartRuleCmd(f))
	cmd.AddCommand(newCouponCmd(f))
	return cmd
}

// --- Catalog Price Rules ---

func newCatalogRuleCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "catalog-rule",
		Aliases: []string{"cr"},
		Short:   "Manage catalog price rules",
	}
	cmd.AddCommand(newCatalogRuleListCmd(f))
	cmd.AddCommand(newCatalogRuleViewCmd(f))
	return cmd
}

func newCatalogRuleListCmd(f *cmdutil.Factory) *cobra.Command {
	var filters []string
	var sorts []string
	var limit, page int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List catalog price rules",
		Long: `List catalog price rules with optional filtering and sorting.

Catalog price rules are applied to products before they are added to the cart.

Examples:
  magecli promo catalog-rule list
  magecli promo catalog-rule list --filter "is_active eq 1"
  magecli promo catalog-rule list --json --jq '.items[] | {rule_id, name, discount_amount}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCatalogRuleList(cmd, f, filters, sorts, limit, page)
		},
	}
	cmd.Flags().StringArrayVar(&filters, "filter", nil, `Filter expression (e.g. "is_active eq 1")`)
	cmd.Flags().StringArrayVar(&sorts, "sort", nil, `Sort expression (e.g. "sort_order:ASC")`)
	cmd.Flags().IntVar(&limit, "limit", 50, "Number of results per page")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	return cmd
}

func runCatalogRuleList(cmd *cobra.Command, f *cmdutil.Factory, filters, sorts []string, limit, page int) error {
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

	search := magento.NewSearch()
	search.SetPageSize(limit)
	search.SetCurrentPage(page)
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

	result, err := client.ListCatalogRules(cmd.Context(), search)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
		if len(result.Items) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No catalog price rules found.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Catalog Price Rules (%d total):\n\n", result.TotalCount)
		_, _ = fmt.Fprintf(ios.Out, "%-5s  %-40s  %-12s  %-10s  %-7s  %s\n",
			"ID", "NAME", "ACTION", "DISCOUNT", "ACTIVE", "DATES")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 100))

		for _, r := range result.Items {
			active := "yes"
			if !r.IsActive {
				active = "no"
			}
			dates := formatDateRange(r.FromDate, r.ToDate)
			_, _ = fmt.Fprintf(ios.Out, "%-5d  %-40s  %-12s  %-10.2f  %-7s  %s\n",
				r.RuleID, truncate(r.Name, 40), r.SimpleAction, r.DiscountAmount, active, dates)
		}
		return nil
	})
}

func newCatalogRuleViewCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view <id>",
		Short: "View a catalog price rule by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid rule ID: %w", err)
			}
			return runCatalogRuleView(cmd, f, id)
		},
	}
}

func runCatalogRuleView(cmd *cobra.Command, f *cmdutil.Factory, id int) error {
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

	rule, err := client.GetCatalogRule(cmd.Context(), id)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, rule, func() error {
		_, _ = fmt.Fprintf(ios.Out, "Rule ID:     %d\n", rule.RuleID)
		_, _ = fmt.Fprintf(ios.Out, "Name:        %s\n", rule.Name)
		if rule.Description != "" {
			_, _ = fmt.Fprintf(ios.Out, "Description: %s\n", rule.Description)
		}
		active := "Yes"
		if !rule.IsActive {
			active = "No"
		}
		_, _ = fmt.Fprintf(ios.Out, "Active:      %s\n", active)
		_, _ = fmt.Fprintf(ios.Out, "Action:      %s\n", rule.SimpleAction)
		_, _ = fmt.Fprintf(ios.Out, "Discount:    %.2f\n", rule.DiscountAmount)
		_, _ = fmt.Fprintf(ios.Out, "Priority:    %d\n", rule.SortOrder)
		if rule.StopRulesProcessing {
			_, _ = fmt.Fprintln(ios.Out, "Stop Rules:  Yes")
		}
		dates := formatDateRange(rule.FromDate, rule.ToDate)
		if dates != "" {
			_, _ = fmt.Fprintf(ios.Out, "Dates:       %s\n", dates)
		}
		if len(rule.CustomerGroupIDs) > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Groups:      %s\n", formatIntSlice(rule.CustomerGroupIDs))
		}
		if len(rule.WebsiteIDs) > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Websites:    %s\n", formatIntSlice(rule.WebsiteIDs))
		}
		return nil
	})
}

// --- Cart Price Rules ---

func newCartRuleCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cart-rule",
		Aliases: []string{"sr"},
		Short:   "Manage cart price rules (sales rules)",
	}
	cmd.AddCommand(newCartRuleListCmd(f))
	cmd.AddCommand(newCartRuleViewCmd(f))
	return cmd
}

func newCartRuleListCmd(f *cmdutil.Factory) *cobra.Command {
	var filters []string
	var sorts []string
	var limit, page int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cart price rules",
		Long: `List cart price rules (sales rules) with optional filtering and sorting.

Cart price rules are applied at checkout and may include coupon codes.

Examples:
  magecli promo cart-rule list
  magecli promo cart-rule list --filter "is_active eq 1"
  magecli promo cart-rule list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCartRuleList(cmd, f, filters, sorts, limit, page)
		},
	}
	cmd.Flags().StringArrayVar(&filters, "filter", nil, `Filter expression (e.g. "is_active eq 1")`)
	cmd.Flags().StringArrayVar(&sorts, "sort", nil, `Sort expression (e.g. "sort_order:ASC")`)
	cmd.Flags().IntVar(&limit, "limit", 50, "Number of results per page")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	return cmd
}

func runCartRuleList(cmd *cobra.Command, f *cmdutil.Factory, filters, sorts []string, limit, page int) error {
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

	search := magento.NewSearch()
	search.SetPageSize(limit)
	search.SetCurrentPage(page)
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

	result, err := client.ListCartRules(cmd.Context(), search)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
		if len(result.Items) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No cart price rules found.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Cart Price Rules (%d total):\n\n", result.TotalCount)
		_, _ = fmt.Fprintf(ios.Out, "%-5s  %-35s  %-12s  %-10s  %-7s  %-10s  %s\n",
			"ID", "NAME", "ACTION", "DISCOUNT", "ACTIVE", "COUPON", "DATES")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 105))

		for _, r := range result.Items {
			active := "yes"
			if !r.IsActive {
				active = "no"
			}
			coupon := r.CouponType
			dates := formatDateRange(r.FromDate, r.ToDate)
			_, _ = fmt.Fprintf(ios.Out, "%-5d  %-35s  %-12s  %-10.2f  %-7s  %-10s  %s\n",
				r.RuleID, truncate(r.Name, 35), r.SimpleAction, r.DiscountAmount, active, coupon, dates)
		}
		return nil
	})
}

func newCartRuleViewCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view <id>",
		Short: "View a cart price rule by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid rule ID: %w", err)
			}
			return runCartRuleView(cmd, f, id)
		},
	}
}

func runCartRuleView(cmd *cobra.Command, f *cmdutil.Factory, id int) error {
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

	rule, err := client.GetCartRule(cmd.Context(), id)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, rule, func() error {
		_, _ = fmt.Fprintf(ios.Out, "Rule ID:         %d\n", rule.RuleID)
		_, _ = fmt.Fprintf(ios.Out, "Name:            %s\n", rule.Name)
		if rule.Description != "" {
			_, _ = fmt.Fprintf(ios.Out, "Description:     %s\n", rule.Description)
		}
		active := "Yes"
		if !rule.IsActive {
			active = "No"
		}
		_, _ = fmt.Fprintf(ios.Out, "Active:          %s\n", active)
		_, _ = fmt.Fprintf(ios.Out, "Action:          %s\n", rule.SimpleAction)
		_, _ = fmt.Fprintf(ios.Out, "Discount:        %.2f\n", rule.DiscountAmount)
		if rule.DiscountQty > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Max Qty:         %.0f\n", rule.DiscountQty)
		}
		if rule.DiscountStep > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Step (buy X):    %d\n", rule.DiscountStep)
		}
		_, _ = fmt.Fprintf(ios.Out, "Priority:        %d\n", rule.SortOrder)
		_, _ = fmt.Fprintf(ios.Out, "Coupon Type:     %s\n", rule.CouponType)
		if rule.ApplyToShipping {
			_, _ = fmt.Fprintln(ios.Out, "Apply Shipping:  Yes")
		}
		if rule.StopRulesProcessing {
			_, _ = fmt.Fprintln(ios.Out, "Stop Rules:      Yes")
		}
		if rule.TimesUsed > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Times Used:      %d\n", rule.TimesUsed)
		}
		if rule.UsesPerCoupon > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Uses/Coupon:     %d\n", rule.UsesPerCoupon)
		}
		if rule.UsesPerCustomer > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Uses/Customer:   %d\n", rule.UsesPerCustomer)
		}
		dates := formatDateRange(rule.FromDate, rule.ToDate)
		if dates != "" {
			_, _ = fmt.Fprintf(ios.Out, "Dates:           %s\n", dates)
		}
		if len(rule.CustomerGroupIDs) > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Groups:          %s\n", formatIntSlice(rule.CustomerGroupIDs))
		}
		if len(rule.WebsiteIDs) > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Websites:        %s\n", formatIntSlice(rule.WebsiteIDs))
		}
		return nil
	})
}

// --- Coupons ---

func newCouponCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coupon",
		Short: "Manage coupon codes",
	}
	cmd.AddCommand(newCouponListCmd(f))
	cmd.AddCommand(newCouponViewCmd(f))
	return cmd
}

func newCouponListCmd(f *cmdutil.Factory) *cobra.Command {
	var filters []string
	var sorts []string
	var limit, page int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List coupon codes",
		Long: `List coupon codes with optional filtering and sorting.

Examples:
  magecli promo coupon list
  magecli promo coupon list --filter "code like %SUMMER%"
  magecli promo coupon list --filter "rule_id eq 5" --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCouponList(cmd, f, filters, sorts, limit, page)
		},
	}
	cmd.Flags().StringArrayVar(&filters, "filter", nil, `Filter expression (e.g. "code like %SUMMER%")`)
	cmd.Flags().StringArrayVar(&sorts, "sort", nil, `Sort expression (e.g. "code:ASC")`)
	cmd.Flags().IntVar(&limit, "limit", 50, "Number of results per page")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	return cmd
}

func runCouponList(cmd *cobra.Command, f *cmdutil.Factory, filters, sorts []string, limit, page int) error {
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

	search := magento.NewSearch()
	search.SetPageSize(limit)
	search.SetCurrentPage(page)
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

	result, err := client.ListCoupons(cmd.Context(), search)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
		if len(result.Items) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No coupons found.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Coupons (%d total):\n\n", result.TotalCount)
		_, _ = fmt.Fprintf(ios.Out, "%-5s  %-8s  %-25s  %-8s  %-10s  %s\n",
			"ID", "RULE ID", "CODE", "USED", "LIMIT", "EXPIRES")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 75))

		for _, c := range result.Items {
			usageLimit := "none"
			if c.UsageLimit > 0 {
				usageLimit = strconv.Itoa(c.UsageLimit)
			}
			expires := "-"
			if c.ExpirationDate != "" {
				expires = c.ExpirationDate
			}
			_, _ = fmt.Fprintf(ios.Out, "%-5d  %-8d  %-25s  %-8d  %-10s  %s\n",
				c.CouponID, c.RuleID, truncate(c.Code, 25), c.TimesUsed, usageLimit, expires)
		}
		return nil
	})
}

func newCouponViewCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view <id>",
		Short: "View a coupon by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid coupon ID: %w", err)
			}
			return runCouponView(cmd, f, id)
		},
	}
}

func runCouponView(cmd *cobra.Command, f *cmdutil.Factory, id int) error {
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

	coupon, err := client.GetCoupon(cmd.Context(), id)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, coupon, func() error {
		_, _ = fmt.Fprintf(ios.Out, "Coupon ID:       %d\n", coupon.CouponID)
		_, _ = fmt.Fprintf(ios.Out, "Rule ID:         %d\n", coupon.RuleID)
		_, _ = fmt.Fprintf(ios.Out, "Code:            %s\n", coupon.Code)
		_, _ = fmt.Fprintf(ios.Out, "Times Used:      %d\n", coupon.TimesUsed)
		if coupon.UsageLimit > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Usage Limit:     %d\n", coupon.UsageLimit)
		} else {
			_, _ = fmt.Fprintln(ios.Out, "Usage Limit:     unlimited")
		}
		if coupon.UsagePerCustomer > 0 {
			_, _ = fmt.Fprintf(ios.Out, "Per Customer:    %d\n", coupon.UsagePerCustomer)
		}
		primary := "No"
		if coupon.IsPrimary {
			primary = "Yes"
		}
		_, _ = fmt.Fprintf(ios.Out, "Primary:         %s\n", primary)
		if coupon.CreatedAt != "" {
			_, _ = fmt.Fprintf(ios.Out, "Created:         %s\n", coupon.CreatedAt)
		}
		if coupon.ExpirationDate != "" {
			_, _ = fmt.Fprintf(ios.Out, "Expires:         %s\n", coupon.ExpirationDate)
		}
		return nil
	})
}

// --- Helpers ---

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatDateRange(from, to string) string {
	if from == "" && to == "" {
		return ""
	}
	if from != "" && to != "" {
		return from + " to " + to
	}
	if from != "" {
		return "from " + from
	}
	return "until " + to
}

func formatIntSlice(ids []int) string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = strconv.Itoa(id)
	}
	return strings.Join(strs, ", ")
}
