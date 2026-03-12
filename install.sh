#!/bin/sh
# install.sh — Download and install the latest magecli binary.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/atlanticbt/magecli/main/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/atlanticbt/magecli/main/install.sh | sh -s -- --dir /usr/local/bin
#   curl -fsSL https://raw.githubusercontent.com/atlanticbt/magecli/main/install.sh | sh -s -- --version v1.2.0
#
set -e

REPO="atlanticbt/magecli"
BINARY="magecli"
INSTALL_DIR="/usr/local/bin"
VERSION=""

# ---------- Parse flags ----------

while [ $# -gt 0 ]; do
  case "$1" in
    --dir)      INSTALL_DIR="$2"; shift 2 ;;
    --version)  VERSION="$2"; shift 2 ;;
    --help|-h)
      echo "Usage: install.sh [--dir <path>] [--version <tag>]"
      echo ""
      echo "Options:"
      echo "  --dir       Install directory (default: /usr/local/bin)"
      echo "  --version   Specific version tag (default: latest)"
      exit 0
      ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# ---------- Detect platform ----------

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *)      echo "Error: unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64)   ARCH="amd64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  *)               echo "Error: unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Detected platform: ${OS}/${ARCH}"

# ---------- Resolve version ----------

if [ -z "$VERSION" ]; then
  echo "Fetching latest release..."
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
  if [ -z "$VERSION" ]; then
    echo "Error: could not determine latest version. Check https://github.com/${REPO}/releases"
    exit 1
  fi
fi

# Strip leading 'v' for the archive filename (goreleaser uses version without v)
VERSION_NUM="${VERSION#v}"

echo "Installing ${BINARY} ${VERSION}..."

# ---------- Download and extract ----------

ARCHIVE="${BINARY}_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading ${URL}..."
curl -fsSL -o "${TMPDIR}/${ARCHIVE}" "$URL"

# Verify checksum if checksums.txt is available
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
if curl -fsSL -o "${TMPDIR}/checksums.txt" "$CHECKSUM_URL" 2>/dev/null; then
  EXPECTED="$(grep "${ARCHIVE}" "${TMPDIR}/checksums.txt" | awk '{print $1}')"
  if [ -n "$EXPECTED" ]; then
    if command -v sha256sum >/dev/null 2>&1; then
      ACTUAL="$(sha256sum "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
      ACTUAL="$(shasum -a 256 "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')"
    else
      ACTUAL=""
    fi
    if [ -n "$ACTUAL" ]; then
      if [ "$EXPECTED" != "$ACTUAL" ]; then
        echo "Error: checksum mismatch!"
        echo "  expected: ${EXPECTED}"
        echo "  actual:   ${ACTUAL}"
        exit 1
      fi
      echo "Checksum verified."
    fi
  fi
fi

tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

# ---------- Install ----------

if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

chmod +x "${INSTALL_DIR}/${BINARY}"

echo ""
echo "${BINARY} ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Run '${BINARY} --version' to verify."

# Check if install dir is in PATH
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "Warning: ${INSTALL_DIR} is not in your PATH."
    echo "Add it with: export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac
