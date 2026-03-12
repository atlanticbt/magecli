package magento

import (
	"context"
	"fmt"
	"net/url"
)

// GetConfigurableChildren returns the simple product variants of a configurable product.
func (c *Client) GetConfigurableChildren(ctx context.Context, sku string) ([]Product, error) {
	path := "/V1/configurable-products/" + url.PathEscape(sku) + "/children"
	var children []Product
	if err := c.get(ctx, path, &children); err != nil {
		return nil, fmt.Errorf("get configurable children %q: %w", sku, err)
	}
	return children, nil
}

// GetConfigurableOptions returns the configurable options for a product.
func (c *Client) GetConfigurableOptions(ctx context.Context, sku string) ([]ConfigurableOption, error) {
	path := "/V1/configurable-products/" + url.PathEscape(sku) + "/options/all"
	var options []ConfigurableOption
	if err := c.get(ctx, path, &options); err != nil {
		return nil, fmt.Errorf("get configurable options %q: %w", sku, err)
	}
	return options, nil
}
