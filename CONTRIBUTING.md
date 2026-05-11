# Contributing to Flares

Flares is built agent-first. We design systems and use agents to implement them. Your agent is your first collaborator — point it at the `AGENTS.md` before opening issues, asking questions, or submitting code.

## The Critical Rule

**You must understand your code.** Using AI agents to write code is not just acceptable, it's how this project works. But you must be able to explain what your changes do and how they interact with the rest of the system. If you can't, don't submit it.

Submitting agent-generated code without understanding it — regardless of how clean it looks — wastes maintainer time and will result in your PR being closed.

## AI Usage

Flares is agent-first, not agent-only. The distinction matters:

- **Do** use agents to explore the codebase, run diagnostics, generate code, and iterate on implementations.
- **Do** read `AGENTS.md` — it contains the testing patterns, architecture, and conventions your agent needs to work effectively in this repo.
- **Do** interrogate your agent until you understand every edge case and interaction in your changes.
- **Don't** submit code you can't explain without your agent open.
- **Don't** use agents as a substitute for understanding the system.

## Prerequisites

Install either [mise](https://mise.jdx.dev/) or [Nix](https://nixos.org/download.html).

```bash
# mise
curl https://mise.run | sh
eval "$(~/.local/bin/mise activate bash)"  # add to your shell rc

# Nix
# nix-shell uses shell.nix at the repo root
```

Project requirements:

- Go 1.26
- golangci-lint
- goreleaser v2
- gofumpt
- Docker (for container builds only — not required for development)

## Getting Started

```bash
# Clone
git clone https://github.com/lfaoro/flares
cd flares

# Set up your dev environment (pick one):
make mise   # mise trust && mise install
make nix    # nix-shell

# Build and test
make build
make test
make lint

# Optional: dry-run release
make reltest
```

## Project Structure

| Path | Purpose |
|------|---------|
| `cmd/flares/main.go` | CLI entrypoint with `newApp()` and subcommands |
| `internal/cloudflare/` | Cloudflare API client, httptest-based mocks for tests |
| `.goreleaser.yml` | Release pipeline (v2, multi-arch, Homebrew/Scoop/Winget/AUR/Nix) |
| `mise.toml` | Tool versions (dev environment) |
| `shell.nix` | Nix dev shell (alternative to mise) |

## Main Tasks

| Task | Purpose |
|------|---------|
| `make dev` | Show available dev environments (mise / nix) |
| `make mise` | `mise install` — install tools via mise |
| `make nix` | Enter nix-shell |
| `make build` | Build binary to `./flares` |
| `make test` | Run all tests |
| `make lint` | golangci-lint (30+ linters) |
| `make vet` | go vet |
| `make fmt` | gofumpt -w . |
| `make reltest` | Dry-run goreleaser snapshot (skips docker) |
| `make release` | Full goreleaser release |
| `make clean` | Remove build artifacts |

## Testing Patterns

Tests live in two packages:

**CLI tests** (`cmd/flares/main_test.go`):
- Call `newApp().Run(args)` — not `main()`, avoids `os.Exit`
- `captureOutput()` helper replaces `os.Stdout`/`os.Stderr` — cannot use `t.Parallel()`
- Set `CLOUDFLARE_API_TOKEN=""` via `t.Setenv` to test missing token

**Cloudflare tests** (`internal/cloudflare/cloudflare_test.go`):
- Each test creates its own `httptest.NewServer` with a fresh `http.NewServeMux`
- Client constructed directly: `&Client{api: srv.URL, token: "...", http: http.DefaultClient}`
- All tests use `t.Parallel()` and `t.Context()`
- Mock both `/zones` and `/zones/{id}/dns_records/export` on the same mux

## Cloudflare Token

For manual testing, you need a Cloudflare API token:

```bash
export CLOUDFLARE_API_TOKEN="your-token"

# Verify it works
curl -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
  https://api.cloudflare.com/client/v4/user/tokens/verify
```

Create one at https://dash.cloudflare.com/profile/api-tokens with scope `Zone.DNS -> Read`.

## Pull Requests

1. Create a feature branch from `main`.
2. Make your changes with tests.
3. Run `make build && make test && make lint` to verify.
4. Open a PR.

### Commit Messages

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

```text
<type>(<scope>): <description>
```

**Types:** `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`

**Examples:**

```text
feat(cli): add json output flag to show command
fix(cloudflare): handle 429 rate limit responses
docs: update install instructions
chore(deps): bump urfave/cli to v2.27
```
