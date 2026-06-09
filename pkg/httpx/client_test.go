package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNew_RequiresBaseURL(t *testing.T) {
	_, err := New(Options{})
	if err == nil {
		t.Error("expected error for empty BaseURL")
	}
}

func TestNew_RequiresScheme(t *testing.T) {
	_, err := New(Options{BaseURL: "example.com"})
	if err == nil {
		t.Error("expected error for missing scheme")
	}
}

func TestNew_ValidOptions(t *testing.T) {
	c, err := New(Options{BaseURL: "https://example.com", Token: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	if c.token != "tok" {
		t.Errorf("token = %q, want tok", c.token)
	}
	if c.userAgent != "magecli" {
		t.Errorf("userAgent = %q, want magecli", c.userAgent)
	}
}

func TestNew_CustomUserAgent(t *testing.T) {
	c, err := New(Options{BaseURL: "https://example.com", UserAgent: "custom/1.0"})
	if err != nil {
		t.Fatal(err)
	}
	if c.userAgent != "custom/1.0" {
		t.Errorf("userAgent = %q, want custom/1.0", c.userAgent)
	}
}

func TestNewRequest_BearerAuth(t *testing.T) {
	c, _ := New(Options{BaseURL: "https://example.com", Token: "my-token"})
	req, err := c.NewRequest(context.Background(), "GET", "/V1/products", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer my-token" {
		t.Errorf("Authorization = %q, want 'Bearer my-token'", got)
	}
}

func TestNewRequest_NoAuth_WhenTokenEmpty(t *testing.T) {
	c, _ := New(Options{BaseURL: "https://example.com"})
	req, _ := c.NewRequest(context.Background(), "GET", "/V1/products", nil)
	if got := req.Header.Get("Authorization"); got != "" {
		t.Errorf("Authorization should be empty, got %q", got)
	}
}

func TestNewRequest_AcceptJSON(t *testing.T) {
	c, _ := New(Options{BaseURL: "https://example.com"})
	req, _ := c.NewRequest(context.Background(), "GET", "/V1/test", nil)
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q, want application/json", got)
	}
}

func TestNewRequest_WithBody(t *testing.T) {
	c, _ := New(Options{BaseURL: "https://example.com"})
	body := map[string]string{"name": "test"}
	req, err := c.NewRequest(context.Background(), "POST", "/V1/products", body)
	if err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type not set for body")
	}
	if req.ContentLength <= 0 {
		t.Error("ContentLength not set")
	}
}

func TestNewRequest_EmptyPath(t *testing.T) {
	c, _ := New(Options{BaseURL: "https://example.com"})
	_, err := c.NewRequest(context.Background(), "GET", "", nil)
	if err == nil {
		t.Error("expected error for empty path")
	}
}

func TestNewRequest_PathResolution(t *testing.T) {
	c, _ := New(Options{BaseURL: "https://example.com/rest/default"})
	req, err := c.NewRequest(context.Background(), "GET", "/V1/products", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(req.URL.Path, "/rest/default/V1/products") {
		t.Errorf("URL path = %q, expected /rest/default/V1/products", req.URL.Path)
	}
}

func TestNewRequest_PreservesEscapedSlashInPath(t *testing.T) {
	// A SKU containing "/" is path-escaped as %2F and must survive instead of
	// being decoded back into a path separator (which would hit a different URL).
	c, _ := New(Options{BaseURL: "https://example.com/rest/default"})
	req, err := c.NewRequest(context.Background(), "GET", "/V1/products/ABC%2FDEF", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(req.URL.String(), "ABC%2FDEF") {
		t.Errorf("escaped slash lost; URL = %q", req.URL.String())
	}
	if strings.Contains(req.URL.String(), "ABC/DEF") {
		t.Errorf("encoded slash was decoded to a separator; URL = %q", req.URL.String())
	}
}

func TestNewRequest_TokenNotSentToForeignHost(t *testing.T) {
	c, _ := New(Options{BaseURL: "https://store.example.com/rest/default", Token: "secret"})

	same, _ := c.NewRequest(context.Background(), "GET", "/V1/products", nil)
	if same.Header.Get("Authorization") == "" {
		t.Error("token should be attached for the configured host")
	}

	foreign, _ := c.NewRequest(context.Background(), "GET", "https://evil.example.org/steal", nil)
	if got := foreign.Header.Get("Authorization"); got != "" {
		t.Errorf("token leaked to foreign host: %q", got)
	}
}

func TestNewRequest_TokenSentWithExplicitDefaultPort(t *testing.T) {
	// An explicit :443 on an https URL is the same host as one without a port;
	// the token must still be attached.
	c, _ := New(Options{BaseURL: "https://store.example.com/rest/default", Token: "secret"})

	req, _ := c.NewRequest(context.Background(), "GET", "https://store.example.com:443/rest/default/V1/products", nil)
	if req.Header.Get("Authorization") == "" {
		t.Error("token should be attached when only the explicit default port differs")
	}

	other, _ := c.NewRequest(context.Background(), "GET", "https://store.example.com:8443/rest/default/V1/products", nil)
	if got := other.Header.Get("Authorization"); got != "" {
		t.Errorf("token leaked to non-default port: %q", got)
	}
}

func TestDecodeError_NamedParameters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Consumer is not authorized to access %resources", "parameters": {"resources": "Magento_Catalog::products"}}`))
	}))
	defer srv.Close()

	c, _ := New(Options{BaseURL: srv.URL, Retry: RetryPolicy{MaxAttempts: 1}})
	req, _ := c.NewRequest(context.Background(), "GET", "/test", nil)
	err := c.Do(req, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Magento_Catalog::products") {
		t.Errorf("named parameter not substituted, got: %v", err)
	}
	var herr *HTTPError
	if !errors.As(err, &herr) || herr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected *HTTPError with status 401, got %T %v", err, err)
	}
}

