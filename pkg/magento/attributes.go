package magento

import (
	"context"
	"fmt"
	"net/url"
)

// GetAttribute retrieves a product attribute by code.
func (c *Client) GetAttribute(ctx context.Context, code string) (*Attribute, error) {
	path := "/V1/products/attributes/" + url.PathEscape(code)
	var attr Attribute
	if err := c.get(ctx, path, &attr); err != nil {
		return nil, fmt.Errorf("get attribute %q: %w", code, err)
	}
	return &attr, nil
}

// GetAttributeOptions retrieves the option values for an attribute.
func (c *Client) GetAttributeOptions(ctx context.Context, code string) ([]AttributeOption, error) {
	path := "/V1/products/attributes/" + url.PathEscape(code) + "/options"
	var options []AttributeOption
	if err := c.get(ctx, path, &options); err != nil {
		return nil, fmt.Errorf("get attribute options %q: %w", code, err)
	}
	return options, nil
}

// ListAttributeSets retrieves attribute sets matching the search criteria.
func (c *Client) ListAttributeSets(ctx context.Context, search *SearchCriteria) (*SearchResult[AttributeSet], error) {
	path := "/V1/products/attribute-sets/sets/list?" + search.Encode()
	var result SearchResult[AttributeSet]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list attribute sets: %w", err)
	}
	return &result, nil
}
