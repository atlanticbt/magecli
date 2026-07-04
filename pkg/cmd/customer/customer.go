package customer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmdutil"
	"github.com/atlanticbt/magecli/pkg/iostreams"
	"github.com/atlanticbt/magecli/pkg/magento"
)

func NewCmdCustomer(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "customer",
		Short: "Search and view customer accounts",
		Long: `Query Magento customer accounts.

Requires the Integration token to have Customers resource access
(Magento Admin > System > Integrations); without it Magento returns 403.

Names and emails are returned by default — they are the lookup keys —
but postal addresses and phone numbers are excluded unless explicitly
requested with --include-addresses.

Examples:
  magecli customer search --filter "email like %@example.com"
  magecli customer view jane@example.com
  magecli customer view 9 --include-addresses --json`,
	}
	cmd.AddCommand(newSearchCmd(f))
	cmd.AddCommand(newViewCmd(f))
	return cmd
}

func newSearchCmd(f *cmdutil.Factory) *cobra.Command {
	var filters, sorts []string
	var limit, page int
	var fields string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search customer accounts",
		Long: `Search customer accounts with filters and sorting.

Returns names and emails, never addresses.

Examples:
  magecli customer search --filter "email like %@example.com"
  magecli customer search --filter "created_at from 2026-01-01" --sort "created_at:DESC"
  magecli customer search --filter "lastname eq Doe" --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(cmd, f, filters, sorts, limit, page, fields)
		},
	}
	cmd.Flags().StringArrayVar(&filters, "filter", nil, `Filter expression (e.g. "email like %@example.com")`)
	cmd.Flags().StringArrayVar(&sorts, "sort", nil, `Sort expression (e.g. "created_at:DESC")`)
	cmd.Flags().IntVar(&limit, "limit", 20, "Number of results per page (1-10000)")
	cmd.Flags().IntVar(&page, "page", 1, "Page number")
	cmd.Flags().StringVar(&fields, "fields", "", `Comma-separated item fields to return (overrides the default projection)`)
	return cmd
}

func runSearch(cmd *cobra.Command, f *cmdutil.Factory, filters, sorts []string, limit, page int, fields string) error {
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
		fields = magento.CustomerSearchFields
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

	result, err := client.SearchCustomers(cmd.Context(), search)
	if err != nil {
		return err
	}

	return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
		if len(result.Items) == 0 {
			_, _ = fmt.Fprintln(ios.Out, "No customers found.")
			return nil
		}
		_, _ = fmt.Fprintf(ios.Out, "Customers (%d total):\n\n", result.TotalCount)
		_, _ = fmt.Fprintf(ios.Out, "%-6s  %-35s  %-25s  %-8s  %s\n",
			"ID", "EMAIL", "NAME", "WEBSITE", "CREATED")
		_, _ = fmt.Fprintln(ios.Out, strings.Repeat("-", 95))
		for _, c := range result.Items {
			name := strings.TrimSpace(mapStr(c, "firstname") + " " + mapStr(c, "lastname"))
			_, _ = fmt.Fprintf(ios.Out, "%-6.0f  %-35s  %-25s  %-8.0f  %s\n",
				mapNum(c, "id"), cmdutil.Truncate(mapStr(c, "email"), 35),
				cmdutil.Truncate(name, 25), mapNum(c, "website_id"), mapStr(c, "created_at"))
		}
		return nil
	})
}

func newViewCmd(f *cmdutil.Factory) *cobra.Command {
	var includeAddresses bool

	cmd := &cobra.Command{
		Use:   "view <id|email>",
		Short: "View a single customer by ID or email",
		Long: `View a single customer account by numeric ID or email address.

Postal addresses and phone numbers are only returned when
--include-addresses is set.

On multi-website stores the same email can belong to several accounts;
when that happens all matches are listed — pick one by ID.

Examples:
  magecli customer view 9
  magecli customer view jane@example.com --json
  magecli customer view 9 --include-addresses --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(cmd, f, args[0], includeAddresses)
		},
	}
	cmd.Flags().BoolVar(&includeAddresses, "include-addresses", false, "Include postal addresses and phone numbers (sensitive)")
	return cmd
}

func runView(cmd *cobra.Command, f *cmdutil.Factory, key string, includeAddresses bool) error {
	ios, client, err := cmdutil.ClientFromCmd(f, cmd)
	if err != nil {
		return err
	}

	fields := magento.CustomerViewFields
	if includeAddresses {
		fields += magento.CustomerAddressFields
	}

	if !strings.Contains(key, "@") {
		id, err := strconv.Atoi(key)
		if err != nil {
			return fmt.Errorf("argument must be a numeric customer ID or an email address, got %q", key)
		}
		customer, err := client.GetCustomer(cmd.Context(), id, fields)
		if err != nil {
			return err
		}
		return writeCustomer(cmd, ios, customer)
	}

	matches, err := client.FindCustomersByEmail(cmd.Context(), key, fields)
	if err != nil {
		return err
	}
	switch len(matches) {
	case 0:
		return fmt.Errorf("no customer was found with email %q", key)
	case 1:
		return writeCustomer(cmd, ios, matches[0])
	default:
		result := map[string]any{
			"note":    fmt.Sprintf("%d customer accounts share this email (one per website on multi-website stores); all are listed. Re-run with a numeric ID to pick one.", len(matches)),
			"matches": matches,
		}
		return cmdutil.WriteOutput(cmd, ios.Out, result, func() error {
			_, _ = fmt.Fprintf(ios.Out, "%d customer accounts share this email (one per website on multi-website stores).\nRe-run with a numeric ID to pick one:\n\n", len(matches))
			for _, m := range matches {
				name := strings.TrimSpace(mapStr(m, "firstname") + " " + mapStr(m, "lastname"))
				_, _ = fmt.Fprintf(ios.Out, "  ID %.0f  %s  (website %.0f)\n",
					mapNum(m, "id"), name, mapNum(m, "website_id"))
			}
			return nil
		})
	}
}

func writeCustomer(cmd *cobra.Command, ios *iostreams.IOStreams, customer map[string]any) error {
	return cmdutil.WriteOutput(cmd, ios.Out, customer, func() error {
		_, _ = fmt.Fprintf(ios.Out, "ID:        %.0f\n", mapNum(customer, "id"))
		_, _ = fmt.Fprintf(ios.Out, "Email:     %s\n", mapStr(customer, "email"))
		name := strings.TrimSpace(mapStr(customer, "firstname") + " " + mapStr(customer, "lastname"))
		_, _ = fmt.Fprintf(ios.Out, "Name:      %s\n", name)
		_, _ = fmt.Fprintf(ios.Out, "Created:   %s\n", mapStr(customer, "created_at"))
		if updated := mapStr(customer, "updated_at"); updated != "" {
			_, _ = fmt.Fprintf(ios.Out, "Updated:   %s\n", updated)
		}
		_, _ = fmt.Fprintf(ios.Out, "Group:     %.0f\n", mapNum(customer, "group_id"))
		_, _ = fmt.Fprintf(ios.Out, "Website:   %.0f\n", mapNum(customer, "website_id"))
		if addrs, ok := customer["addresses"].([]any); ok {
			_, _ = fmt.Fprintf(ios.Out, "Addresses: %d on file (see --json for details)\n", len(addrs))
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
