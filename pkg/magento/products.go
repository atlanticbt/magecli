package magento

import (
	"context"
	"fmt"
	"net/url"
)

// ListProducts retrieves products matching the search criteria.
func (c *Client) ListProducts(ctx context.Context, search *SearchCriteria) (*SearchResult[Product], error) {
	path := "/V1/products?" + search.Encode()
	var result SearchResult[Product]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	return &result, nil
}

// GetProduct retrieves a single product by SKU.
func (c *Client) GetProduct(ctx context.Context, sku string) (*Product, error) {
	path := "/V1/products/" + url.PathEscape(sku)
	var product Product
	if err := c.get(ctx, path, &product); err != nil {
		return nil, fmt.Errorf("get product %q: %w", sku, err)
	}
	return &product, nil
}

// GetProductMedia retrieves media gallery entries for a product.
func (c *Client) GetProductMedia(ctx context.Context, sku string) ([]MediaEntry, error) {
	path := "/V1/products/" + url.PathEscape(sku) + "/media"
	var entries []MediaEntry
	if err := c.get(ctx, path, &entries); err != nil {
		return nil, fmt.Errorf("get product media %q: %w", sku, err)
	}
	return entries, nil
}