func TestSubstituteParams_PrefixKeysDoNotClobber(t *testing.T) {
	got := substituteParams(
		"%fieldName (%field) is invalid",
		[]byte(`{"field": "sku", "fieldName": "SKU"}`),
	)
	if got != "SKU (sku) is invalid" {
		t.Errorf("substituteParams = %q, want %q", got, "SKU (sku) is invalid")
	}
}

func TestDo_NoRetryOnNonIdempotentMethod(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c, _ := New(Options{BaseURL: srv.URL, Retry: RetryPolicy{MaxAttempts: 4}})
	req, _ := c.NewRequest(context.Background(), "POST", "/test", map[string]string{"a": "b"})
	_ = c.Do(req, nil)
	if attempts != 1 {
		t.Errorf("POST should not be retried; got %d attempts", attempts)
	}
}

func TestDo_DecodesJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"key": "value"})
	}))
	defer srv.Close()

	c, _ := New(Options{BaseURL: srv.URL})
	req, _ := c.NewRequest(context.Background(), "GET", "/test", nil)

	var result map[string]string
	err := c.Do(req, &result)
	if err != nil {
		t.Fatal(err)
	}
	if result["key"] != "value" {
		t.Errorf("result = %v", result)
	}
}

func TestDo_NilRequest(t *testing.T) {
	c, _ := New(Options{BaseURL: "https://example.com"})
	err := c.Do(nil, nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
}

func TestDo_HTTPError_MagentoFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Field %1 is required.", "parameters": ["sku"]}`))
	}))
	defer srv.Close()

	c, _ := New(Options{BaseURL: srv.URL, Retry: RetryPolicy{MaxAttempts: 1}})
	req, _ := c.NewRequest(context.Background(), "GET", "/test", nil)
	err := c.Do(req, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "sku") {
		t.Errorf("error should contain param substitution, got: %v", err)
	}
}

func TestDo_RetryOn500(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "server error"}`))
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer srv.Close()

	c, _ := New(Options{
		BaseURL: srv.URL,
		Retry:   RetryPolicy{MaxAttempts: 3, InitialBackoff: time.Millisecond, MaxBackoff: 10 * time.Millisecond},
	})
	req, _ := c.NewRequest(context.Background(), "GET", "/test", nil)

	var result map[string]string
	err := c.Do(req, &result)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestDo_RetryOn429(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, _ := New(Options{
		BaseURL: srv.URL,
		Retry:   RetryPolicy{MaxAttempts: 2, InitialBackoff: time.Millisecond, MaxBackoff: 10 * time.Millisecond},
	})
	req, _ := c.NewRequest(context.Background(), "GET", "/test", nil)
	err := c.Do(req, nil)
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestDo_ETagCaching(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("If-None-Match") == `"abc123"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"abc123"`)
		json.NewEncoder(w).Encode(map[string]string{"data": "cached"})
	}))
	defer srv.Close()

	c, _ := New(Options{BaseURL: srv.URL, EnableCache: true, Retry: RetryPolicy{MaxAttempts: 1}})

	// First request — cache miss
	req1, _ := c.NewRequest(context.Background(), "GET", "/test", nil)
	var r1 map[string]string
	if err := c.Do(req1, &r1); err != nil {
		t.Fatal(err)
	}
	if r1["data"] != "cached" {
		t.Errorf("first call data = %q", r1["data"])
	}

	// Second request — should get 304 and use cache
	req2, _ := c.NewRequest(context.Background(), "GET", "/test", nil)
	var r2 map[string]string
	if err := c.Do(req2, &r2); err != nil {
		t.Fatal(err)
	}
	if r2["data"] != "cached" {
		t.Errorf("second call data = %q", r2["data"])
	}
	if calls != 2 {
		t.Errorf("server calls = %d, want 2", calls)
	}
}

func TestShouldRetryStatus(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{404, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
	}
	for _, tt := range tests {
		if got := shouldRetryStatus(tt.code); got != tt.want {
			t.Errorf("shouldRetryStatus(%d) = %v, want %v", tt.code, got, tt.want)
		}
	}
}
