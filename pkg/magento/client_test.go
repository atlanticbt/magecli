package magento

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates a Client pointing at the given httptest.Server.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	c, err := New(ClientOptions{
		BaseURL:   server.URL,
		Token:     "test-token",
		StoreCode: "default",
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNew_RequiresBaseURL(t *testing.T) {
	_, err := New(ClientOptions{})
	if err == nil {
		t.Error("expected error for empty BaseURL")
	}
}

func TestNew_DefaultStoreCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	c, err := New(ClientOptions{BaseURL: srv.URL, Token: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	if c.storeCode != "default" {
		t.Errorf("storeCode = %q, want default", c.storeCode)
	}
}

func TestNew_CustomStoreCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	c, err := New(ClientOptions{BaseURL: srv.URL, Token: "tok", StoreCode: "french"})
	if err != nil {
		t.Fatal(err)
	}
	if c.storeCode != "french" {
		t.Errorf("storeCode = %q, want french", c.storeCode)
	}
}

func TestClient_BearerToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_ = c.get(context.Background(), "/V1/store/storeViews", nil)

	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q, want 'Bearer test-token'", gotAuth)
	}
}

func TestClient_RequestPath(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_ = c.get(context.Background(), "/V1/products", nil)

	want := "/rest/default/V1/products"
	if gotPath != want {
		t.Errorf("request path = %q, want %q", gotPath, want)
	}
}

func TestListProducts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SearchResult[Product]{
			Items:      []Product{{ID: 1, SKU: "TEST-1", Name: "Test Product"}},
			TotalCount: 1,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	result, err := c.ListProducts(context.Background(), NewSearch())
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", result.TotalCount)
	}
	if result.Items[0].SKU != "TEST-1" {
		t.Errorf("SKU = %q, want TEST-1", result.Items[0].SKU)
	}
}

func TestGetProduct(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Product{ID: 42, SKU: "SHOE-1", Name: "Running Shoe", Price: 99.99})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	p, err := c.GetProduct(context.Background(), "SHOE-1")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Running Shoe" {
		t.Errorf("Name = %q, want Running Shoe", p.Name)
	}
	if p.Price != 99.99 {
		t.Errorf("Price = %f, want 99.99", p.Price)
	}
}

func TestGetProductMedia(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]MediaEntry{
			{ID: 1, MediaType: "image", File: "/m/y/image.jpg"},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	media, err := c.GetProductMedia(context.Background(), "SKU-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(media) != 1 {
		t.Fatalf("got %d entries, want 1", len(media))
	}
	if media[0].File != "/m/y/image.jpg" {
		t.Errorf("File = %q", media[0].File)
	}
}

func TestGetCategoryTree(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Category{
			ID: 1, Name: "Root", ChildrenData: []Category{
				{ID: 2, Name: "Clothing"},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	tree, err := c.GetCategoryTree(context.Background(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if tree.Name != "Root" {
		t.Errorf("Name = %q, want Root", tree.Name)
	}
	if len(tree.ChildrenData) != 1 {
		t.Fatalf("children = %d, want 1", len(tree.ChildrenData))
	}
}

func TestGetCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Category{ID: 5, Name: "Shoes", IsActive: true})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	cat, err := c.GetCategory(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Name != "Shoes" || !cat.IsActive {
		t.Errorf("got %+v", cat)
	}
}

func TestGetCategoryProducts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]CategoryProduct{
			{SKU: "P1", Position: 1, CategoryID: "5"},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	products, err := c.GetCategoryProducts(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 1 || products[0].SKU != "P1" {
		t.Errorf("unexpected products: %+v", products)
	}
}

func TestGetAttribute(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Attribute{AttributeCode: "color", FrontendLabel: "Color"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	attr, err := c.GetAttribute(context.Background(), "color")
	if err != nil {
		t.Fatal(err)
	}
	if attr.FrontendLabel != "Color" {
		t.Errorf("FrontendLabel = %q, want Color", attr.FrontendLabel)
	}
}

func TestGetAttributeOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]AttributeOption{
			{Label: "Red", Value: "10"},
			{Label: "Blue", Value: "11"},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	opts, err := c.GetAttributeOptions(context.Background(), "color")
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) != 2 {
		t.Fatalf("got %d options, want 2", len(opts))
	}
}

func TestListAttributeSets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SearchResult[AttributeSet]{
			Items:      []AttributeSet{{AttributeSetID: 4, AttributeSetName: "Default"}},
			TotalCount: 1,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	result, err := c.ListAttributeSets(context.Background(), NewSearch())
	if err != nil {
		t.Fatal(err)
	}
	if result.Items[0].AttributeSetName != "Default" {
		t.Errorf("unexpected set: %+v", result.Items[0])
	}
}

func TestGetStockStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(StockItem{ItemID: 1, Qty: 42, IsInStock: true})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	item, err := c.GetStockStatus(context.Background(), "SKU-1")
	if err != nil {
		t.Fatal(err)
	}
	if item.Qty != 42 || !item.IsInStock {
		t.Errorf("unexpected stock: %+v", item)
	}
}

func TestGetConfigurableChildren(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Product{
			{SKU: "CHILD-1", Name: "Small"},
			{SKU: "CHILD-2", Name: "Large"},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	children, err := c.GetConfigurableChildren(context.Background(), "PARENT")
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 2 {
		t.Fatalf("got %d children, want 2", len(children))
	}
}

func TestGetConfigurableOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]ConfigurableOption{
			{ID: 1, Label: "Size", Values: []ConfigurableValue{{ValueIndex: 10}}},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	opts, err := c.GetConfigurableOptions(context.Background(), "PARENT")
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) != 1 || opts[0].Label != "Size" {
		t.Errorf("unexpected options: %+v", opts)
	}
}

