package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client wraps HTTP access with Magento-aware defaults (Bearer token auth).
type Client struct {
	baseURL   *url.URL
	token     string
	userAgent string

	httpClient *http.Client

	enableCache bool
	cacheMu     sync.RWMutex
	cache       map[string]*cacheEntry

	retry RetryPolicy
	debug bool
}

type Options struct {
	BaseURL   string
	Token     string
	UserAgent string
	Timeout   time.Duration

	EnableCache bool
	Retry       RetryPolicy
	Debug       bool
}

type RetryPolicy struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

type cacheEntry struct {
	etag     string
	body     []byte
	storedAt time.Time
}

func New(opts Options) (*Client, error) {
	if opts.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	base, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	if base.Scheme == "" {
		return nil, fmt.Errorf("base URL must include scheme (e.g. https)")
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &Client{
		baseURL: base,
		token:   strings.TrimSpace(opts.Token),
		userAgent: func() string {
			if opts.UserAgent != "" {
				return opts.UserAgent
			}
			return "magecli"
		}(),
		httpClient:  &http.Client{Timeout: timeout},
		enableCache: opts.EnableCache,
		cache:       make(map[string]*cacheEntry),
	}

	if opts.Debug || os.Getenv("MAGECLI_HTTP_DEBUG") != "" {
		client.debug = true
	}

	policy := opts.Retry
	if policy.MaxAttempts == 0 {
		policy.MaxAttempts = 3
	}
	if policy.InitialBackoff == 0 {
		policy.InitialBackoff = 200 * time.Millisecond
	}
	if policy.MaxBackoff == 0 {
		policy.MaxBackoff = 2 * time.Second
	}
	client.retry = policy

	return client, nil
}

func (c *Client) NewRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("path is required")
	}

	var rel *url.URL
	var err error

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		rel, err = url.Parse(path)
		if err != nil {
			return nil, fmt.Errorf("parse request URL: %w", err)
		}
	} else {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		rel, err = url.Parse(path)
		if err != nil {
			return nil, fmt.Errorf("parse request path: %w", err)
		}
	}

	if rel.Path == "" {
		rel.Path = "/"
	}

	u := *c.baseURL
	basePath := c.baseURL.Path
	if strings.HasPrefix(path, "/") && basePath != "" {
		// Preserve the escaped form (RawPath) so that percent-encoded segments
		// — e.g. a SKU containing "/" encoded as %2F — survive instead of being
		// decoded back into path separators.
		if strings.HasPrefix(rel.Path, basePath) {
			u.Path = rel.Path
			u.RawPath = rel.EscapedPath()
		} else {
			u.Path = strings.TrimSuffix(basePath, "/") + rel.Path
			u.RawPath = strings.TrimSuffix(c.baseURL.EscapedPath(), "/") + rel.EscapedPath()
		}
	} else {
		resolved := c.baseURL.ResolveReference(rel)
		u = *resolved
	}
	u.RawQuery = rel.RawQuery

	var payload []byte
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
	}

	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(payload))
		data := payload
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(data)), nil
		}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	// Bearer token authentication for Magento 2. Pin the credential to the
	// configured host so it is never attached to an unrelated absolute URL
	// (e.g. `api https://attacker.example/...`), which would leak the token.
	if c.token != "" && sameHost(&u, c.baseURL) {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, nil
}

// sameHost reports whether two URLs target the same scheme, host, and port,
// treating an absent port as the scheme's default so that an explicit :443
// (or :80) still matches a base URL without one.
func sameHost(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) &&
		strings.EqualFold(a.Hostname(), b.Hostname()) &&
		effectivePort(a) == effectivePort(b)
}

func effectivePort(u *url.URL) string {
	if p := u.Port(); p != "" {
		return p
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return "443"
	case "http":
		return "80"
	}
	return ""
}

func (c *Client) Do(req *http.Request, v any) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	attempts := 0
	for {
		attemptReq, err := cloneRequest(req)
		if err != nil {
			return err
		}

		if c.enableCache && attemptReq.Method == http.MethodGet {
			if etag := c.cachedETag(attemptReq); etag != "" {
				attemptReq.Header.Set("If-None-Match", etag)
			}
		}

		if c.debug {
			fmt.Fprintf(os.Stderr, "--> %s %s\n", attemptReq.Method, attemptReq.URL.String())
		}

		resp, err := c.httpClient.Do(attemptReq)
		if err != nil {
			if !isIdempotent(req.Method) || !c.shouldRetry(attempts) {
				if c.debug {
					fmt.Fprintf(os.Stderr, "<-- network error: %v\n", err)
				}
				return err
			}
			attempts++
			continueRetry, waitErr := c.backoff(req.Context(), attempts, resp)
			if waitErr != nil {
				return waitErr
			}
			if !continueRetry {
				return err
			}
			continue
		}

		if c.debug {
			fmt.Fprintf(os.Stderr, "<-- %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
		}

		if resp.StatusCode == http.StatusNotModified && c.enableCache && attemptReq.Method == http.MethodGet {
			_ = resp.Body.Close()
			return c.applyCachedResponse(attemptReq, v)
		}

		if shouldRetryStatus(resp.StatusCode) && isIdempotent(req.Method) {
			bodyBytes, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if !c.shouldRetry(attempts) {
				if len(bodyBytes) > 0 {
					resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				}
				return decodeError(resp)
			}
			attempts++
			continueRetry, waitErr := c.backoff(req.Context(), attempts, resp)
			if waitErr != nil {
				return waitErr
			}
			if !continueRetry {
				if len(bodyBytes) > 0 {
					resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				}
				return decodeError(resp)
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			defer func() { _ = resp.Body.Close() }()
			return decodeError(resp)
		}

		if v == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if c.enableCache && attemptReq.Method == http.MethodGet {
				c.storeCache(attemptReq, nil, resp.Header.Get("ETag"))
			}
			return nil
		}

		if writer, ok := v.(io.Writer); ok {
			_, err := io.Copy(writer, resp.Body)
			_ = resp.Body.Close()
			return err
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return err
		}

		if c.enableCache && attemptReq.Method == http.MethodGet && resp.Header.Get("ETag") != "" {
			c.storeCache(attemptReq, bodyBytes, resp.Header.Get("ETag"))
		}

		if len(bodyBytes) == 0 {
			return nil
		}

		return json.Unmarshal(bodyBytes, v)
	}
}

