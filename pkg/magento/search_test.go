package magento

import (
	"net/url"
	"strings"
	"testing"
)

func TestNewSearch_Defaults(t *testing.T) {
	s := NewSearch()
	encoded := s.Encode()
	vals, err := url.ParseQuery(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got := vals.Get("searchCriteria[pageSize]"); got != "20" {
		t.Errorf("default pageSize = %q, want 20", got)
	}
	if got := vals.Get("searchCriteria[currentPage]"); got != "1" {
		t.Errorf("default currentPage = %q, want 1", got)
	}
}

func TestSetPageSize(t *testing.T) {
	s := NewSearch()
	s.SetPageSize(50)
	vals, _ := url.ParseQuery(s.Encode())
	if got := vals.Get("searchCriteria[pageSize]"); got != "50" {
		t.Errorf("pageSize = %q, want 50", got)
	}
}

func TestSetPageSize_IgnoresZeroAndNegative(t *testing.T) {
	s := NewSearch()
	s.SetPageSize(0)
	vals, _ := url.ParseQuery(s.Encode())
	if got := vals.Get("searchCriteria[pageSize]"); got != "20" {
		t.Errorf("pageSize after SetPageSize(0) = %q, want 20 (default)", got)
	}

	s.SetPageSize(-5)
	vals, _ = url.ParseQuery(s.Encode())
	if got := vals.Get("searchCriteria[pageSize]"); got != "20" {
		t.Errorf("pageSize after SetPageSize(-5) = %q, want 20 (default)", got)
	}
}

func TestSetCurrentPage(t *testing.T) {
	s := NewSearch()
	s.SetCurrentPage(3)
	vals, _ := url.ParseQuery(s.Encode())
	if got := vals.Get("searchCriteria[currentPage]"); got != "3" {
		t.Errorf("currentPage = %q, want 3", got)
	}
}

func TestAddFilter(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		wantField string
		wantOp    string
		wantValue string
		wantErr   bool
	}{
		{"eq filter", "status eq 1", "status", "eq", "1", false},
		{"like filter", "name like %shirt%", "name", "like", "%shirt%", false},
		{"gt filter", "price gt 50", "price", "gt", "50", false},
		{"null filter (no value)", "special_price null", "special_price", "null", "", false},
		{"notnull filter", "description notnull", "description", "notnull", "", false},
		{"in filter", "type_id in simple,configurable", "type_id", "in", "simple,configurable", false},
		{"finset filter", "category_ids finset 42", "category_ids", "finset", "42", false},
		{"empty expression", "", "", "", "", true},
		{"field only", "name", "", "", "", true},
		{"invalid operator", "name badop value", "", "", "", true},
		{"operator needs value", "price gt", "", "", "", true},
		{"case insensitive operator", "price GT 50", "price", "gt", "50", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSearch()
			err := s.AddFilter(tt.expr)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			vals, _ := url.ParseQuery(s.Encode())
			prefix := "searchCriteria[filter_groups][0][filters][0]"
			if got := vals.Get(prefix + "[field]"); got != tt.wantField {
				t.Errorf("field = %q, want %q", got, tt.wantField)
			}
			if got := vals.Get(prefix + "[condition_type]"); got != tt.wantOp {
				t.Errorf("condition_type = %q, want %q", got, tt.wantOp)
			}
			if got := vals.Get(prefix + "[value]"); got != tt.wantValue {
				t.Errorf("value = %q, want %q", got, tt.wantValue)
			}
		})
	}
}

func TestAddFilter_MultipleGroupsAreANDed(t *testing.T) {
	s := NewSearch()
	_ = s.AddFilter("name like %shirt%")
	_ = s.AddFilter("price gt 50")

	vals, _ := url.ParseQuery(s.Encode())

	if got := vals.Get("searchCriteria[filter_groups][0][filters][0][field]"); got != "name" {
		t.Errorf("first filter field = %q, want name", got)
	}
	if got := vals.Get("searchCriteria[filter_groups][1][filters][0][field]"); got != "price" {
		t.Errorf("second filter field = %q, want price", got)
	}
}

func TestAddSort(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		wantField string
		wantDir   string
		wantErr   bool
	}{
		{"asc sort", "price:ASC", "price", "ASC", false},
		{"desc sort", "name:DESC", "name", "DESC", false},
		{"default direction", "price", "price", "ASC", false},
		{"lowercase direction", "price:asc", "price", "ASC", false},
		{"empty field", ":ASC", "", "", true},
		{"invalid direction", "price:RANDOM", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSearch()
			err := s.AddSort(tt.expr)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			vals, _ := url.ParseQuery(s.Encode())
			prefix := "searchCriteria[sortOrders][0]"
			if got := vals.Get(prefix + "[field]"); got != tt.wantField {
				t.Errorf("sort field = %q, want %q", got, tt.wantField)
			}
			if got := vals.Get(prefix + "[direction]"); got != tt.wantDir {
				t.Errorf("sort direction = %q, want %q", got, tt.wantDir)
			}
		})
	}
}

func TestAddSort_Multiple(t *testing.T) {
	s := NewSearch()
	_ = s.AddSort("price:ASC")
	_ = s.AddSort("name:DESC")

	encoded := s.Encode()
	if !strings.Contains(encoded, "sortOrders%5D%5B0%5D%5Bfield%5D=price") &&
		!strings.Contains(encoded, "sortOrders][0][field]=price") {
		t.Errorf("expected first sort on price, got: %s", encoded)
	}
}

func TestAllValidOperators(t *testing.T) {
	operators := []string{"eq", "neq", "gt", "gteq", "lt", "lteq", "like", "nlike", "in", "nin", "from", "to", "finset"}
	for _, op := range operators {
		t.Run(op, func(t *testing.T) {
			s := NewSearch()
			if err := s.AddFilter("field " + op + " value"); err != nil {
				t.Errorf("operator %q should be valid, got error: %v", op, err)
			}
		})
	}

	// null and notnull don't need a value
	for _, op := range []string{"null", "notnull"} {
		t.Run(op, func(t *testing.T) {
			s := NewSearch()
			if err := s.AddFilter("field " + op); err != nil {
				t.Errorf("operator %q should be valid without value, got error: %v", op, err)
			}
		})
	}
}
