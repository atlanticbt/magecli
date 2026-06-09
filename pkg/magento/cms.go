package magento

import (
	"context"
	"fmt"
	"net/url"
)

// CMSPageMetadataFields lists every CMSPage field except the HTML content
// body, for use with Magento's fields= response filter so the body never
// leaves the server. Keep in sync with the CMSPage struct tags.
const CMSPageMetadataFields = "id,identifier,title,content_heading,active,sort_order,page_layout,meta_title,meta_keywords,meta_description,creation_time,update_time"

// CMSBlockMetadataFields is the CMSBlock equivalent of CMSPageMetadataFields.
const CMSBlockMetadataFields = "id,identifier,title,active,creation_time,update_time"

// ListCMSPages retrieves CMS pages matching search criteria.
func (c *Client) ListCMSPages(ctx context.Context, search *SearchCriteria) (*SearchResult[CMSPage], error) {
	path := "/V1/cmsPage/search?" + search.Encode()
	var result SearchResult[CMSPage]
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list CMS pages: %w", err)
	}
	return &result, nil
}

// GetCMSPage retrieves a single CMS page by ID. When fields is non-empty it is
// passed through as Magento's `fields=` response filter to shrink the payload.
func (c *Client) GetCMSPage(ctx context.Context, id int, fields string) (*CMSPage, error) {
	path := fmt.Sprintf("/V1/cmsPage/%d", id)
	if fields != "" {
		path += "?fields=" + url.QueryEscape(fields)
	}
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

// GetCMSBlock retrieves a single CMS block by ID. When fields is non-empty it
// is passed through as Magento's `fields=` response filter.
func (c *Client) GetCMSBlock(ctx context.Context, id int, fields string) (*CMSBlock, error) {
	path := fmt.Sprintf("/V1/cmsBlock/%d", id)
	if fields != "" {
		path += "?fields=" + url.QueryEscape(fields)
	}
	var block CMSBlock
	if err := c.get(ctx, path, &block); err != nil {
		return nil, fmt.Errorf("get CMS block %d: %w", id, err)
	}
	return &block, nil
}
