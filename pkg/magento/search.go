package magento

import (
	"fmt"
	"net/url"
	"strings"
)

// SearchCriteria builds Magento 2 REST API search query parameters.
type SearchCriteria struct {
	filterGroups []filterGroup
	sortOrders   []sortOrder
	pageSize     int
	currentPage  int
}

type filterGroup struct {
	filters []filter
}

type filter struct {
	field     string
	value     string
	condition string
}

type sortOrder struct {
	field     string
	direction string
}

// NewSearch creates a new SearchCriteria builder.
func NewSearch() *SearchCriteria {
	return &SearchCriteria{
		pageSize:    20,
		currentPage: 1,
	}
}

// AddFilter adds a filter to a new filter group (AND with other groups).
// Format: "field operator value" e.g. "name like %shirt%", "price gt 50"
func (s *SearchCriteria) AddFilter(expr string) error {
	f, err := parseFilter(expr)
	if err != nil {
		return err
	}
	s.filterGroups = append(s.filterGroups, filterGroup{filters: []filter{f}})
	return nil
}

// AddSort adds a sort order. Format: "field:direction" e.g. "price:ASC"
func (s *SearchCriteria) AddSort(expr string) error {
	parts := strings.SplitN(expr, ":", 2)
	field := strings.TrimSpace(parts[0])
	if field == "" {
		return fmt.Errorf("sort field is required")
	}
	direction := "ASC"
	if len(parts) == 2 {
		direction = strings.ToUpper(strings.TrimSpace(parts[1]))
	}
	if direction != "ASC" && direction != "DESC" {
		return fmt.Errorf("sort direction must be ASC or DESC, got %q", direction)
	}
	s.sortOrders = append(s.sortOrders, sortOrder{field: field, direction: direction})
	return nil
}

// MaxPageSize is the upper bound for results per page to prevent
// accidental resource exhaustion on large catalogs.
const MaxPageSize = 10000

// SetPageSize sets the number of results per page (capped at MaxPageSize).
func (s *SearchCriteria) SetPageSize(size int) {
	if size > 0 {
		if size > MaxPageSize {
			size = MaxPageSize
		}
		s.pageSize = size
	}
}

// SetCurrentPage sets the page number (1-based).
func (s *SearchCriteria) SetCurrentPage(page int) {
	if page > 0 {
		s.currentPage = page
	}
}

// Encode returns the search criteria as URL query parameters.
func (s *SearchCriteria) Encode() string {
	params := url.Values{}

	for i, group := range s.filterGroups {
		for j, f := range group.filters {
			prefix := fmt.Sprintf("searchCriteria[filter_groups][%d][filters][%d]", i, j)
			params.Set(prefix+"[field]", f.field)
			params.Set(prefix+"[value]", f.value)
			params.Set(prefix+"[condition_type]", f.condition)
		}
	}

	for i, so := range s.sortOrders {
		prefix := fmt.Sprintf("searchCriteria[sortOrders][%d]", i)
		params.Set(prefix+"[field]", so.field)
		params.Set(prefix+"[direction]", so.direction)
	}

	params.Set("searchCriteria[pageSize]", fmt.Sprintf("%d", s.pageSize))
	params.Set("searchCriteria[currentPage]", fmt.Sprintf("%d", s.currentPage))

	return params.Encode()
}

var validOperators = map[string]bool{
	"eq": true, "neq": true,
	"gt": true, "gteq": true,
	"lt": true, "lteq": true,
	"like": true, "nlike": true,
	"in": true, "nin": true,
	"null": true, "notnull": true,
	"from": true, "to": true,
	"finset": true,
}

func parseFilter(expr string) (filter, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return filter{}, fmt.Errorf("empty filter expression")
	}

	parts := strings.SplitN(expr, " ", 3)
	if len(parts) < 2 {
		return filter{}, fmt.Errorf("filter must be 'field operator [value]', got %q", expr)
	}

	field := strings.TrimSpace(parts[0])
	op := strings.ToLower(strings.TrimSpace(parts[1]))

	if !validOperators[op] {
		return filter{}, fmt.Errorf("unknown filter operator %q (valid: eq, neq, gt, gteq, lt, lteq, like, nlike, in, nin, null, notnull, from, to, finset)", op)
	}

	value := ""
	if len(parts) == 3 {
		value = strings.TrimSpace(parts[2])
	}

	// null/notnull don't require a value
	if op != "null" && op != "notnull" && value == "" {
		return filter{}, fmt.Errorf("filter operator %q requires a value", op)
	}

	return filter{field: field, value: value, condition: op}, nil
}
