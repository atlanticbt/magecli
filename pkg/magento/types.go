package magento

// SearchResult is the standard Magento 2 paginated response.
type SearchResult[T any] struct {
	Items      []T `json:"items"`
	TotalCount int `json:"total_count"`
	SearchCriteria struct {
		FilterGroups []struct {
			Filters []struct {
				Field         string `json:"field"`
				Value         string `json:"value"`
				ConditionType string `json:"condition_type"`
			} `json:"filters"`
		} `json:"filter_groups"`
		PageSize    int `json:"page_size"`
		CurrentPage int `json:"current_page"`
	} `json:"search_criteria"`
}

// Product represents a Magento 2 catalog product.
type Product struct {
	ID               int                    `json:"id"`
	SKU              string                 `json:"sku"`
	Name             string                 `json:"name"`
	Price            float64                `json:"price"`
	Status           int                    `json:"status"`
	Visibility       int                    `json:"visibility"`
	TypeID           string                 `json:"type_id"`
	Weight           float64                `json:"weight"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	AttributeSetID   int                    `json:"attribute_set_id"`
	CustomAttributes []CustomAttribute      `json:"custom_attributes,omitempty"`
	ExtensionAttrs   map[string]any         `json:"extension_attributes,omitempty"`
	MediaGallery     []MediaEntry           `json:"media_gallery_entries,omitempty"`
	ProductLinks     []ProductLink          `json:"product_links,omitempty"`
	Options          []ProductOption        `json:"options,omitempty"`
	TierPrices       []TierPrice            `json:"tier_prices,omitempty"`
}

type CustomAttribute struct {
	AttributeCode string `json:"attribute_code"`
	Value         any    `json:"value"`
}

type MediaEntry struct {
	ID        int      `json:"id"`
	MediaType string   `json:"media_type"`
	Label     string   `json:"label"`
	Position  int      `json:"position"`
	Disabled  bool     `json:"disabled"`
	Types     []string `json:"types"`
	File      string   `json:"file"`
}

type ProductLink struct {
	SKU                 string `json:"sku"`
	LinkType            string `json:"link_type"`
	LinkedProductSKU    string `json:"linked_product_sku"`
	LinkedProductType   string `json:"linked_product_type"`
	Position            int    `json:"position"`
}

type ProductOption struct {
	ProductSKU string         `json:"product_sku"`
	OptionID   int            `json:"option_id"`
	Title      string         `json:"title"`
	Type       string         `json:"type"`
	SortOrder  int            `json:"sort_order"`
	IsRequired bool           `json:"is_require"`
	Values     []OptionValue  `json:"values,omitempty"`
}

type OptionValue struct {
	Title    string  `json:"title"`
	SortOrder int    `json:"sort_order"`
	Price    float64 `json:"price"`
	PriceType string `json:"price_type"`
	ValueID  int     `json:"option_type_id"`
}

type TierPrice struct {
	CustomerGroupID int     `json:"customer_group_id"`
	Qty             float64 `json:"qty"`
	Value           float64 `json:"value"`
}

// ConfigurableOption represents a configurable product attribute option.
type ConfigurableOption struct {
	ID             int               `json:"id"`
	AttributeID    string            `json:"attribute_id"`
	Label          string            `json:"label"`
	Position       int               `json:"position"`
	Values         []ConfigurableValue `json:"values"`
	ProductID      int               `json:"product_id"`
}

type ConfigurableValue struct {
	ValueIndex int `json:"value_index"`
}

// Category represents a Magento 2 category.
type Category struct {
	ID              int        `json:"id"`
	ParentID        int        `json:"parent_id"`
	Name            string     `json:"name"`
	IsActive        bool       `json:"is_active"`
	Position        int        `json:"position"`
	Level           int        `json:"level"`
	ProductCount    int        `json:"product_count"`
	ChildrenData    []Category `json:"children_data,omitempty"`
	CustomAttributes []CustomAttribute `json:"custom_attributes,omitempty"`
}

// Attribute represents a Magento 2 EAV attribute.
type Attribute struct {
	AttributeID    int               `json:"attribute_id"`
	AttributeCode  string            `json:"attribute_code"`
	FrontendInput  string            `json:"frontend_input"`
	FrontendLabel  string            `json:"default_frontend_label"`
	IsRequired     bool              `json:"is_required"`
	IsUserDefined  bool              `json:"is_user_defined"`
	Options        []AttributeOption `json:"options,omitempty"`
	EntityTypeID   string            `json:"entity_type_id"`
}

type AttributeOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// AttributeSet represents a Magento 2 attribute set.
type AttributeSet struct {
	AttributeSetID   int    `json:"attribute_set_id"`
	AttributeSetName string `json:"attribute_set_name"`
	SortOrder        int    `json:"sort_order"`
	EntityTypeID     int    `json:"entity_type_id"`
}

// StockItem represents Magento 2 inventory data.
type StockItem struct {
	ItemID       int     `json:"item_id"`
	ProductID    int     `json:"product_id"`
	StockID      int     `json:"stock_id"`
	Qty          float64 `json:"qty"`
	IsInStock    bool    `json:"is_in_stock"`
	IsQtyDecimal bool    `json:"is_qty_decimal"`
	MinQty       float64 `json:"min_qty"`
	MinSaleQty   float64 `json:"min_sale_qty"`
	MaxSaleQty   float64 `json:"max_sale_qty"`
}

// InventorySourceItem represents MSI (Multi-Source Inventory) source items.
type InventorySourceItem struct {
	SKU        string  `json:"sku"`
	SourceCode string  `json:"source_code"`
	Quantity   float64 `json:"quantity"`
	Status     int     `json:"status"`
}

// StoreView represents a Magento 2 store view.
type StoreView struct {
	ID           int    `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	WebsiteID    int    `json:"website_id"`
	StoreGroupID int    `json:"store_group_id"`
	IsActive     int    `json:"is_active"`
}

