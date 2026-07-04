package magento

import (
	"context"
	"fmt"
	"net/url"
)

// Default field projections for sales endpoints. A single untrimmed order
// runs 20-60KB of extension attributes; these keep payloads sane while
// covering what list/view output needs. Exported so downstream consumers
// (e.g. magento-mcp) share one definition. Live-verified against
// Magento 2.4.7.
const (
	OrderListFields = "increment_id,entity_id,status,state,created_at,grand_total,total_paid,total_refunded,total_qty_ordered,order_currency_code,customer_email,customer_firstname,customer_lastname,store_id"
	// OrderViewFields adds line items and a billing address projection that
	// deliberately stops at city/region/postcode/country — no street or
	// telephone unless explicitly requested via a fields override.
	OrderViewFields      = OrderListFields + ",subtotal,shipping_amount,tax_amount,discount_amount,shipping_description,payment[method],items[sku,name,product_type,qty_ordered,qty_invoiced,qty_shipped,qty_refunded,price,row_total],billing_address[firstname,lastname,city,region,postcode,country_id]"
	InvoiceListFields    = "entity_id,increment_id,order_id,state,grand_total,order_currency_code,created_at"
	ShipmentListFields   = "entity_id,increment_id,order_id,created_at,total_qty,tracks[track_number,carrier_code,title]"
	CreditmemoListFields = "entity_id,increment_id,order_id,state,grand_total,order_currency_code,created_at"
)

const (
	salesTotalsPageSize = 500
	salesTotalsMaxPages = 20
)

// GenericResult is the Magento search envelope decoded without a typed item
// struct, for endpoints whose payloads are field-projected dynamically
// (orders, invoices, shipments, credit memos, customers).
type GenericResult struct {
	Items      []map[string]any `json:"items"`
	TotalCount int              `json:"total_count"`
}

