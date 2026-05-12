# Flares 🔥

Cloudflare DNS backup tool — exports DNS records as BIND-formatted zone files to stdout or disk.

## Quick Start

```bash
export CLOUDFLARE_API_TOKEN="KClp4y8BgD2LQiz2..."

# Show DNS records for a domain
flares show example.com

# Export as BIND zone file
flares export example.com

# All zones at once
flares show --all
flares export --all

# List zones
flares zones
```

Get your token at https://dash.cloudflare.com/profile/api-tokens (Create Token → Zone.DNS → Read → All zones).

## Install

### One-liner

```bash
curl -sfL https://github.com/lfaoro/flares/releases/latest/download/install.sh | sh
```

Install a specific version to a custom path:

```bash
curl -sfL https://github.com/lfaoro/flares/releases/latest/download/install.sh | sh -s -- -b /usr/local/bin v4.0.1
```

### Package Managers

| Platform | Command |
|----------|---------|
| macOS / Linux | `brew install lfaoro/tap/flares` |
| Windows | `scoop bucket add lfaoro https://github.com/lfaoro/tap`<br>`scoop install lfaoro/flares` |
| Windows | `winget install lfaoro.flares` |
| Arch Linux | `yay -S flares-bin` (AUR) |
| Nix | `nix profile install github:lfaoro/tap#flares` |
| Go | `go install github.com/lfaoro/flares/cmd/flares@latest` |

### Pre-built Binaries

Download the latest archive for your platform from the [releases page](https://github.com/lfaoro/flares/releases), then:

```bash
tar xzf flares_*.tar.gz
sudo mv flares /usr/local/bin/
```

### Docker

```bash
docker pull ghcr.io/lfaoro/flares
docker run --rm -e CLOUDFLARE_API_TOKEN="$CLOUDFLARE_API_TOKEN" \
  ghcr.io/lfaoro/flares show example.com
```

## Usage

```
flares [--token TOKEN] [--debug] <command> [flags] [<domain>...]

Commands:
  show      Print DNS records
    --all, -a             All zones
    --output, -o FORMAT   Output format: text (default) or json

  export    Write BIND zone files
    --all, -a             All zones

  zones     List all zone IDs and names

Global flags:
  --token, -t   Cloudflare API token  [$CLOUDFLARE_API_TOKEN, $CF_API_TOKEN]
  --debug, -d   Enable debug output   [$FLARES_DEBUG]
  --threads, -c Max concurrent API requests for --all (default: 10) [$FLARES_THREADS]
```

### JSON Output

```bash
flares show --output json example.com
```

Returns `{"example.com": "; Domain: example.com\n...BIND records..."}` — useful for scripting.

### Using --all

```bash
# Dump all zones to stdout
flares show --all

# Export every zone to its own file in the current directory
flares export --all
```

Concurrent exports are throttled to 10 simultaneous requests to avoid Cloudflare rate limits.

## Token

Create a token at https://dash.cloudflare.com/profile/api-tokens with:

- **Permissions**: Zone → DNS → Read
- **Zone Resources**: Include → All zones

Verify your token:

```bash
curl -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
  https://api.cloudflare.com/client/v4/user/tokens/verify
```

## Development

```bash
nix-shell                 # Go 1.26, golangci-lint, goreleaser, gofumpt
make build               # Build binary
make test                # Run all tests
make lint                # golangci-lint (30+ linters)
make reltest             # Dry-run goreleaser snapshot
```
