---
name: magecli
version: 1.0.0
description: Query Magento 2 stores via REST API - products, categories, attributes, inventory, CMS, store config
triggers:
  - magento
  - magecli
  - product catalog
  - category tree
  - inventory stock
  - magento api
  - sku lookup
  - cms page
  - cms block
  - store config
  - store views
  - magento config
  - config compare
  - environment config
  - catalog price rule
  - cart price rule
  - coupon
  - promotion
  - sales rule
  - promo
---

# magecli - Magento 2 CLI

Query Magento 2 stores via the REST API. Designed for AI agents and developers.

## Dependency Check

```bash
magecli --version
```

If not installed, build from source:
```bash
cd /path/to/magecli && make build
# Binary at bin/magecli
```

## Auth Setup

Magento 2.3.2+ requires authentication for all REST endpoints.

1. Create an Integration token in Magento Admin > System > Integrations
2. Store the token:
```bash
magecli auth login https://store.example.com --token <bearer-token>
```

Or use environment variable:
```bash
export MAGECLI_TOKEN=<bearer-token>
```

3. Create a context (read-only by default):
```bash
magecli context create production --host store.example.com --store-code default --set-active

# To allow write operations (POST/PUT/DELETE) via the api command:
magecli context create production --host store.example.com --set-active --allow-writes
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `magecli product list` | List/search products |
| `magecli product view <sku>` | View product details |
| `magecli product media <sku>` | List product images |
| `magecli product children <sku>` | List configurable variants |
| `magecli product options <sku>` | List configurable options |
| `magecli product search <term>` | Quick search products by name |
| `magecli product url <url-key>` | Find product by URL key |
| `magecli category tree` | Display category tree |
| `magecli category view <id>` | View category details |
| `magecli category products <id>` | List products in category |
| `magecli attribute view <code>` | View attribute definition |
| `magecli attribute options <code>` | List attribute option values |
| `magecli attribute sets` | List attribute sets |
| `magecli inventory status <sku>` | Check stock status |
| `magecli store views` | List store views |
| `magecli store config [code]` | Show store configuration |
| `magecli store groups` | List store groups |
| `magecli store websites` | List websites |
| `magecli config list` | List config as path=value pairs |
| `magecli config get <path>` | Get specific config value |
| `magecli config dump` | Dump all config for diffing |
| `magecli promo catalog-rule list` | List catalog price rules |
| `magecli promo catalog-rule view <id>` | View catalog price rule |
| `magecli promo cart-rule list` | List cart price rules |
| `magecli promo cart-rule view <id>` | View cart price rule |
| `magecli promo coupon list` | List coupon codes |
| `magecli promo coupon view <id>` | View coupon details |
| `magecli cms page list` | List CMS pages |
| `magecli cms page view <id>` | View CMS page details |
| `magecli cms block list` | List CMS blocks |
| `magecli cms block view <id>` | View CMS block details |
| `magecli api <path>` | Raw API escape hatch |
| `magecli update` | Self-update to latest release |

## Filtering & Sorting

```bash
# Filter products
magecli product list --filter "name like %shirt%" --filter "price gt 50"
magecli product list --filter "status eq 1" --filter "type_id eq configurable"

# Sort results
magecli product list --sort "price:ASC" --sort "name:DESC"

# Pagination
magecli product list --limit 50 --page 2
```

**Filter operators:** eq, neq, gt, gteq, lt, lteq, like, nlike, in, nin, null, notnull, from, to, finset

## Output Modes

```bash
# JSON output (for programmatic use)
magecli product list --json

# YAML output
magecli product view SKU123 --yaml

# jq filtering
magecli product list --json --jq '.items[].sku'
magecli product view SKU123 --json --jq '.name'

# Go template
magecli product list --json --template '{{range .items}}{{.sku}} {{.name}}{{"\n"}}{{end}}'
```

## Context Management

```bash
magecli context create staging --host staging.example.com --store-code default
magecli context create dev --host dev.example.com --allow-writes --set-active
magecli context use staging
magecli context list
magecli context delete staging

# Per-command override
magecli product list --context production
```

Contexts are **read-only by default**. The `api` command blocks POST/PUT/DELETE unless the context was created with `--allow-writes`. All first-class commands (product, category, etc.) use GET only and are unaffected.

## Raw API Access

```bash
# GET requests work on any context
magecli api /V1/store/storeViews --json
magecli api /V1/cmsPage/1 --json
magecli api /V1/products -P "searchCriteria[pageSize]=5" --json

# Write operations require --allow-writes on the context
magecli api /V1/products -X POST -d '{"product": {...}}' --json
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `MAGECLI_TOKEN` | Bearer token (bypasses keyring) |
| `MAGECLI_CONFIG_DIR` | Config directory override |
| `MAGECLI_HTTP_DEBUG` | Enable HTTP debug logging |
| `MAGECLI_PAGER` | Pager command override |
| `MAGECLI_ALLOW_INSECURE_STORE` | Allow encrypted file keyring fallback |