// StoreConfig represents Magento 2 store configuration.
type StoreConfig struct {
	ID                    int    `json:"id"`
	Code                  string `json:"code"`
	WebsiteID             int    `json:"website_id"`
	Locale                string `json:"locale"`
	BaseCurrencyCode      string `json:"base_currency_code"`
	DefaultDisplayCurrency string `json:"default_display_currency_code"`
	Timezone              string `json:"timezone"`
	WeightUnit            string `json:"weight_unit"`
	BaseURL               string `json:"base_url"`
	BaseLinkURL           string `json:"base_link_url"`
	BaseStaticURL         string `json:"base_static_url"`
	BaseMediaURL          string `json:"base_media_url"`
	SecureBaseURL         string `json:"secure_base_url"`
	SecureBaseLinkURL     string `json:"secure_base_link_url"`
	SecureBaseStaticURL   string `json:"secure_base_static_url"`
	SecureBaseMediaURL    string `json:"secure_base_media_url"`
}

// StoreGroup represents a Magento 2 store group.
type StoreGroup struct {
	ID              int    `json:"id"`
	WebsiteID       int    `json:"website_id"`
	Name            string `json:"name"`
	RootCategoryID  int    `json:"root_category_id"`
	DefaultStoreID  int    `json:"default_store_id"`
	Code            string `json:"code"`
}

// Website represents a Magento 2 website.
type Website struct {
	ID              int    `json:"id"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	DefaultGroupID  int    `json:"default_group_id"`
}

// CatalogRule represents a Magento 2 catalog price rule.
type CatalogRule struct {
	RuleID             int    `json:"rule_id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	IsActive           bool   `json:"is_active"`
	StopRulesProcessing bool  `json:"stop_rules_processing"`
	SortOrder          int    `json:"sort_order"`
	SimpleAction       string `json:"simple_action"`
	DiscountAmount     float64 `json:"discount_amount"`
	FromDate           string `json:"from_date"`
	ToDate             string `json:"to_date"`
	CustomerGroupIDs   []int  `json:"customer_group_ids"`
	WebsiteIDs         []int  `json:"website_ids"`
	ExtensionAttrs     map[string]any `json:"extension_attributes,omitempty"`
}

// CartRule represents a Magento 2 cart price rule (sales rule).
type CartRule struct {
	RuleID              int            `json:"rule_id"`
	Name                string         `json:"name"`
	Description         string         `json:"description"`
	IsActive            bool           `json:"is_active"`
	StopRulesProcessing bool           `json:"stop_rules_processing"`
	SortOrder           int            `json:"sort_order"`
	SimpleAction        string         `json:"simple_action"`
	DiscountAmount      float64        `json:"discount_amount"`
	DiscountQty         float64        `json:"discount_qty"`
	DiscountStep        int            `json:"discount_step"`
	ApplyToShipping     bool           `json:"apply_to_shipping"`
	TimesUsed           int            `json:"times_used"`
	IsRSS               bool           `json:"is_rss"`
	CouponType          string         `json:"coupon_type"`
	UseAutoGeneration   bool           `json:"use_auto_generation"`
	UsesPerCoupon       int            `json:"uses_per_coupon"`
	UsesPerCustomer     int            `json:"uses_per_customer"`
	FromDate            string         `json:"from_date"`
	ToDate              string         `json:"to_date"`
	CustomerGroupIDs    []int          `json:"customer_group_ids"`
	WebsiteIDs          []int          `json:"website_ids"`
	StoreLabels         []StoreLabel   `json:"store_labels,omitempty"`
	ExtensionAttrs      map[string]any `json:"extension_attributes,omitempty"`
}

// StoreLabel represents a label for a specific store view.
type StoreLabel struct {
	StoreID int    `json:"store_id"`
	Label   string `json:"store_label"`
}

// Coupon represents a Magento 2 coupon.
type Coupon struct {
	CouponID        int    `json:"coupon_id"`
	RuleID          int    `json:"rule_id"`
	Code            string `json:"code"`
	UsageLimit      int    `json:"usage_limit"`
	UsagePerCustomer int   `json:"usage_per_customer"`
	TimesUsed       int    `json:"times_used"`
	IsPrimary       bool   `json:"is_primary"`
	Type            int    `json:"type"`
	CreatedAt       string `json:"created_at"`
	ExpirationDate  string `json:"expiration_date"`
}

// CMSPage represents a Magento 2 CMS page.
type CMSPage struct {
	ID              int    `json:"id"`
	Identifier      string `json:"identifier"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	ContentHeading  string `json:"content_heading"`
	Active          bool   `json:"active"`
	SortOrder       string `json:"sort_order"`
	PageLayout      string `json:"page_layout"`
	MetaTitle       string `json:"meta_title"`
	MetaKeywords    string `json:"meta_keywords"`
	MetaDescription string `json:"meta_description"`
	CreationTime    string `json:"creation_time"`
	UpdateTime      string `json:"update_time"`
}

// CMSBlock represents a Magento 2 CMS static block.
type CMSBlock struct {
	ID           int    `json:"id"`
	Identifier   string `json:"identifier"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Active       bool   `json:"active"`
	CreationTime string `json:"creation_time"`
	UpdateTime   string `json:"update_time"`
}
