package magento

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListCatalogRules(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SearchResult[CatalogRule]{
			Items: []CatalogRule{
				{RuleID: 1, Name: "10% Off Electronics", SimpleAction: "by_percent", DiscountAmount: 10, IsActive: true},
			},
			TotalCount: 1,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	result, err := c.ListCatalogRules(context.Background(), NewSearch())
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", result.TotalCount)
	}
	if result.Items[0].Name != "10% Off Electronics" {
		t.Errorf("Name = %q, want '10%% Off Electronics'", result.Items[0].Name)
	}
}

func TestGetCatalogRule(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CatalogRule{
			RuleID:          1,
			Name:            "Summer Sale",
			Description:     "20% off summer items",
			IsActive:        true,
			SimpleAction:    "by_percent",
			DiscountAmount:  20,
			FromDate:        "2026-06-01",
			ToDate:          "2026-08-31",
			CustomerGroupIDs: []int{0, 1},
			WebsiteIDs:      []int{1},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.GetCatalogRule(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if rule.Name != "Summer Sale" {
		t.Errorf("Name = %q, want 'Summer Sale'", rule.Name)
	}
	if rule.DiscountAmount != 20 {
		t.Errorf("DiscountAmount = %f, want 20", rule.DiscountAmount)
	}
	if !rule.IsActive {
		t.Error("expected rule to be active")
	}
	if len(rule.CustomerGroupIDs) != 2 {
		t.Errorf("CustomerGroupIDs len = %d, want 2", len(rule.CustomerGroupIDs))
	}
}

func TestListCartRules(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SearchResult[CartRule]{
			Items: []CartRule{
				{
					RuleID:         5,
					Name:           "Free Shipping Over $100",
					SimpleAction:   "by_percent",
					DiscountAmount: 100,
					IsActive:       true,
					CouponType:     "NO_COUPON",
				},
			},
			TotalCount: 1,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	result, err := c.ListCartRules(context.Background(), NewSearch())
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", result.TotalCount)
	}
	if result.Items[0].CouponType != "NO_COUPON" {
		t.Errorf("CouponType = %q, want NO_COUPON", result.Items[0].CouponType)
	}
}

func TestGetCartRule(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CartRule{
			RuleID:          5,
			Name:            "Buy 3 Get 1 Free",
			IsActive:        true,
			SimpleAction:    "buy_x_get_y",
			DiscountAmount:  1,
			DiscountStep:    3,
			CouponType:      "SPECIFIC_COUPON",
			UsesPerCoupon:   100,
			UsesPerCustomer: 1,
			TimesUsed:       42,
			ApplyToShipping: false,
			FromDate:        "2026-01-01",
			ToDate:          "2026-12-31",
			CustomerGroupIDs: []int{1},
			WebsiteIDs:      []int{1},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	rule, err := c.GetCartRule(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if rule.Name != "Buy 3 Get 1 Free" {
		t.Errorf("Name = %q, want 'Buy 3 Get 1 Free'", rule.Name)
	}
	if rule.DiscountStep != 3 {
		t.Errorf("DiscountStep = %d, want 3", rule.DiscountStep)
	}
	if rule.TimesUsed != 42 {
		t.Errorf("TimesUsed = %d, want 42", rule.TimesUsed)
	}
	if rule.CouponType != "SPECIFIC_COUPON" {
		t.Errorf("CouponType = %q, want SPECIFIC_COUPON", rule.CouponType)
	}
}

func TestListCoupons(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SearchResult[Coupon]{
			Items: []Coupon{
				{CouponID: 1, RuleID: 5, Code: "SUMMER20", TimesUsed: 10, UsageLimit: 100},
				{CouponID: 2, RuleID: 5, Code: "WINTER15", TimesUsed: 0, UsageLimit: 50},
			},
			TotalCount: 2,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	result, err := c.ListCoupons(context.Background(), NewSearch())
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
	if result.Items[0].Code != "SUMMER20" {
		t.Errorf("Code = %q, want SUMMER20", result.Items[0].Code)
	}
}

func TestGetCoupon(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Coupon{
			CouponID:        1,
			RuleID:          5,
			Code:            "SUMMER20",
			UsageLimit:      100,
			UsagePerCustomer: 1,
			TimesUsed:       42,
			IsPrimary:       true,
			CreatedAt:       "2026-01-15 10:00:00",
			ExpirationDate:  "2026-12-31 23:59:59",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	coupon, err := c.GetCoupon(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if coupon.Code != "SUMMER20" {
		t.Errorf("Code = %q, want SUMMER20", coupon.Code)
	}
	if coupon.TimesUsed != 42 {
		t.Errorf("TimesUsed = %d, want 42", coupon.TimesUsed)
	}
	if !coupon.IsPrimary {
		t.Error("expected coupon to be primary")
	}
	if coupon.UsageLimit != 100 {
		t.Errorf("UsageLimit = %d, want 100", coupon.UsageLimit)
	}
}

func TestListCatalogRules_RequestPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(SearchResult[CatalogRule]{})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, _ = c.ListCatalogRules(context.Background(), NewSearch())

	want := "/rest/default/V1/catalogRules/search"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestListCartRules_RequestPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(SearchResult[CartRule]{})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, _ = c.ListCartRules(context.Background(), NewSearch())

	want := "/rest/default/V1/salesRules/search"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestListCoupons_RequestPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(SearchResult[Coupon]{})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, _ = c.ListCoupons(context.Background(), NewSearch())

	want := "/rest/default/V1/coupons/search"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestGetCatalogRule_RequestPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(CatalogRule{})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, _ = c.GetCatalogRule(context.Background(), 42)

	want := "/rest/default/V1/catalogRules/42"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestGetCartRule_RequestPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(CartRule{})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, _ = c.GetCartRule(context.Background(), 7)

	want := "/rest/default/V1/salesRules/7"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestGetCoupon_RequestPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewEncoder(w).Encode(Coupon{})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, _ = c.GetCoupon(context.Background(), 99)

	want := "/rest/default/V1/coupons/99"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}
