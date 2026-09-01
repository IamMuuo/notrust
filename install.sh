#!/usr/bin/env bash

set -euo pipefail

REPO="iammuuo/notrust"
PREFIX="${PREFIX:-$HOME/.local}"
CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/notrust"
SYSTEMD_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

die() {
    echo "error: $*" >&2
    exit 1
}

command -v curl >/dev/null 2>&1 || die "curl is required"
command -v tar >/dev/null 2>&1 || die "tar is required"
command -v systemctl >/dev/null 2>&1 || die "systemctl is required"

[[ "$(uname -s)" == "Linux" ]] || die "notrust currently supports Linux only"

case "$(uname -m)" in
    x86_64)
        ARCH="x86_64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        die "unsupported architecture: $(uname -m)"
        ;;
esac

echo "Installing notrust..."

# Get latest release metadata.
RELEASE_JSON="$TMP_DIR/release.json"

curl -fsSL \
    -o "$RELEASE_JSON" \
    "https://api.github.com/repos/$REPO/releases/latest"

VERSION="$(grep -m1 '"tag_name":' "$RELEASE_JSON" | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')"

[[ -n "$VERSION" ]] || die "could not determine latest release"

ARCHIVE="notrust_${VERSION#v}_Linux_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"

echo "  version: $VERSION"
echo "  arch:    $ARCH"
echo "  downloading..."

curl -fL \
    --progress-bar \
    -o "$TMP_DIR/$ARCHIVE" \
    "$URL"

echo "  extracting..."

mkdir -p "$TMP_DIR/release"

tar -xzf "$TMP_DIR/$ARCHIVE" \
    -C "$TMP_DIR/release"

[[ -f "$TMP_DIR/release/notrust" ]] ||
    die "release archive does not contain notrust"

[[ -f "$TMP_DIR/release/notrustd" ]] ||
    die "release archive does not contain notrustd"

# Install binaries.
mkdir -p "$PREFIX/bin"

install -m 0755 \
    "$TMP_DIR/release/notrust" \
    "$PREFIX/bin/notrust"

install -m 0755 \
    "$TMP_DIR/release/notrustd" \
    "$PREFIX/bin/notrustd"

# Install config without overwriting an existing one.
mkdir -p "$CONFIG_DIR"

if [[ ! -f "$CONFIG_DIR/config.yaml" ]]; then
    install -m 0644 \
        "$TMP_DIR/release/config.example.yaml" \
        "$CONFIG_DIR/config.yaml"

    echo "  created: $CONFIG_DIR/config.yaml"
else
    echo "  keeping existing config"
fi

# Install user systemd service.
mkdir -p "$SYSTEMD_DIR"

install -m 0644 \
    "$TMP_DIR/release/notrust.service" \
    "$SYSTEMD_DIR/notrust.service"

systemctl --user daemon-reload
systemctl --user enable --now notrust.service

echo
echo "notrust installed successfully."
echo
echo "  binary:  $PREFIX/bin/notrust"
echo "  daemon:  $PREFIX/bin/notrustd"
echo "  config:  $CONFIG_DIR/config.yaml"
echo "  service: notrust.service"
