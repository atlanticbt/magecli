package magento

import (
	"context"
	"fmt"
)

// Customer field projections. Names and emails are returned by default —
// they are the lookup keys these commands exist for — but postal addresses,
// phone numbers, and other sensitive attributes are excluded unless the
// caller explicitly opts in (see CustomerAddressFields).
const (
	CustomerSearchFields = "id,email,firstname,lastname,created_at,group_id,store_id,website_id"
	CustomerViewFields   = CustomerSearchFields + ",updated_at,middlename,prefix,suffix"
	// CustomerAddressFields is appended to CustomerViewFields when the
	// caller opts in to postal addresses and phone numbers.
	CustomerAddressFields = ",addresses,default_billing,default_shipping"
)

// SearchCustomers retrieves customer accounts matching search criteria.
// Requires the token's integration to have Customers ACL scopes.
func (c *Client) SearchCustomers(ctx context.Context, search *SearchCriteria) (*GenericResult, error) {
	result, err := c.searchGeneric(ctx, "/V1/customers/search", search)
	if err != nil {
		return nil, fmt.Errorf("search customers: %w", err)
	}
	return result, nil
}

// GetCustomer retrieves a single customer by ID.
func (c *Client) GetCustomer(ctx context.Context, id int, fields string) (map[string]any, error) {
	out, err := c.getGeneric(ctx, fmt.Sprintf("/V1/customers/%d", id), fields)
	if err != nil {
		return nil, fmt.Errorf("get customer %d: %w", id, err)
	}
	return out, nil
}

// FindCustomersByEmail retrieves every customer account with the given
// email. On multi-website stores the same email can belong to several
// accounts; returning an arbitrary one would be silently wrong (found live
// against a multi-website store), so all matches are returned and the
// caller decides how to present them.
func (c *Client) FindCustomersByEmail(ctx context.Context, email, fields string) ([]map[string]any, error) {
	search := NewSearch()
	if err := search.AddFilter("email eq " + email); err != nil {
		return nil, err
	}
	search.SetPageSize(10)
	search.SetFields(fields) // includes website_id, which disambiguates matches
	result, err := c.searchGeneric(ctx, "/V1/customers/search", search)
	if err != nil {
		return nil, fmt.Errorf("find customers by email: %w", err)
	}
	return result.Items, nil
}
