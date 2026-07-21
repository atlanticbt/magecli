// Package magentotest provides a fake Magento 2 REST server for tests,
// with canned fixtures for every endpoint the command catalog uses and a
// knob to force error responses. It is exported so that downstream
// consumers of pkg/magento (e.g. magento-mcp) can test against the same
// API-shape assumptions.
package magentotest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// Server is a fake Magento store. All routes are registered under
// /rest/default/V1/..., matching a client constructed with StoreCode
// "default".
type Server struct {
	*httptest.Server

	mu     sync.Mutex
	forced int
}

// ForceStatus makes every subsequent request fail with the given HTTP status
// and a Magento-style error body. Pass 0 to restore normal behavior.
func (s *Server) ForceStatus(code int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.forced = code
}

func (s *Server) forcedStatus() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.forced
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func notFound(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusNotFound, map[string]string{"message": msg})
}

// searchResult mimics Magento's search envelope.
func searchResult(items any, total int) map[string]any {
	return map[string]any{"items": items, "total_count": total}
}

// filterValue extracts the value of a searchCriteria filter on the given
// field from the request query, or "" if absent.
func filterValue(r *http.Request, field string) string {
	return filterValueCond(r, field, "")
}

// filterValueCond extracts the value of a searchCriteria filter on the given
// field AND condition type (e.g. "from", "to", "eq"). An empty cond matches
// any condition — but note a field can appear in several filter groups (a
// created_at range is two filters), so pass cond whenever that's possible:
// query-param map iteration is otherwise nondeterministic.
func filterValueCond(r *http.Request, field, cond string) string {
	q := r.URL.Query()
	for k, vs := range q {
		if strings.HasSuffix(k, "[field]") && len(vs) > 0 && vs[0] == field {
			prefix := strings.TrimSuffix(k, "[field]")
			if cond == "" || q.Get(prefix+"[condition_type]") == cond {
				return q.Get(prefix + "[value]")
			}
		}
	}
	return ""
}

