package magento

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetStoreConfigRaw(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":                          float64(1),
				"code":                        "default",
				"website_id":                  float64(0),
				"locale":                      "en_US",
				"base_currency_code":          "USD",
				"default_display_currency_code": "USD",
				"timezone":                    "America/New_York",
				"weight_unit":                 "lbs",
				"base_url":                    "https://example.com/",
				"secure_base_url":             "https://example.com/",
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	raw, err := c.GetStoreConfigRaw(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 1 {
		t.Fatalf("got %d configs, want 1", len(raw))
	}
	if raw[0]["locale"] != "en_US" {
		t.Errorf("locale = %v, want en_US", raw[0]["locale"])
	}
}

func TestGetStoreConfigRaw_MultipleStores(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": float64(1), "code": "default", "locale": "en_US"},
			{"id": float64(2), "code": "french", "locale": "fr_FR"},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	raw, err := c.GetStoreConfigRaw(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 2 {
		t.Fatalf("got %d configs, want 2", len(raw))
	}
}

func TestGetConfigEntries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":               float64(1),
				"code":             "default",
				"website_id":       float64(0),
				"locale":           "en_US",
				"base_currency_code": "USD",
				"timezone":         "America/New_York",
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	entries, err := c.GetConfigEntries(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected entries, got none")
	}

	// Verify metadata keys (id, code, website_id) are excluded
	for _, e := range entries {
		if e.Path == "id" || e.Path == "code" || e.Path == "website_id" {
			t.Errorf("metadata key %q should be excluded", e.Path)
		}
	}

	// Verify known path mapping
	found := false
	for _, e := range entries {
		if e.Path == "general/locale/code" && e.Value == "en_US" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected entry for general/locale/code = en_US")
	}
}

func TestGetConfigEntries_FilterByStore(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": float64(1), "code": "default", "locale": "en_US"},
			{"id": float64(2), "code": "french", "locale": "fr_FR"},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	entries, err := c.GetConfigEntries(context.Background(), "french")
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if e.Scope != "french" {
			t.Errorf("expected scope 'french', got %q", e.Scope)
		}
	}

	found := false
	for _, e := range entries {
		if e.Path == "general/locale/code" && e.Value == "fr_FR" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected entry for general/locale/code = fr_FR")
	}
}

func TestGetConfigEntries_SortedByPathWithinScope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":               float64(1),
				"code":             "default",
				"timezone":         "America/New_York",
				"locale":           "en_US",
				"base_currency_code": "USD",
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	entries, err := c.GetConfigEntries(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i < len(entries); i++ {
		if entries[i].Path < entries[i-1].Path {
			t.Errorf("entries not sorted: %q came after %q", entries[i].Path, entries[i-1].Path)
		}
	}
}

func TestGetConfigEntries_UnmappedKeysUseSyntheticPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":             float64(1),
				"code":           "default",
				"some_new_field": "some_value",
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	entries, err := c.GetConfigEntries(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, e := range entries {
		if e.Path == "store/some_new_field" && e.Value == "some_value" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected unmapped key to appear as store/some_new_field")
	}
}

func TestFilterConfigEntries(t *testing.T) {
	entries := []ConfigEntry{
		{Path: "general/locale/code", Value: "en_US", Scope: "default"},
		{Path: "general/locale/timezone", Value: "America/New_York", Scope: "default"},
		{Path: "currency/options/base", Value: "USD", Scope: "default"},
		{Path: "web/secure/base_url", Value: "https://example.com/", Scope: "default"},
		{Path: "web/unsecure/base_url", Value: "http://example.com/", Scope: "default"},
	}

	tests := []struct {
		name   string
		filter string
		want   int
	}{
		{name: "empty filter returns all", filter: "", want: 5},
		{name: "prefix match", filter: "general/locale", want: 2},
		{name: "keyword match", filter: "currency", want: 1},
		{name: "partial path match", filter: "web/", want: 2},
		{name: "case insensitive", filter: "WEB/SECURE", want: 1},
		{name: "no match", filter: "nonexistent/path", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterConfigEntries(entries, tt.filter)
			if len(got) != tt.want {
				t.Errorf("FilterConfigEntries(%q) returned %d entries, want %d", tt.filter, len(got), tt.want)
			}
		})
	}
}

func TestFilterConfigEntries_ContainsMatch(t *testing.T) {
	entries := []ConfigEntry{
		{Path: "web/secure/base_url", Value: "https://example.com/", Scope: "default"},
		{Path: "web/unsecure/base_url", Value: "http://example.com/", Scope: "default"},
		{Path: "general/locale/code", Value: "en_US", Scope: "default"},
	}

	// "secure" should match both web/secure and web/unsecure since it's a contains match
	got := FilterConfigEntries(entries, "secure")
	if len(got) != 2 {
		t.Errorf("expected 2 entries containing 'secure', got %d", len(got))
	}
}

func TestConfigPathMapping(t *testing.T) {
	// Verify all expected mappings exist
	expected := map[string]string{
		"locale":                        "general/locale/code",
		"base_currency_code":            "currency/options/base",
		"default_display_currency_code": "currency/options/default",
		"timezone":                      "general/locale/timezone",
		"weight_unit":                   "general/locale/weight_unit",
		"base_url":                      "web/unsecure/base_url",
		"base_link_url":                 "web/unsecure/base_link_url",
		"base_static_url":               "web/unsecure/base_static_url",
		"base_media_url":                "web/unsecure/base_media_url",
		"secure_base_url":               "web/secure/base_url",
		"secure_base_link_url":          "web/secure/base_link_url",
		"secure_base_static_url":        "web/secure/base_static_url",
		"secure_base_media_url":         "web/secure/base_media_url",
	}

	for key, wantPath := range expected {
		gotPath, ok := jsonKeyToConfigPath[key]
		if !ok {
			t.Errorf("missing mapping for JSON key %q", key)
			continue
		}
		if gotPath != wantPath {
			t.Errorf("jsonKeyToConfigPath[%q] = %q, want %q", key, gotPath, wantPath)
		}
	}
}
