# magecli

Magento 2 CLI tool for AI agents and developers. Queries Magento 2 stores via REST API.

## Build

```bash
make build        # Output: bin/magecli
make test         # Run tests
make tidy         # go mod tidy
make sync-skills  # Sync skills to .claude/ and .codex/
```

## Architecture

Modeled after bitbucket-cli (`bkt`). Key patterns:

- **Factory pattern**: `pkg/cmdutil/Factory` wires shared dependencies (config, IO, browser, pager)
- **Multi-context config**: `~/.config/magecli/config.yml` with hosts and named contexts
- **Keyring auth**: Bearer tokens in OS keyring via `github.com/99designs/keyring`, env override via `MAGECLI_TOKEN`
- **SearchCriteria builder**: `pkg/magento/search.go` translates `--filter` flags to Magento's query params
- **Output formatting**: `--json`, `--yaml`, `--jq`, `--template` via `pkg/format/`

## Project Structure

```
cmd/magecli/          Entry point
internal/
  magecmd/            Root wiring, signal handling
  build/              Version info via ldflags
  config/             YAML config, multi-store contexts
  secret/             OS keyring for bearer tokens
pkg/
  cmd/                Cobra command implementations
    root/             Command tree assembly
    factory/          Factory constructor
    auth/             login, status, logout
    context/          create, list, use, delete
    product/          list, search, view, media, children, options, url
    category/         tree, view, products
    attribute/        view, options, sets
    inventory/        status
    store/            views, config, groups, websites
    cms/              page list/view, block list/view
    api/              Raw API escape hatch
  cmdutil/            Factory, context resolution, output helpers
  magento/            REST client + SearchCriteria builder
  httpx/              HTTP client (Bearer auth, retry, cache)
  format/             JSON/YAML/jq/template output
  iostreams/          Terminal IO
  pager/              Pager manager
  progress/           Spinner
  prompter/           Interactive prompts
  browser/            URL opener
```

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/99designs/keyring` - OS keyring
- `github.com/itchyny/gojq` - jq support
- `golang.org/x/term` - Terminal detection
- `gopkg.in/yaml.v3` - YAML config + output

## Auth

Magento 2.3.2+ requires authentication for all REST reads. Uses Integration Bearer Tokens (permanent tokens from Magento Admin > System > Integrations).

## Write Safety

Contexts are **read-only by default**. The `api` command blocks non-GET/HEAD methods unless the context has `allow_writes: true` (set via `context create --allow-writes`). All first-class commands use GET only.
