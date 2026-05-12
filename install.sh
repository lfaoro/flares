#!/bin/sh
set -eu

usage() {
  cat <<EOF
Usage: $0 [-b <bindir>] [-d] [<version>]

  -b <bindir>  Install directory (default: /usr/local/bin)
  -d           Debug output
  <version>    Release tag (default: latest)

Examples:
  curl -sfL https://github.com/lfaoro/flares/releases/latest/download/install.sh | sh
  curl -sfL https://github.com/lfaoro/flares/releases/latest/download/install.sh | sh -s -- -b /usr/local/bin v4.0.1
EOF
  exit 2
}

log() { echo "$@" 1>&2; }
debug() { [ -n "$DEBUG" ] && log "debug:" "$@"; }

BINDIR=/usr/local/bin
TAG=

while getopts "b:dh" arg; do
  case "$arg" in
    b) BINDIR="$OPTARG" ;;
    d) DEBUG=1 ;;
    h) usage ;;
    *) usage ;;
  esac
done
shift $((OPTIND - 1))
TAG=${1:-latest}

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH=amd64 ;;
  aarch64) ARCH=arm64 ;;
  armv8*) ARCH=arm64 ;;
esac

case "$OS" in
  msys_nt*|mingw*) echo "Windows is not supported"; exit 1 ;;
esac

BINARY=flares

FORMAT=tar.gz
OWNER=lfaoro
REPO=flares
RELEASES=https://github.com/$OWNER/$REPO/releases

if [ "$TAG" = latest ]; then
  TAG=$(curl -sfL "$RELEASES/latest" -o /dev/null -w '%{redirect_url}' | sed 's|.*/v|v|')
fi

VERSION=${TAG#v}
ARCHIVE="flares_${VERSION}_${OS}_${ARCH}.${FORMAT}"
DOWNLOAD="$RELEASES/download/$TAG/$ARCHIVE"
CHECKSUM_URL="$RELEASES/download/$TAG/checksums.txt"

tmpdir=$(mktemp -d)
debug "downloading $ARCHIVE from $DOWNLOAD"
curl -sfL "$DOWNLOAD" -o "$tmpdir/$ARCHIVE"

debug "verifying checksum"
curl -sfL "$CHECKSUM_URL" -o "$tmpdir/checksums.txt"
(cd "$tmpdir" && sha256sum --quiet --check --ignore-missing checksums.txt 2>/dev/null || \
  shasum -a 256 --quiet --check --ignore-missing checksums.txt 2>/dev/null || \
  { log "checksum verification failed"; exit 1; })

tar -xzf "$tmpdir/$ARCHIVE" -C "$tmpdir"
mkdir -p "$BINDIR"
install "$tmpdir/$BINARY" "$BINDIR/$BINARY"
log "installed flares $VERSION to $BINDIR/$BINARY"

rm -rf "$tmpdir"
