package magento

import (
	"context"
	"fmt"
)

// ListCatalogRules retrieves catalog price rules matching search criteria.
func (c *Client) ListCatalogRules(ctx context.Context, search *SearchCriteria) (*SearchResult[CatalogRule], error) {
	path := "/V1/catalogRules/search?" + search.Encode()
	var result SearchResult[CatalogRule]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list catalog rules: %w", err)
	}
	return &result, nil
}

// GetCatalogRule retrieves a single catalog price rule by ID.
func (c *Client) GetCatalogRule(ctx context.Context, id int) (*CatalogRule, error) {
	path := fmt.Sprintf("/V1/catalogRules/%d", id)
	var rule CatalogRule
	if err := c.get(ctx, path, &rule); err != nil {
		return nil, fmt.Errorf("get catalog rule %d: %w", id, err)
	}
	return &rule, nil
}

// ListCartRules retrieves cart price rules (sales rules) matching search criteria.
func (c *Client) ListCartRules(ctx context.Context, search *SearchCriteria) (*SearchResult[CartRule], error) {
	path := "/V1/salesRules/search?" + search.Encode()
	var result SearchResult[CartRule]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list cart rules: %w", err)
	}
	return &result, nil
}

// GetCartRule retrieves a single cart price rule by ID.
func (c *Client) GetCartRule(ctx context.Context, id int) (*CartRule, error) {
	path := fmt.Sprintf("/V1/salesRules/%d", id)
	var rule CartRule
	if err := c.get(ctx, path, &rule); err != nil {
		return nil, fmt.Errorf("get cart rule %d: %w", id, err)
	}
	return &rule, nil
}

// ListCoupons retrieves coupons matching search criteria.
func (c *Client) ListCoupons(ctx context.Context, search *SearchCriteria) (*SearchResult[Coupon], error) {
	path := "/V1/coupons/search?" + search.Encode()
	var result SearchResult[Coupon]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list coupons: %w", err)
	}
	return &result, nil
}

// GetCoupon retrieves a single coupon by ID.
func (c *Client) GetCoupon(ctx context.Context, id int) (*Coupon, error) {
	path := fmt.Sprintf("/V1/coupons/%d", id)
	var coupon Coupon
	if err := c.get(ctx, path, &coupon); err != nil {
		return nil, fmt.Errorf("get coupon %d: %w", id, err)
	}
	return &coupon, nil
}
