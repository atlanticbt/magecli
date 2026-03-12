package magento

import (
	"context"
	"fmt"
)

// ListStoreViews retrieves all store views.
func (c *Client) ListStoreViews(ctx context.Context) ([]StoreView, error) {
	var views []StoreView
	if err := c.get(ctx, "/V1/store/storeViews", &views); err != nil {
		return nil, fmt.Errorf("list store views: %w", err)
	}
	return views, nil
}

// ListStoreConfigs retrieves store configuration for all store views.
func (c *Client) ListStoreConfigs(ctx context.Context) ([]StoreConfig, error) {
	var configs []StoreConfig
	if err := c.get(ctx, "/V1/store/storeConfigs", &configs); err != nil {
		return nil, fmt.Errorf("list store configs: %w", err)
	}
	return configs, nil
}

// ListStoreGroups retrieves all store groups.
func (c *Client) ListStoreGroups(ctx context.Context) ([]StoreGroup, error) {
	var groups []StoreGroup
	if err := c.get(ctx, "/V1/store/storeGroups", &groups); err != nil {
		return nil, fmt.Errorf("list store groups: %w", err)
	}
	return groups, nil
}

// ListWebsites retrieves all websites.
func (c *Client) ListWebsites(ctx context.Context) ([]Website, error) {
	var websites []Website
	if err := c.get(ctx, "/V1/store/websites", &websites); err != nil {
		return nil, fmt.Errorf("list websites: %w", err)
	}
	return websites, nil
}
