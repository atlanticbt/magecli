package magento_test

import (
	"context"
	"testing"

	"github.com/atlanticbt/magecli/pkg/magento"
	"github.com/atlanticbt/magecli/pkg/magentotest"
)

func TestSearchCustomers_DefaultProjectionExcludesAddresses(t *testing.T) {
	srv := magentotest.New(t)
	c := newClient(t, srv)

	search := magento.NewSearch()
	search.SetFields(magento.CustomerSearchFields)
	if err := search.AddFilter("email like %jane@example.com%"); err != nil {
		t.Fatal(err)
	}

	result, err := c.SearchCustomers(context.Background(), search)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(result.Items))
	}
	if _, ok := result.Items[0]["addresses"]; ok {
		t.Error("default projection must not include addresses")
	}
}

func TestGetCustomer_AddressesOptIn(t *testing.T) {
	srv := magentotest.New(t)
	c := newClient(t, srv)
	ctx := context.Background()

	without, err := c.GetCustomer(ctx, 9, magento.CustomerViewFields)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := without["addresses"]; ok {
		t.Error("addresses returned without opt-in")
	}

	with, err := c.GetCustomer(ctx, 9, magento.CustomerViewFields+magento.CustomerAddressFields)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := with["addresses"]; !ok {
		t.Error("addresses missing despite opt-in")
	}
}

func TestFindCustomersByEmail_MultiWebsiteReturnsAllMatches(t *testing.T) {
	srv := magentotest.New(t)
	c := newClient(t, srv)

	// bob@example.com has one account per website in the fixtures; an
	// arbitrary single match would be silently wrong.
	matches, err := c.FindCustomersByEmail(context.Background(), "bob@example.com", magento.CustomerViewFields)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2", len(matches))
	}

	none, err := c.FindCustomersByEmail(context.Background(), "nobody@example.com", magento.CustomerViewFields)
	if err != nil {
		t.Fatal(err)
	}
	if len(none) != 0 {
		t.Errorf("got %d matches for unknown email, want 0", len(none))
	}
}