// searchGeneric GETs a search-envelope endpoint (path?searchCriteria...) and
// decodes the response dynamically.
func (c *Client) searchGeneric(ctx context.Context, path string, search *SearchCriteria) (*GenericResult, error) {
	var result GenericResult
	if err := c.get(ctx, path+"?"+search.Encode(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// getGeneric GETs a single-entity endpoint, decoding dynamically. Unlike
// search endpoints, single-entity endpoints take fields= WITHOUT the
// items[...] wrapper, so fields is passed through as-is.
func (c *Client) getGeneric(ctx context.Context, path, fields string) (map[string]any, error) {
	if fields != "" {
		path += "?fields=" + url.QueryEscape(fields)
	}
	var out map[string]any
	if err := c.get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ListOrders retrieves orders matching search criteria. Requires the token's
// integration to have Sales ACL scopes.
func (c *Client) ListOrders(ctx context.Context, search *SearchCriteria) (*GenericResult, error) {
	result, err := c.searchGeneric(ctx, "/V1/orders", search)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	return result, nil
}

// GetOrder retrieves a single order by its internal entity ID.
func (c *Client) GetOrder(ctx context.Context, entityID int, fields string) (map[string]any, error) {
	out, err := c.getGeneric(ctx, fmt.Sprintf("/V1/orders/%d", entityID), fields)
	if err != nil {
		return nil, fmt.Errorf("get order %d: %w", entityID, err)
	}
	return out, nil
}

// GetOrderByIncrementID retrieves a single order by its human-facing order
// number (increment_id). Returns nil (no error) when no order matches.
func (c *Client) GetOrderByIncrementID(ctx context.Context, incrementID, fields string) (map[string]any, error) {
	search := NewSearch()
	if err := search.AddFilter("increment_id eq " + incrementID); err != nil {
		return nil, err
	}
	search.SetPageSize(1)
	search.SetFields(fields)
	result, err := c.searchGeneric(ctx, "/V1/orders", search)
	if err != nil {
		return nil, fmt.Errorf("get order %s: %w", incrementID, err)
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	return result.Items[0], nil
}

// ListInvoices retrieves invoices (billing documents) matching search criteria.
func (c *Client) ListInvoices(ctx context.Context, search *SearchCriteria) (*GenericResult, error) {
	result, err := c.searchGeneric(ctx, "/V1/invoices", search)
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}
	return result, nil
}

// ListShipments retrieves shipments (including tracking) matching search criteria.
func (c *Client) ListShipments(ctx context.Context, search *SearchCriteria) (*GenericResult, error) {
	result, err := c.searchGeneric(ctx, "/V1/shipments", search)
	if err != nil {
		return nil, fmt.Errorf("list shipments: %w", err)
	}
	return result, nil
}

// ListCreditmemos retrieves credit memos (refund documents) matching search criteria.
func (c *Client) ListCreditmemos(ctx context.Context, search *SearchCriteria) (*GenericResult, error) {
	result, err := c.searchGeneric(ctx, "/V1/creditmemos", search)
	if err != nil {
		return nil, fmt.Errorf("list credit memos: %w", err)
	}
	return result, nil
}

// CurrencyTotals accumulates order totals for one currency.
type CurrencyTotals struct {
	Gross    float64 `json:"gross"`
	Refunded float64 `json:"refunded"`
}

// SalesTotals is the result of summing order totals over a date range.
// Gross is grand_total (includes tax and shipping), not net revenue.
type SalesTotals struct {
	OrderCount int                        `json:"order_count"`
	Totals     map[string]*CurrencyTotals `json:"totals"`
	From       string                     `json:"from"`
	To         string                     `json:"to"`
	Complete   bool                       `json:"complete"`
	Note       string                     `json:"note,omitempty"`
}

// SumSalesTotals sums order grand totals over a date range, grouped by
// currency. to defaults to now when empty; status optionally restricts to
// one order status. It scans at most 10,000 orders (500 per page, 20 pages);
// Complete is false and Note explains when the range was too large.
func (c *Client) SumSalesTotals(ctx context.Context, from, to, status string) (*SalesTotals, error) {
	result := &SalesTotals{
		Totals:   map[string]*CurrencyTotals{},
		From:     from,
		To:       to,
		Complete: true,
	}
	if to == "" {
		result.To = "now"
	}

	for page := 1; ; page++ {
		if page > salesTotalsMaxPages {
			result.Complete = false
			result.Note = fmt.Sprintf("More than %d orders matched; totals cover only the first %d. Narrow the date range and sum the pieces.",
				salesTotalsPageSize*salesTotalsMaxPages, salesTotalsPageSize*salesTotalsMaxPages)
			break
		}
		search := NewSearch()
		if err := search.AddFilter("created_at from " + from); err != nil {
			return nil, fmt.Errorf("invalid from date %q: %w", from, err)
		}
		if to != "" {
			if err := search.AddFilter("created_at to " + to); err != nil {
				return nil, fmt.Errorf("invalid to date %q: %w", to, err)
			}
		}
		if status != "" {
			if err := search.AddFilter("status eq " + status); err != nil {
				return nil, err
			}
		}
		search.SetPageSize(salesTotalsPageSize)
		search.SetCurrentPage(page)
		// SetFields wraps into items[...],total_count itself — pass the
		// unwrapped list (live-verified against Magento 2.4.7).
		search.SetFields("grand_total,total_refunded,order_currency_code,status")

		out, err := c.searchGeneric(ctx, "/V1/orders", search)
		if err != nil {
			return nil, fmt.Errorf("sum sales totals: %w", err)
		}
		for _, item := range out.Items {
			cur, _ := item["order_currency_code"].(string)
			if cur == "" {
				cur = "unknown"
			}
			t := result.Totals[cur]
			if t == nil {
				t = &CurrencyTotals{}
				result.Totals[cur] = t
			}
			if g, ok := item["grand_total"].(float64); ok {
				t.Gross += g
			}
			if rf, ok := item["total_refunded"].(float64); ok {
				t.Refunded += rf
			}
			result.OrderCount++
		}
		if len(out.Items) < salesTotalsPageSize {
			break
		}
	}

	return result, nil
}
