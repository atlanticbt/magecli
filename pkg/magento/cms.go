package magento

import (
	"context"
	"fmt"
)

// ListCMSPages retrieves CMS pages matching search criteria.
func (c *Client) ListCMSPages(ctx context.Context, search *SearchCriteria) (*SearchResult[CMSPage], error) {
	path := "/V1/cmsPage/search?" + search.Encode()
	var result SearchResult[CMSPage]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list CMS pages: %w", err)
	}
	return &result, nil
}

// GetCMSPage retrieves a single CMS page by ID.
func (c *Client) GetCMSPage(ctx context.Context, id int) (*CMSPage, error) {
	path := fmt.Sprintf("/V1/cmsPage/%d", id)
	var page CMSPage
	if err := c.get(ctx, path, &page); err != nil {
		return nil, fmt.Errorf("get CMS page %d: %w", id, err)
	}
	return &page, nil
}

// ListCMSBlocks retrieves CMS blocks matching search criteria.
func (c *Client) ListCMSBlocks(ctx context.Context, search *SearchCriteria) (*SearchResult[CMSBlock], error) {
	path := "/V1/cmsBlock/search?" + search.Encode()
	var result SearchResult[CMSBlock]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list CMS blocks: %w", err)
	}
	return &result, nil
}

// GetCMSBlock retrieves a single CMS block by ID.
func (c *Client) GetCMSBlock(ctx context.Context, id int) (*CMSBlock, error) {
	path := fmt.Sprintf("/V1/cmsBlock/%d", id)
	var block CMSBlock
	if err := c.get(ctx, path, &block); err != nil {
		return nil, fmt.Errorf("get CMS block %d: %w", id, err)
	}
	return &block, nil
}
