# magecli

[![CI](https://github.com/atlanticbt/magecli/actions/workflows/ci.yml/badge.svg)](https://github.com/atlanticbt/magecli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/atlanticbt/magecli)](https://goreportcard.com/report/github.com/atlanticbt/magecli)
[![Go Reference](https://pkg.go.dev/badge/github.com/atlanticbt/magecli.svg)](https://pkg.go.dev/github.com/atlanticbt/magecli)
[![Release](https://img.shields.io/github/v/release/atlanticbt/magecli)](https://github.com/atlanticbt/magecli/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A CLI tool for querying Magento 2 stores via the REST API. Built for AI agents (Claude Code, Codex) and developers.

**Full documentation, examples, and AI agent integration guide:** https://atlanticbt.github.io/magecli/

## Install

**Linux / macOS** (one-liner):
```bash
curl -fsSL https://raw.githubusercontent.com/atlanticbt/magecli/main/install.sh | sh
```

**Windows** (PowerShell):
```powershell
irm https://raw.githubusercontent.com/atlanticbt/magecli/main/install.ps1 | iex
```

**From Go:**
```bash
go install github.com/atlanticbt/magecli/cmd/magecli@latest
```

**From source:**
```bash
git clone https://github.com/atlanticbt/magecli.git && cd magecli
make build
# Binary at bin/magecli
```

The install scripts detect your OS/architecture, download the latest release from GitHub, verify the SHA-256 checksum, and place the binary in your PATH. Use `--dir` to change the install location or `--version` to pin a specific release:

```bash
curl -fsSL https://raw.githubusercontent.com/atlanticbt/magecli/main/install.sh | sh -s -- --dir ~/.local/bin --version v1.0.0
```

## Quick Start

```bash
# 1. Authenticate with your Magento store
magecli auth login https://store.example.com --token <integration-bearer-token>

# 2. Create a context (read-only by default)
magecli context create production --host store.example.com --set-active

# Or allow write operations via the api command
# magecli context create production --host store.example.com --set-active --allow-writes

# 3. Query products
magecli product list --filter "name like %shirt%" --json
magecli product view SKU123 --json --jq '.name'
magecli category tree --json
magecli inventory status SKU123 --json
```

## Commands

| Command | Description |
|---------|-------------|
| `auth login/status/logout` | Manage authentication |
| `context create/list/use/delete` | Manage store contexts |
| `product list/search/view/media/children/options/url` | Catalog products |
| `category tree/view/products` | Category browsing |
| `attribute view/options/sets` | Product attributes |
| `inventory status` | Stock status |
| `store views/config/groups/websites` | Store configuration |
| `config list/get/dump` | System configuration |
| `promo catalog-rule/cart-rule/coupon list/view` | Promotions & coupons |
| `cms page list/view`, `cms block list/view` | CMS content |
| `api` | Raw REST API escape hatch (read-only by default) |

## Authentication

Magento 2.3.2+ requires authentication for all REST API access. magecli uses **Integration Access Tokens** (permanent bearer tokens) created in the Magento Admin.

### Creating an Integration Token

1. Log in to the **Magento Admin Panel**
2. Navigate to **System > Extensions > Integrations**
3. Click **Add New Integration**
4. On the **Integration Info** tab:
   - Enter a name (e.g., `magecli`)
   - Leave the callback/identity URLs blank
5. On the **API** tab:
   - Select the resource access the integration needs (use **All** for full read access, or scope to specific resources like Catalog, CMS, etc.)
6. Click **Save** and then **Activate**
7. Confirm the permissions in the popup — Magento will display four tokens:
   - Consumer Key
   - Consumer Secret
   - **Access Token** — this is the one magecli needs
   - Access Token Secret
8. Copy the **Access Token** value

### Storing the Token

```bash
# Store token in OS keyring
magecli auth login https://store.example.com --token <access-token>

# Or use environment variable
export MAGECLI_TOKEN=<access-token>

# On headless servers without a keyring, use the insecure file store
magecli auth login https://store.example.com --token <access-token> --allow-insecure-store
```

## Filtering & Sorting

```bash
magecli product list --filter "name like %shirt%" --filter "price gt 50" --sort "price:ASC" --limit 20
```

**Operators:** eq, neq, gt, gteq, lt, lteq, like, nlike, in, nin, null, notnull, from, to, finset

## Output Formats

```bash
magecli product list --json                          # JSON
magecli product list --yaml                          # YAML
magecli product list --json --jq '.items[].sku'     # jq filter
magecli product list --json --template '...'         # Go template
```

## AI Agent Integration

Install the magecli skill for Claude Code or Codex:

```bash
npx skills add atlanticbt/magecli
```

## License

MIT
