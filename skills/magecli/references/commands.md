# magecli Command Reference

## auth

### auth login
```
magecli auth login <host> --token <token> [--allow-insecure-store]
```
Store Magento Integration bearer token for a host.

### auth status
```
magecli auth status [--json]
```
Show authentication status for all configured hosts.

### auth logout
```
magecli auth logout <host>
```
Remove stored credentials for a host.

## context

### context create
```
magecli context create <name> --host <key> [--store-code <code>] [--set-active] [--allow-writes]
```
Create a new CLI context linking a name to a host and store code. Contexts are read-only by default. Use `--allow-writes` to permit POST/PUT/DELETE via the `api` command.

### context list
```
magecli context list [--json]
```
List all configured contexts. Active context marked with *.

### context use
```
magecli context use <name>
```
Set the active context.

### context delete
```
magecli context delete <name>
```
Delete a context.

## product

### product list
```
magecli product list [--filter "field op value"] [--sort "field:DIR"] [--limit N] [--page N] [--json]
```
Search and list products. Supports multiple --filter and --sort flags.

Filter operators: eq, neq, gt, gteq, lt, lteq, like, nlike, in, nin, null, notnull, from, to, finset

### product view
```
magecli product view <sku> [--json]
```
View full product details by SKU.

### product media
```
magecli product media <sku> [--json]
```
List media gallery entries (images/videos) for a product.

### product children
```
magecli product children <sku> [--json]
```
List simple product variants of a configurable product.

### product options
```
magecli product options <sku> [--json]
```
List configurable product options (color, size, etc.).

### product search
```
magecli product search <term> [--limit N] [--json]
```
Quick search for products by name. Shortcut for `product list --filter "name like %term%"`.

### product url
```
magecli product url <url-key> [--json]
```
Find a product by its URL key (e.g., `blue-shirt`). Returns the matching product details.

## category

### category tree
```
magecli category tree [--root <id>] [--depth <n>] [--json]
```
Display the category hierarchy as a tree.

### category view
```
magecli category view <id> [--json]
```
View category details by ID.

### category products
```
magecli category products <id> [--json]
```
List products assigned to a category.

## attribute

### attribute view
```
magecli attribute view <code> [--json]
```
View a product attribute definition including options.

### attribute options
```
magecli attribute options <code> [--json]
```
List option values for a dropdown/select attribute.

### attribute sets
```
magecli attribute sets [--filter "..."] [--json]
```
List product attribute sets.

## inventory

### inventory status
```
magecli inventory status <sku> [--json]
```
Check stock status, quantity, and sale limits for a product.

## store

### store views
```
magecli store views [--json]
```
List all store views with ID, code, name, website, group, and active status.

### store config
```
magecli store config [store-code] [--json]
```
Show store configuration (locale, currency, timezone, URLs). Optionally filter by store code.

### store groups
```
magecli store groups [--json]
```
List all store groups.

### store websites
```
magecli store websites [--json]
```
List all websites.

## config

### config list
```
magecli config list [--filter <path-or-keyword>] [--json]
```
List all queryable Magento configuration values as path=value pairs. Use `--filter` to narrow results by config path prefix or keyword. Output is designed for environment comparison:
```
diff <(magecli -c staging config list) <(magecli -c prod config list)
```

### config get
```
magecli config get <path> [--json]
```
Get the value of a specific Magento configuration path (e.g., `general/locale/code`, `web/secure/base_url`). Supports exact and prefix matches. Shows values for each store scope.

### config dump
```
magecli config dump [--json] [--yaml]
```
Dump all queryable configuration as structured data. Best used with `--json` or `--yaml` for diffing between environments.

## promo

### promo catalog-rule list
```
magecli promo catalog-rule list [--filter "field op value"] [--sort "field:DIR"] [--limit N] [--page N] [--json]
```
List catalog price rules (applied to products before they are added to the cart). Alias: `promo cr list`.

### promo catalog-rule view
```
magecli promo catalog-rule view <id> [--json]
```
View a catalog price rule by ID. Shows name, action, discount amount, date range, customer groups, and websites.

### promo cart-rule list
```
magecli promo cart-rule list [--filter "field op value"] [--sort "field:DIR"] [--limit N] [--page N] [--json]
```
List cart price rules (sales rules, applied at checkout). Alias: `promo sr list`.

### promo cart-rule view
```
magecli promo cart-rule view <id> [--json]
```
View a cart price rule by ID. Shows name, action, discount, coupon type, usage limits, date range, and more.

### promo coupon list
```
magecli promo coupon list [--filter "field op value"] [--sort "field:DIR"] [--limit N] [--page N] [--json]
```
List coupon codes. Filter by code, rule_id, or other fields.

### promo coupon view
```
magecli promo coupon view <id> [--json]
```
View coupon details including code, usage stats, limits, and expiration.

## cms

### cms page list
```
magecli cms page list [--filter "field op value"] [--limit N] [--json]
```
List CMS pages with optional filtering.

### cms page view
```
magecli cms page view <id> [--content] [--json]
```
View CMS page details. Use `--content` to include HTML content.

### cms block list
```
magecli cms block list [--filter "field op value"] [--limit N] [--json]
```
List CMS static blocks with optional filtering.

### cms block view
```
magecli cms block view <id> [--content] [--json]
```
View CMS block details. Use `--content` to include HTML content.

## api

### api (raw)
```
magecli api <path> [-X METHOD] [-P key=value] [-F key=value] [-d json] [-H "Key: Value"] [--json]
```
Make raw Magento REST API requests. Path is relative to /rest/{store_code}/.

GET and HEAD requests are always allowed. Write methods (POST, PUT, DELETE, PATCH) require the active context to have been created with `--allow-writes`.

## Global Flags

| Flag | Description |
|------|-------------|
| `-c, --context <name>` | Use a specific context |
| `--store-code <code>` | Override store code |
| `--json` | JSON output |
| `--yaml` | YAML output |
| `--jq <expr>` | jq filter (requires --json) |
| `--template <tmpl>` | Go template output |