// New starts a fake Magento server with fixtures. It is closed automatically
// when the test finishes.
func New(t *testing.T) *Server {
	t.Helper()
	s := &Server{}
	mux := http.NewServeMux()

	// fixture registers a static JSON response.
	fixture := func(pattern string, v any) {
		mux.HandleFunc("GET "+pattern, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, v)
		})
	}

	product := map[string]any{"sku": "ABC-1", "name": "Test Shirt", "price": 19.99, "status": 1}

	// Products. The list endpoint honors a url_key filter so product url
	// lookups can miss; all other filters return the fixture product.
	mux.HandleFunc("GET /rest/default/V1/products", func(w http.ResponseWriter, r *http.Request) {
		if key := filterValue(r, "url_key"); key != "" && key != "test-shirt" {
			writeJSON(w, http.StatusOK, searchResult([]any{}, 0))
			return
		}
		writeJSON(w, http.StatusOK, searchResult([]any{product}, 1))
	})
	mux.HandleFunc("GET /rest/default/V1/products/{sku}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("sku") != "ABC-1" {
			notFound(w, fmt.Sprintf("The product that was requested doesn't exist: %s", r.PathValue("sku")))
			return
		}
		writeJSON(w, http.StatusOK, product)
	})
	fixture("/rest/default/V1/products/ABC-1/media", []any{
		map[string]any{"id": 1, "file": "/t/s/shirt.jpg", "media_type": "image", "position": 1},
	})
	fixture("/rest/default/V1/configurable-products/ABC-1/children", []any{
		map[string]any{"sku": "ABC-1-S", "name": "Test Shirt S"},
	})
	fixture("/rest/default/V1/configurable-products/ABC-1/options/all", []any{
		map[string]any{"id": 1, "attribute_id": "93", "label": "Size"},
	})

	// Categories
	fixture("/rest/default/V1/categories", map[string]any{
		"id": 2, "name": "Default Category", "children_data": []any{
			map[string]any{"id": 3, "name": "Shirts", "children_data": []any{}},
		},
	})
	mux.HandleFunc("GET /rest/default/V1/categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "3" {
			notFound(w, "No such entity with id = "+r.PathValue("id"))
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"id": 3, "name": "Shirts", "is_active": true})
	})
	fixture("/rest/default/V1/categories/3/products", []any{
		map[string]any{"sku": "ABC-1", "position": 1, "category_id": "3"},
	})

	// Attributes
	mux.HandleFunc("GET /rest/default/V1/products/attributes/{code}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("code") != "color" {
			notFound(w, "No such attribute: "+r.PathValue("code"))
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"attribute_code": "color", "frontend_input": "select",
			"default_frontend_label": "Color",
		})
	})
	fixture("/rest/default/V1/products/attributes/color/options", []any{
		map[string]any{"label": "Red", "value": "4"},
	})
	fixture("/rest/default/V1/products/attribute-sets/sets/list", searchResult([]any{
		map[string]any{"attribute_set_id": 4, "attribute_set_name": "Default"},
	}, 1))

	// Inventory
	mux.HandleFunc("GET /rest/default/V1/stockItems/{sku}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("sku") != "ABC-1" {
			notFound(w, "The stock item wasn't found.")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item_id": 1, "qty": 42, "is_in_stock": true})
	})

	// Store
	fixture("/rest/default/V1/store/storeViews", []any{
		map[string]any{"id": 1, "code": "default", "name": "Default Store View", "is_active": 1},
	})
	fixture("/rest/default/V1/store/storeConfigs", []any{
		map[string]any{
			"id": 1, "code": "default", "locale": "en_US",
			"base_currency_code": "USD", "timezone": "America/New_York",
			"base_url": "https://shop.example.com/",
		},
	})
	fixture("/rest/default/V1/store/storeGroups", []any{
		map[string]any{"id": 1, "code": "main", "name": "Main Website Store", "root_category_id": 2},
	})
	fixture("/rest/default/V1/store/websites", []any{
		map[string]any{"id": 1, "code": "base", "name": "Main Website"},
	})

	// Promotions
	fixture("/rest/default/V1/catalogRules/search", searchResult([]any{
		map[string]any{"rule_id": 1, "name": "10% off shirts", "is_active": true},
	}, 1))
	fixture("/rest/default/V1/catalogRules/1", map[string]any{
		"rule_id": 1, "name": "10% off shirts", "is_active": true, "discount_amount": 10,
	})
	fixture("/rest/default/V1/salesRules/search", searchResult([]any{
		map[string]any{"rule_id": 2, "name": "Free shipping over $50", "is_active": true},
	}, 1))
	fixture("/rest/default/V1/salesRules/2", map[string]any{
		"rule_id": 2, "name": "Free shipping over $50", "is_active": true,
	})
	fixture("/rest/default/V1/coupons/search", searchResult([]any{
		map[string]any{"coupon_id": 7, "code": "SAVE10", "rule_id": 2},
	}, 1))
	fixture("/rest/default/V1/coupons/7", map[string]any{
		"coupon_id": 7, "code": "SAVE10", "rule_id": 2, "times_used": 3,
	})

	// CMS — fixtures include content so tests can verify it is excluded by
	// default. When fields= excludes content, strip it, mimicking Magento's
	// response filter.
	cmsPage := map[string]any{
		"id": 5, "identifier": "about-us", "title": "About Us",
		"active": true, "content": "<h1>About</h1>",
	}
	cmsBlock := map[string]any{
		"id": 9, "identifier": "footer-links", "title": "Footer Links",
		"active": true, "content": "<ul><li>Links</li></ul>",
	}
	stripContent := func(v map[string]any, r *http.Request) map[string]any {
		fields := r.URL.Query().Get("fields")
		if fields == "" {
			return v
		}
		out := map[string]any{}
		for k, val := range v {
			if k == "content" {
				continue
			}
			out[k] = val
		}
		return out
	}
	mux.HandleFunc("GET /rest/default/V1/cmsPage/search", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, searchResult([]any{stripContent(cmsPage, r)}, 1))
	})
	mux.HandleFunc("GET /rest/default/V1/cmsPage/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "5" {
			notFound(w, "No such page.")
			return
		}
		writeJSON(w, http.StatusOK, stripContent(cmsPage, r))
	})
	mux.HandleFunc("GET /rest/default/V1/cmsBlock/search", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, searchResult([]any{stripContent(cmsBlock, r)}, 1))
	})
	mux.HandleFunc("GET /rest/default/V1/cmsBlock/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "9" {
			notFound(w, "No such block.")
			return
		}
		writeJSON(w, http.StatusOK, stripContent(cmsBlock, r))
	})

	// Modules endpoint used only by raw api escape-hatch tests.
	fixture("/rest/default/V1/modules", []string{"Magento_Catalog", "Magento_Cms"})

	// Sales. The orders list honors increment_id and created_at filters so
	// view-by-number and sales-totals date ranges are testable.
	order := map[string]any{
		"entity_id": 42, "increment_id": "000000042", "status": "complete",
		"state": "complete", "created_at": "2026-06-15 12:00:00",
		"grand_total": 129.99, "total_paid": 129.99, "total_refunded": 10.0,
		"total_qty_ordered": 2, "order_currency_code": "USD",
		"customer_email": "jane@example.com", "customer_firstname": "Jane",
		"customer_lastname": "Doe", "store_id": 1,
		"items": []any{map[string]any{"sku": "ABC-1", "name": "Test Shirt", "qty_ordered": 2, "price": 59.99, "row_total": 119.98}},
		"billing_address": map[string]any{
			"firstname": "Jane", "lastname": "Doe", "city": "Raleigh",
			"region": "North Carolina", "postcode": "27601", "country_id": "US",
			"street": []string{"1 Main St"}, "telephone": "555-0100",
		},
	}
	mux.HandleFunc("GET /rest/default/V1/orders", func(w http.ResponseWriter, r *http.Request) {
		if inc := filterValue(r, "increment_id"); inc != "" && inc != "000000042" {
			writeJSON(w, http.StatusOK, searchResult([]any{}, 0))
			return
		}
		if from := filterValueCond(r, "created_at", "from"); from != "" && from > "2026-06-15 12:00:00" {
			writeJSON(w, http.StatusOK, searchResult([]any{}, 0))
			return
		}
		writeJSON(w, http.StatusOK, searchResult([]any{order}, 1))
	})
	mux.HandleFunc("GET /rest/default/V1/orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "42" {
			notFound(w, "Requested entity doesn't exist")
			return
		}
		writeJSON(w, http.StatusOK, order)
	})
	fixture("/rest/default/V1/invoices", searchResult([]any{
		map[string]any{"entity_id": 7, "increment_id": "000000007", "order_id": 42, "state": 2, "grand_total": 129.99, "order_currency_code": "USD"},
	}, 1))
	fixture("/rest/default/V1/shipments", searchResult([]any{
		map[string]any{"entity_id": 3, "increment_id": "000000003", "order_id": 42, "total_qty": 2,
			"tracks": []any{map[string]any{"track_number": "1Z999", "carrier_code": "ups", "title": "UPS Ground"}}},
	}, 1))
	fixture("/rest/default/V1/creditmemos", searchResult([]any{
		map[string]any{"entity_id": 5, "increment_id": "000000005", "order_id": 42, "state": 2, "grand_total": 10.0, "order_currency_code": "USD"},
	}, 1))

	// Customers. Search honors the email filter; the fixture includes an
	// addresses array so the exclude-by-default behavior is testable.
	customer := map[string]any{
		"id": 9, "email": "jane@example.com", "firstname": "Jane", "lastname": "Doe",
		"created_at": "2025-01-01 00:00:00", "group_id": 1, "store_id": 1, "website_id": 1,
		"addresses": []any{map[string]any{
			"city": "Raleigh", "street": []string{"1 Main St"}, "telephone": "555-0100",
		}},
	}
	stripAddresses := func(v map[string]any, r *http.Request) map[string]any {
		fields := r.URL.Query().Get("fields")
		if fields == "" || strings.Contains(fields, "addresses") {
			return v
		}
		out := map[string]any{}
		for k, val := range v {
			if k == "addresses" {
				continue
			}
			out[k] = val
		}
		return out
	}
	// bob@ exists twice — one account per website — to exercise the
	// multi-website duplicate-email path.
	bob1 := map[string]any{"id": 21, "email": "bob@example.com", "firstname": "Bob", "lastname": "One", "website_id": 1}
	bob2 := map[string]any{"id": 22, "email": "bob@example.com", "firstname": "Bob", "lastname": "Two", "website_id": 2}
	mux.HandleFunc("GET /rest/default/V1/customers/search", func(w http.ResponseWriter, r *http.Request) {
		// Treat % as a like-wildcard: strip it and substring-match.
		email := strings.ReplaceAll(filterValue(r, "email"), "%", "")
		var items []any
		for _, cust := range []map[string]any{customer, bob1, bob2} {
			if email == "" || strings.Contains(cust["email"].(string), email) {
				items = append(items, stripAddresses(cust, r))
			}
		}
		writeJSON(w, http.StatusOK, searchResult(items, len(items)))
	})
	mux.HandleFunc("GET /rest/default/V1/customers/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "9" {
			notFound(w, "No such entity with customerId = "+r.PathValue("id"))
			return
		}
		writeJSON(w, http.StatusOK, stripAddresses(customer, r))
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if code := s.forcedStatus(); code != 0 {
			writeJSON(w, code, map[string]string{"message": "forced error"})
			return
		}
		// SearchCriteria.SetFields wraps its input in items[...],total_count;
		// passing an already-wrapped list produces items[items[...]] and
		// Magento returns useless empty objects. Fail loudly so tests catch
		// the double-wrap (found live against Magento 2.4.7).
		if strings.Contains(r.URL.Query().Get("fields"), "items[items[") {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"message": "magentotest: double-wrapped fields= parameter (items[items[...]): pass SetFields the unwrapped list",
			})
			return
		}
		mux.ServeHTTP(w, r)
	})

	s.Server = httptest.NewServer(handler)
	t.Cleanup(s.Close)
	return s
}
