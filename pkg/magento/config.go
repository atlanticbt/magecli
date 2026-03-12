package magento

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// ConfigEntry represents a single configuration path and its value.
type ConfigEntry struct {
	Path  string `json:"path"`
	Value string `json:"value"`
	Scope string `json:"scope"`
}

// jsonKeyToConfigPath maps Magento REST API JSON field names to their
// corresponding system configuration paths in core_config_data.
var jsonKeyToConfigPath = map[string]string{
	"locale":                        "general/locale/code",
	"base_currency_code":            "currency/options/base",
	"default_display_currency_code": "currency/options/default",
	"timezone":                      "general/locale/timezone",
	"weight_unit":                   "general/locale/weight_unit",
	"base_url":                      "web/unsecure/base_url",
	"base_link_url":                 "web/unsecure/base_link_url",
	"base_static_url":               "web/unsecure/base_static_url",
	"base_media_url":                "web/unsecure/base_media_url",
	"secure_base_url":               "web/secure/base_url",
	"secure_base_link_url":          "web/secure/base_link_url",
	"secure_base_static_url":        "web/secure/base_static_url",
	"secure_base_media_url":         "web/secure/base_media_url",
}

// metadataKeys are JSON keys that represent store metadata, not config values.
var metadataKeys = map[string]bool{
	"id":         true,
	"code":       true,
	"website_id": true,
}

// GetStoreConfigRaw retrieves store configuration as raw maps to capture all
// fields returned by the API, including any not mapped in our struct.
func (c *Client) GetStoreConfigRaw(ctx context.Context) ([]map[string]any, error) {
	var configs []map[string]any
	if err := c.get(ctx, "/V1/store/storeConfigs", &configs); err != nil {
		return nil, fmt.Errorf("get store configs: %w", err)
	}
	return configs, nil
}

// GetConfigEntries retrieves store configuration and returns it as a flat list
// of ConfigEntry values with Magento config paths. If storeCode is non-empty,
// results are filtered to that store.
func (c *Client) GetConfigEntries(ctx context.Context, storeCode string) ([]ConfigEntry, error) {
	raw, err := c.GetStoreConfigRaw(ctx)
	if err != nil {
		return nil, err
	}

	var entries []ConfigEntry
	for _, store := range raw {
		code, _ := store["code"].(string)
		if code == "" {
			code = "default"
		}
		if storeCode != "" && code != storeCode {
			continue
		}

		for key, val := range store {
			if metadataKeys[key] {
				continue
			}
			path, ok := jsonKeyToConfigPath[key]
			if !ok {
				// For unmapped keys, use the JSON key as-is under a
				// synthetic path so nothing is silently dropped.
				path = "store/" + key
			}
			entries = append(entries, ConfigEntry{
				Path:  path,
				Value: fmt.Sprintf("%v", val),
				Scope: code,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Scope != entries[j].Scope {
			return entries[i].Scope < entries[j].Scope
		}
		return entries[i].Path < entries[j].Path
	})

	return entries, nil
}

// FilterConfigEntries returns entries whose path starts with prefix or
// contains the search term (case-insensitive).
func FilterConfigEntries(entries []ConfigEntry, filter string) []ConfigEntry {
	if filter == "" {
		return entries
	}
	lower := strings.ToLower(filter)
	var out []ConfigEntry
	for _, e := range entries {
		if strings.HasPrefix(strings.ToLower(e.Path), lower) ||
			strings.Contains(strings.ToLower(e.Path), lower) {
			out = append(out, e)
		}
	}
	return out
}
