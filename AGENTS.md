# Flares — Agent Guide

**Module**: `github.com/lfaoro/flares` · **Go**: 1.26.3 · **CLI**: `urfave/cli/v3`

See also [CONTRIBUTING.md](CONTRIBUTING.md) for the full contribution guide (agent usage, pull requests, commit conventions).

> **Before every push**: update [CHANGES.md](CHANGES.md) with unreleased changes and keep AGENTS.md in sync with the current codebase. These files are the contract between agents and maintainers.

> **Never commit or push without asking first.** Stage changes, show the diff, and wait for explicit instructions before committing or pushing to any remote.

## Commands

```
flares show [--all, -a] [--output, -o FORMAT] [<domain>...]    Print DNS records (text or json)
flares export [--all, -a] [<domain>...]                          Write BIND zone files
flares zones                                                      List all zones
```

Global flags: `--token` (`$CLOUDFLARE_API_TOKEN`, fallback `$CF_API_TOKEN`), `--debug` (`$FLARES_DEBUG`), `--threads` (`$FLARES_THREADS`, default 10). Required scope: `Zone.DNS -> Read`.

## Build & Test

```bash
make dev                           # choose dev environment (mise or nix)
make check                         # full CI suite: tidy-check → build → vet → lint → test
make hooks                         # install .githooks/pre-push (runs make check before push)
make test                          # go test -v -race -shuffle=on -count=1 ./...
make lint                          # golangci-lint run ./... (30+ linters, see .golangci.yml)
make reltest                       # goreleaser snapshot without docker
```

Pre-push hook auto-installed via `make hooks`. It runs `make check` (go mod tidy check, build, vet, lint, test with race detector). This matches the CI pipeline — if `make check` passes locally, CI should too.

## Testing Patterns

**CLI tests** (`cmd/flares/main_test.go`):
- Call `runApp(args...)` which calls `newCmd().Run()` — not `main()`, avoids `os.Exit`
- `captureOutput()` helper replaces global `os.Stdout`/`os.Stderr` — **cannot** use `t.Parallel()`
- Use `t.Setenv("CLOUDFLARE_API_TOKEN", "")` to test missing token
- `TestMain` runs `goleak.VerifyTestMain(m)` for goroutine leak detection
- Use `respondZoneResp(w, id, name)` helper for mock zone responses
- `FuzzWriteFile` fuzzes the `writeFile` filename validation invariant

**Cloudflare tests** (`internal/cloudflare/cloudflare_test.go`):
- **Table-driven**: `TestZones` (9 subtests) and `TestExport` (5 subtests) with named subtests
- Each subtest creates its own `httptest.NewServer` with a fresh `http.NewServeMux`
- Client constructed manually: `&Client{api: srv.URL, token: "...", http: http.DefaultClient}`
- All subtests use `t.Parallel()` and `t.Context()` (Go 1.24+)
- Mock both `/zones` and `/zones/{id}/dns_records/export` endpoints on same mux
- Error paths tested: 401, 403, 429, 500, bad JSON, network errors, domain not found

## Architecture

```
cmd/flares/main.go          → newCmd(), cli.Command with subcommands, signal handling
internal/cloudflare/         → Client{api, token, http}, New(token) (*Client, error)
  .Zones(ctx)                → GET /zones?per_page=50&page=N, paginated
  .Export(ctx, domain)       → GET /zones/{id}/dns_records/export, returns BIND text
  do(ctx, method, path, q)   → JSON API helper: newRequest → do → status check → decode
  doRaw(ctx, method, path, q)→ Raw API helper: newRequest → do → status check → bytes (for Export)
  newRequest(...)             → shared request builder: url parse, auth header, context
```

- `do()` checks 401/403/429/all non-2xx explicitly before JSON decode
- `doRaw()` checks that status is 200 before returning raw bytes
- `export --all` uses semaphore (`maxConcurrent=10`, configurable via `--threads`) to limit concurrent API calls
- `show --output json` wraps BIND output in `map[string]string` JSON
- `writeFile` rejects path traversal: empty, `.`, `..`, `/`, `\` in domain names

## Code Conventions

- `fmt.Errorf("...: %w", err)` — no `pkg/errors`
- `simpleError` type for sentinel errors (`ErrNoToken`, `ErrDomainNF`)
- `os.ReadFile`/`os.WriteFile` — no `io/ioutil`
- Return errors through CLI framework — no `os.Exit` in command handlers
- `main()` wraps via `os.Exit(run())` so `defer cancel()` runs before exit

## Security

- `writeFile` validates domain names before writing (path traversal prevention)
- `--api-url` is hidden and warns on stderr when overridden
- `gosec` runs in CI via `security.yml`
- `govulncheck` runs in CI via `security.yml`
- Fuzz test verifies `writeFile` invariant: any accepted name writes inside the intended directory
- Tests cover all `do()` error paths: 401, 403, 429, 500, bad JSON, network errors

## Release

- **goreleaser v2** config at repo root (`.goreleaser.yml`)
- Version injected via `-ldflags -X main.version/commit/date`
- `make release` / `make reltest` for dry-run
- CI: GitHub Actions (`.github/workflows/ci.yml`, `release.yml`, `security.yml`)
- Dependencies: Dependabot (`.github/dependabot.yml`)
- Package managers: Homebrew (`lfaoro/tap`), Scoop (`lfaoro/tap/scoop`), Winget (`lfaoro/tap`), AUR (`lfaoro/aur`), Nix (`lfaoro/nix`)
- Docker: multi-arch via `dockers_v2` (`alpine:3.21`, `ARG TARGETPLATFORM`)
- `reltest` always skips docker (requires daemon)

## Dev Environments

Two options to install the toolchain (Go 1.26, golangci-lint, goreleaser, gofumpt):

```bash
# mise (tools defined in mise.toml)
make mise

# nix (tools defined in shell.nix)
make nix
```

## Token Verification

```bash
curl -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
  https://api.cloudflare.com/client/v4/user/tokens/verify
```
