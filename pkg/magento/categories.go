package magento

import (
	"context"
	"fmt"
	"net/url"
)

// GetCategoryTree retrieves the full category tree from a root ID.
func (c *Client) GetCategoryTree(ctx context.Context, rootID int, depth int) (*Category, error) {
	params := url.Values{}
	if rootID > 0 {
		params.Set("rootCategoryId", fmt.Sprintf("%d", rootID))
	}
	if depth > 0 {
		params.Set("depth", fmt.Sprintf("%d", depth))
	}

	path := "/V1/categories"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var tree Category
	if err := c.get(ctx, path, &tree); err != nil {
		return nil, fmt.Errorf("get category tree: %w", err)
	}
	return &tree, nil
}

// GetCategory retrieves a single category by ID.
func (c *Client) GetCategory(ctx context.Context, id int) (*Category, error) {
	path := fmt.Sprintf("/V1/categories/%d", id)
	var cat Category
	if err := c.get(ctx, path, &cat); err != nil {
		return nil, fmt.Errorf("get category %d: %w", id, err)
	}
	return &cat, nil
}

// GetCategoryProducts retrieves products assigned to a category.
func (c *Client) GetCategoryProducts(ctx context.Context, categoryID int) ([]CategoryProduct, error) {
	path := fmt.Sprintf("/V1/categories/%d/products", categoryID)
	var products []CategoryProduct
	if err := c.get(ctx, path, &products); err != nil {
		return nil, fmt.Errorf("get category products %d: %w", categoryID, err)
	}
	return products, nil
}

// CategoryProduct represents a product assignment within a category.
type CategoryProduct struct {
	SKU            string `json:"sku"`
	Position       int    `json:"position"`
	CategoryID     string `json:"category_id"`
}
