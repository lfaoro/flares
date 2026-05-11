# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Unreleased

### Changed

- Upgraded Go from 1.17 to 1.26
- Migrated CLI framework from urfave/cli v1 to urfave/cli v3
- Replaced `github.com/pkg/errors` with `fmt.Errorf` wrapping
- Replaced deprecated `io/ioutil` with `os`/`io` equivalents
- Split CLI into subcommands: `show`, `export`, `zones`
- Removed `--export` flag in favor of `export` subcommand
- Replaced `New(token)` panicking with `New(token) (*Client, error)`
- Updated Cloudflare API response struct to match current API spec
- Moved from `log.Fatal` to proper error returns
- Restructured `main()` to `os.Exit(run())` for deferred cleanup
- Changed primary token env var from `CF_API_TOKEN` to `CLOUDFLARE_API_TOKEN`
- Changed license from BSD 3-Clause to MIT
- Updated Dockerfile to `dockers_v2` pattern (`ARG TARGETPLATFORM`)
- Rebuilt `.goreleaser.yml` for v2 format with multi-arch + package managers
- Rebuilt `.golangci.yml` with 30+ linters
- Cleaned up indirect dependencies in `go.mod`

### Added

- `--threads` flag for controlling export concurrency (default 10)
- `--api-url` hidden flag for testing with mock servers
- `--output json` flag for `show` command
- Concurrency semaphore for `export --all` to prevent API rate limiting
- HTTP error handling in `do()` helper (401/403/429/non-2xx)
- GitHub Actions CI/CD: `ci.yml`, `release.yml`, `security.yml`
- Dependabot configuration for Go, Docker, and Actions
- CLI end-to-end tests with `httptest` mock servers
- HTTP error status tests (403 on zones and export endpoints)
- `AGENTS.md` for LLM-assisted development
- `CONTRIBUTING.md` with agent-first workflow guide
- `shell.nix` for Nix dev environment
- `mise.toml` for mise tool versions
- `CHANGES.md` for tracking notable changes
- Makefile targets: `build-all`, `fmt`, `vet`, `lint`, `dev`, `mise`, `test`

### Removed

- `static/` directory (outdated demo recordings)
- `github.com/pkg/errors` dependency
- `--export` flag (replaced by `export` subcommand)

### Fixed

- `exitAfterDefer`: `os.Exit` now runs after deferred cleanup
- Linter warnings: errcheck, gosec, revive, gocritic, errname, testifylint, usetesting
