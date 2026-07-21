package magento_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/atlanticbt/magecli/pkg/magento"
	"github.com/atlanticbt/magecli/pkg/magentotest"
)

func newClient(t *testing.T, srv *magentotest.Server) *magento.Client {
	t.Helper()
	c, err := magento.New(magento.ClientOptions{
		BaseURL:   srv.URL,
		Token:     "test-token",
		StoreCode: "default",
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestListOrders_DefaultProjection(t *testing.T) {
	srv := magentotest.New(t)
	c := newClient(t, srv)

	search := magento.NewSearch()
	// The fake server 500s on a double-wrapped fields= parameter, so this
	// also proves the default projection is passed unwrapped.
	search.SetFields(magento.OrderListFields)

	result, err := c.ListOrders(context.Background(), search)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalCount != 1 || len(result.Items) != 1 {
		t.Fatalf("got %d items (total %d), want 1", len(result.Items), result.TotalCount)
	}
	if got := result.Items[0]["increment_id"]; got != "000000042" {
		t.Errorf("increment_id = %v, want 000000042", got)
	}
}

func TestGetOrderByIncrementID(t *testing.T) {
	srv := magentotest.New(t)
	c := newClient(t, srv)

	order, err := c.GetOrderByIncrementID(context.Background(), "000000042", magento.OrderViewFields)
	if err != nil {
		t.Fatal(err)
	}
	if order == nil {
		t.Fatal("order not found")
	}
	if got := order["status"]; got != "complete" {
		t.Errorf("status = %v, want complete", got)
	}

	missing, err := c.GetOrderByIncrementID(context.Background(), "000000999", magento.OrderViewFields)
	if err != nil {
		t.Fatal(err)
	}
	if missing != nil {
		t.Errorf("expected nil for unknown order number, got %v", missing)
	}
}

func TestGetOrder_ByEntityID(t *testing.T) {
	srv := magentotest.New(t)
	c := newClient(t, srv)

	order, err := c.GetOrder(context.Background(), 42, magento.OrderViewFields)
	if err != nil {
		t.Fatal(err)
	}
	if got := order["increment_id"]; got != "000000042" {
		t.Errorf("increment_id = %v, want 000000042", got)
	}
}

func TestListDocuments(t *testing.T) {
	srv := magentotest.New(t)
	c := newClient(t, srv)
	ctx := context.Background()

	tests := []struct {
		name   string
		fields string
		list   func(*magento.SearchCriteria) (*magento.GenericResult, error)
	}{
		{"invoices", magento.InvoiceListFields, func(s *magento.SearchCriteria) (*magento.GenericResult, error) {
			return c.ListInvoices(ctx, s)
		}},
		{"shipments", magento.ShipmentListFields, func(s *magento.SearchCriteria) (*magento.GenericResult, error) {
			return c.ListShipments(ctx, s)
		}},
		{"creditmemos", magento.CreditmemoListFields, func(s *magento.SearchCriteria) (*magento.GenericResult, error) {
			return c.ListCreditmemos(ctx, s)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			search := magento.NewSearch()
			search.SetFields(tt.fields)
			result, err := tt.list(search)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Items) != 1 {
				t.Fatalf("got %d items, want 1", len(result.Items))
			}
		})
	}
}

func TestSumSalesTotals(t *testing.T) {
	srv := magentotest.New(t)
	c := newClient(t, srv)

	totals, err := c.SumSalesTotals(context.Background(), "2026-06-01", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if totals.OrderCount != 1 {
		t.Errorf("order count = %d, want 1", totals.OrderCount)
	}
	usd := totals.Totals["USD"]
	if usd == nil {
		t.Fatalf("no USD totals; got %v", totals.Totals)
	}
	if usd.Gross != 129.99 {
		t.Errorf("gross = %v, want 129.99", usd.Gross)
	}
	if usd.Refunded != 10.0 {
		t.Errorf("refunded = %v, want 10.0", usd.Refunded)
	}
	if !totals.Complete {
		t.Error("expected Complete = true")
	}
	if totals.To != "now" {
		t.Errorf("to = %q, want now", totals.To)
	}
}

func TestSumSalesTotals_EmptyRange(t *testing.T) {
	srv := magentotest.New(t)
	c := newClient(t, srv)

	// The fake server returns no orders for from-dates after the fixture.
	totals, err := c.SumSalesTotals(context.Background(), "2026-07-01", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if totals.OrderCount != 0 {
		t.Errorf("order count = %d, want 0", totals.OrderCount)
	}
	if len(totals.Totals) != 0 {
		t.Errorf("totals = %v, want empty", totals.Totals)
	}
}

func TestSales_ForcedError(t *testing.T) {
	srv := magentotest.New(t)
	c := newClient(t, srv)
	srv.ForceStatus(http.StatusForbidden)

	_, err := c.ListOrders(context.Background(), magento.NewSearch())
	if err == nil {
		t.Fatal("expected error from forced 403")
	}
}
