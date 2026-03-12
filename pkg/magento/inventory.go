package magento

import (
	"context"
	"fmt"
	"net/url"
)

// GetStockStatus retrieves the stock/inventory status for a product SKU.
func (c *Client) GetStockStatus(ctx context.Context, sku string) (*StockItem, error) {
	path := "/V1/stockItems/" + url.PathEscape(sku)
	var item StockItem
	if err := c.get(ctx, path, &item); err != nil {
		return nil, fmt.Errorf("get stock status %q: %w", sku, err)
	}
	return &item, nil
}
