# magecli

Magento 2 CLI tool for AI agents and developers. Queries Magento 2 stores via REST API.

## Build

Go 1.25+ required (see go.mod).

```bash
make build        # Output: bin/magecli (version from git tags/commit)
make test         # Run tests
make fmt          # go fmt
make lint         # golangci-lint run
make tidy         # go mod tidy
make verify       # go mod verify
make vulncheck    # govulncheck ./...
make clean        # rm bin/ and dist/
make sync-skills  # Sync skills/ → .claude/skills/ and .codex/skills/
make check-skills # Verify skill copies are in sync
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
    config/           System config get, list, dump
    promo/            catalog-rule, cart-rule, coupon list/view
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

## Config File

Located at `~/.config/magecli/config.yml` (override with `MAGECLI_CONFIG_DIR`):

```yaml
version: 1
active_context: production
contexts:
  production:
    host: store.example.com
    store_code: default
    allow_writes: false
hosts:
  store.example.com:
    base_url: https://store.example.com
```

Tokens are stored in the OS keyring, never written to the config file.

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

## Environment Variables

| Variable | Description |
|----------|-------------|
| `MAGECLI_TOKEN` | Bearer token (bypasses OS keyring) |
| `MAGECLI_CONFIG_DIR` | Config directory override (default: `~/.config/magecli`) |
| `MAGECLI_HTTP_DEBUG` | Enable HTTP request/response logging |
| `MAGECLI_PAGER` | Pager command override |
| `MAGECLI_ALLOW_INSECURE_STORE` | Allow encrypted file keyring fallback |

## Testing

Tests use standard Go table-driven patterns with `t.Run()` subtests. No mocking frameworks — tests use `t.TempDir()` for config isolation. Run with `make test`.

## Releasing

Releases are automated via GitHub Actions + GoReleaser:

1. Push a version tag: `git tag v0.1.0 && git push origin v0.1.0`
2. The `release.yml` workflow triggers, building binaries for linux/darwin/windows on amd64/arm64
3. GoReleaser creates a GitHub Release with archives, checksums, and changelog

Config: `.goreleaser.yml`. Version, commit, and date are injected via ldflags into `internal/build`.

## Skills (AI Agent Docs)

The `skills/magecli/` directory is the **source of truth** for AI agent documentation. Run `make sync-skills` to copy to `.claude/skills/` and `.codex/skills/`. Run `make check-skills` to verify they're in sync. Always edit `skills/` first, then sync.
