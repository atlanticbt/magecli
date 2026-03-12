package magento

import (
	"context"
	"fmt"
	"net/url"

	"github.com/atlanticbt/magecli/pkg/httpx"
)

// Client provides access to the Magento 2 REST API.
type Client struct {
	http      *httpx.Client
	storeCode string
}

// ClientOptions configures a Magento client.
type ClientOptions struct {
	BaseURL   string
	Token     string
	StoreCode string

	EnableCache bool
	Retry       httpx.RetryPolicy
	Debug       bool
}

// New creates a new Magento REST API client.
func New(opts ClientOptions) (*Client, error) {
	baseURL := opts.BaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	// Ensure base URL ends with /rest
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	// Build the REST base path: /rest/{storeCode} or /rest/default
	storeCode := opts.StoreCode
	if storeCode == "" {
		storeCode = "default"
	}

	u.Path = fmt.Sprintf("/rest/%s", storeCode)

	httpClient, err := httpx.New(httpx.Options{
		BaseURL:     u.String(),
		Token:       opts.Token,
		EnableCache: opts.EnableCache,
		Retry:       opts.Retry,
		Debug:       opts.Debug,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		http:      httpClient,
		storeCode: storeCode,
	}, nil
}

// HTTP returns the underlying HTTP client.
func (c *Client) HTTP() *httpx.Client {
	return c.http
}

// get performs a GET request and decodes the response.
func (c *Client) get(ctx context.Context, path string, v any) error {
	req, err := c.http.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	return c.http.Do(req, v)
}