// HTTPError represents a non-2xx HTTP response. It carries the status code so
// that the top-level command runner can map failures to distinct exit codes
// (e.g. 404 vs 401) and emit targeted remediation hints.
type HTTPError struct {
	StatusCode int
	Status     string
	Message    string
}

func (e *HTTPError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Status, e.Message)
	}
	return e.Status
}

func decodeError(resp *http.Response) error {
	herr := &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status}

	data, err := io.ReadAll(resp.Body)
	if err != nil || len(data) == 0 {
		return herr
	}

	// Magento 2 error format: {"message": "...", "parameters": [...] | {...}}
	var magentoErr struct {
		Message    string          `json:"message"`
		Parameters json.RawMessage `json:"parameters"`
	}
	if json.Unmarshal(data, &magentoErr) == nil && magentoErr.Message != "" {
		herr.Message = substituteParams(magentoErr.Message, magentoErr.Parameters)
		return herr
	}

	herr.Message = strings.TrimSpace(string(data))
	return herr
}

// substituteParams resolves Magento error placeholders. Magento returns
// parameters either positionally (["a","b"] → %1, %2) or as a named map
// ({"resources":"x"} → %resources); handle both.
func substituteParams(msg string, raw json.RawMessage) string {
	if len(raw) == 0 {
		return msg
	}

	// Named parameters: {"key": "value"}. Replace longer keys first so a key
	// that is a prefix of another ("%field" vs "%fieldName") cannot clobber it;
	// map iteration order would make the corruption nondeterministic.
	var named map[string]any
	if json.Unmarshal(raw, &named) == nil && len(named) > 0 {
		keys := make([]string, 0, len(named))
		for key := range named {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
		for _, key := range keys {
			msg = strings.ReplaceAll(msg, "%"+key, fmt.Sprintf("%v", named[key]))
		}
		return msg
	}

	// Positional parameters: ["a", "b"]. Replace from the highest index down so
	// that "%1" does not clobber the "%1" prefix of "%10".
	var positional []any
	if json.Unmarshal(raw, &positional) == nil {
		for i := len(positional) - 1; i >= 0; i-- {
			placeholder := fmt.Sprintf("%%%d", i+1)
			msg = strings.ReplaceAll(msg, placeholder, fmt.Sprintf("%v", positional[i]))
		}
	}
	return msg
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	newReq := req.Clone(req.Context())
	newReq.Header = req.Header.Clone()
	if req.Body != nil {
		if req.GetBody == nil {
			return nil, fmt.Errorf("request body cannot be replayed")
		}
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		newReq.Body = body
	}
	return newReq, nil
}

func shouldRetryStatus(code int) bool {
	return code == http.StatusTooManyRequests || (code >= 500 && code <= 599)
}

// isIdempotent reports whether a request method is safe to retry automatically.
// Only GET/HEAD are retried so that a non-idempotent write (POST/PUT/DELETE via
// the `api` command) is never silently replayed after a 5xx or network error.
func isIdempotent(method string) bool {
	m := strings.ToUpper(method)
	return m == http.MethodGet || m == http.MethodHead
}

func (c *Client) shouldRetry(attempts int) bool {
	return attempts+1 < c.retry.MaxAttempts
}

func (c *Client) backoff(ctx context.Context, attempts int, resp *http.Response) (bool, error) {
	if attempts >= c.retry.MaxAttempts {
		return false, nil
	}
	delay := c.retry.InitialBackoff
	if attempts > 1 {
		delay *= time.Duration(1 << (attempts - 1))
	}
	if delay > c.retry.MaxBackoff {
		delay = c.retry.MaxBackoff
	}
	if resp != nil {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if secs, err := strconv.Atoi(retryAfter); err == nil {
				delay = time.Duration(secs) * time.Second
			}
		}
	}
	if delay <= 0 {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			return true, nil
		}
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-timer.C:
		return true, nil
	}
}

func (c *Client) cacheKey(req *http.Request) string {
	return req.Method + " " + req.URL.String()
}

func (c *Client) cachedETag(req *http.Request) string {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()
	if entry, ok := c.cache[c.cacheKey(req)]; ok {
		return entry.etag
	}
	return ""
}

func (c *Client) storeCache(req *http.Request, body []byte, etag string) {
	if etag == "" || len(body) == 0 {
		return
	}
	c.cacheMu.Lock()
	c.cache[c.cacheKey(req)] = &cacheEntry{etag: etag, body: append([]byte(nil), body...), storedAt: time.Now()}
	c.cacheMu.Unlock()
}

func (c *Client) applyCachedResponse(req *http.Request, v any) error {
	if v == nil {
		return nil
	}
	c.cacheMu.RLock()
	entry, ok := c.cache[c.cacheKey(req)]
	c.cacheMu.RUnlock()
	if !ok {
		return fmt.Errorf("cached response missing for %s", req.URL)
	}
	if writer, ok := v.(io.Writer); ok {
		_, err := writer.Write(entry.body)
		return err
	}
	if len(entry.body) == 0 {
		return nil
	}
	return json.Unmarshal(entry.body, v)
}
