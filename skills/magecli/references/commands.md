# magecli Command Reference

## auth

### auth login
```
magecli auth login <host> [--token <token>] [--allow-insecure-store]
```
Register a host and store its Magento Integration bearer token in the OS
keyring. Run **without** `--token` in a terminal to enter the token at a
hidden prompt — preferred, since it keeps the secret out of shell history and
process lists. `--token` exists only for non-TTY automation; never place a
real token value in a generated command. With `MAGECLI_TOKEN` set in the
environment (by the user or CI secret store), `auth login <host>` just
registers the host and stores nothing.

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
magecli product list [--filter "field op value"] [--sort "field:DIR"] [--limit N] [--page N] [--fields "a,b,c"] [--json]
```
Search and list products. Supports multiple --filter and --sort flags.
`--limit` defaults to 20; values outside 1-10000 are an error. `--fields` restricts the returned
item fields (e.g. `sku,name,price`) to shrink the payload — most usefully by
omitting `custom_attributes`. `--fields` requires a structured output mode
(`--json`, `--yaml`, or `--template`). Pass the bare field list — do not wrap
it in `items[...]`; the command does that, and a double wrap makes Magento
return empty objects (magecli rejects it).

Filter operators: eq, neq, gt, gteq, lt, lteq, like, nlike, in, nin, null, notnull, from, to, finset

Bulk lookup (one request): `--filter "sku in ABC-1,ABC-2,ABC-3"`

### product view
```
magecli product view <sku> [--fields "a,b,c"] [--json]
```
View full product details by SKU. A missing SKU is an error (exit 3), making
this usable as an existence check. `--fields` trims the response and requires
`--json`, `--yaml`, or `--template`.

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
magecli product search <term> [--limit N] [--page N] [--sort "field:DIR"] [--fields "a,b,c"] [--json]
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
magecli attribute sets [--filter "..."] [--limit N] [--page N] [--json]
```
List product attribute sets. `--limit` defaults to 100.

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

## sales

All sales commands require the Integration token to have **Sales** resource
ACLs (Magento Admin > System > Integrations); without them Magento returns
403 (exit 4). List commands apply curated default field projections (an
untrimmed order runs 20-60KB); override with `--fields`.

### sales order list
```
magecli sales order list [--filter "field op value"] [--sort "field:DIR"] [--limit N] [--page N] [--fields "a,b,c"] [--json]
```
List orders. Returns order numbers, status, totals, and customer name/email
per order by default. Useful filters: `status eq processing`,
`created_at from 2026-06-01`, `customer_email eq jane@example.com`.

### sales order view
```
magecli sales order view <order number> [--id N] [--fields "a,b,c"] [--json]
```
View a single order by its human-facing order number (increment_id), or by
internal entity ID via `--id`. Includes line items and totals. The billing
address is limited to city/region/postcode — request street or telephone via
`--fields` only if truly needed.

### sales invoice list
```
magecli sales invoice list [--filter "field op value"] [--limit N] [--page N] [--json]
```
List invoices (billing documents for orders). Filter by `order_id eq N` to
find a specific order's invoices.

### sales shipment list
```
magecli sales shipment list [--filter "field op value"] [--limit N] [--page N] [--json]
```
List shipments including tracking numbers and carriers. Filter by
`order_id eq N` to find a specific order's shipments.

### sales creditmemo list
```
magecli sales creditmemo list [--filter "field op value"] [--limit N] [--page N] [--json]
```
List credit memos (refund documents). Filter by `order_id eq N` to find a
specific order's refunds.

### sales totals
```
magecli sales totals --from <date> [--to <date>] [--status <status>] [--json]
```
Sum order grand totals over a date range, grouped by currency. `--to`
defaults to now. Note: gross is grand_total (includes tax and shipping), not
net revenue. Scans at most 10,000 orders; the result says if the range was
too large to complete — narrow the range and sum the pieces.

## customer

Customer commands require the Integration token to have **Customers**
resource ACLs; without them Magento returns 403 (exit 4). Names and emails
are returned by default — they are the lookup keys — but postal addresses and
phone numbers are excluded unless explicitly requested.

### customer search
```
magecli customer search [--filter "field op value"] [--sort "field:DIR"] [--limit N] [--page N] [--fields "a,b,c"] [--json]
```
Search customer accounts, e.g. `--filter "email like %@example.com"` or
`--filter "created_at from 2026-01-01"`. Never returns addresses.

### customer view
```
magecli customer view <id|email> [--include-addresses] [--json]
```
View a single customer by numeric ID or email. Postal addresses and phone
numbers are only returned with `--include-addresses`. On multi-website stores
the same email can belong to several accounts; when that happens all matches
are listed — re-run with a numeric ID to pick one.

## cms

### cms page list
```
magecli cms page list [--filter "field op value"] [--limit N] [--page N] [--json]
```
List CMS pages with optional filtering. HTML bodies are omitted from list output;
retrieve a single body via `cms page view <id> --content`. You can still search
bodies with `--filter "content like %term%"`.

### cms page view
```
magecli cms page view <id> [--content] [--json]
```
View CMS page details. The HTML body is omitted (in JSON/YAML too) unless `--content` is passed.

### cms block list
```
magecli cms block list [--filter "field op value"] [--limit N] [--page N] [--json]
```
List CMS static blocks with optional filtering. HTML bodies are omitted from list output.

### cms block view
```
magecli cms block view <id> [--content] [--json]
```
View CMS block details. The HTML body is omitted (in JSON/YAML too) unless `--content` is passed.

## api

### api (raw)
```
magecli api <path> [-X METHOD] [-P key=value] [-F key=value] [-d json] [-H "Key: Value"] [--json]
```
Make raw Magento REST API requests. Path is relative to /rest/{store_code}/.

GET and HEAD requests are always allowed. Write methods (POST, PUT, DELETE, PATCH) require the active context to have been created with `--allow-writes`.

## update

### update
```
magecli update [--force]
```
Self-update magecli to the latest GitHub release. Downloads the appropriate binary for your OS/architecture, verifies the SHA-256 checksum, and replaces the current binary. Use `--force` to reinstall even if already on the latest version.

## Global Flags

| Flag | Description |
|------|-------------|
| `-c, --context <name>` | Use a specific context |
| `--store-code <code>` | Override store code (honored by all commands) |
| `--json` | JSON output |
| `--yaml` | YAML output |
| `--jq <expr>` | jq filter (requires --json) |
| `--template <tmpl>` | Go template output |

## Exit Codes

| Exit | Meaning |
|------|---------|
| 0 | Success |
| 1 | Usage/input/config error |
| 2 | Network failure (DNS, connection refused, timeout) |
| 3 | HTTP 404 — resource not found |
| 4 | HTTP 401/403 — authentication/authorization failure |
| 5 | Other HTTP error (4xx/5xx) |
| 8 | Operation pending |