func TestListCMSPages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SearchResult[CMSPage]{
			Items:      []CMSPage{{ID: 1, Identifier: "home", Title: "Home Page", Active: true}},
			TotalCount: 1,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	result, err := c.ListCMSPages(context.Background(), NewSearch())
	if err != nil {
		t.Fatal(err)
	}
	if result.Items[0].Identifier != "home" {
		t.Errorf("unexpected page: %+v", result.Items[0])
	}
}

func TestGetCMSPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CMSPage{ID: 11, Identifier: "home", Title: "Home Page", Content: "<h1>Hi</h1>"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	page, err := c.GetCMSPage(context.Background(), 11)
	if err != nil {
		t.Fatal(err)
	}
	if page.Title != "Home Page" || page.Content != "<h1>Hi</h1>" {
		t.Errorf("unexpected page: %+v", page)
	}
}

func TestListCMSBlocks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SearchResult[CMSBlock]{
			Items:      []CMSBlock{{ID: 1, Identifier: "footer", Title: "Footer Block"}},
			TotalCount: 1,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	result, err := c.ListCMSBlocks(context.Background(), NewSearch())
	if err != nil {
		t.Fatal(err)
	}
	if result.Items[0].Identifier != "footer" {
		t.Errorf("unexpected block: %+v", result.Items[0])
	}
}

func TestGetCMSBlock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CMSBlock{ID: 31, Identifier: "contact-us", Title: "Contact"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	block, err := c.GetCMSBlock(context.Background(), 31)
	if err != nil {
		t.Fatal(err)
	}
	if block.Title != "Contact" {
		t.Errorf("unexpected block: %+v", block)
	}
}

func TestListStoreViews(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]StoreView{
			{ID: 1, Code: "default", Name: "Default Store View", IsActive: 1},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	views, err := c.ListStoreViews(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 || views[0].Code != "default" {
		t.Errorf("unexpected views: %+v", views)
	}
}

func TestListStoreConfigs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]StoreConfig{
			{ID: 1, Code: "default", Locale: "en_US", BaseCurrencyCode: "USD"},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	configs, err := c.ListStoreConfigs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if configs[0].Locale != "en_US" {
		t.Errorf("unexpected config: %+v", configs[0])
	}
}

func TestListStoreGroups(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]StoreGroup{{ID: 1, Name: "Main", Code: "main"}})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	groups, err := c.ListStoreGroups(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Errorf("got %d groups, want 1", len(groups))
	}
}

func TestListWebsites(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]Website{{ID: 1, Code: "base", Name: "Main Website"}})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	websites, err := c.ListWebsites(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if websites[0].Name != "Main Website" {
		t.Errorf("unexpected website: %+v", websites[0])
	}
}

func TestClient_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "The consumer isn't authorized to access %resources.", "parameters": {"resources": "Magento_Catalog::products"}}`))
	}))
	defer srv.Close()

	c, _ := New(ClientOptions{BaseURL: srv.URL, Token: "bad-token", StoreCode: "default"})
	_, err := c.ListProducts(context.Background(), NewSearch())
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestClient_404Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Requested product doesn't exist"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetProduct(context.Background(), "NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}
