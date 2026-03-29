#!/usr/bin/env bash

set -e

REPO="Sadham-Hussian/recall"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Detect ARCH
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
  ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
fi

# Get latest version
VERSION=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep tag_name | cut -d '"' -f 4)

FILE="recall_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/${VERSION}/${FILE}"

echo "Installing recall ${VERSION} for ${OS}_${ARCH}..."

curl -LO "$URL"
tar -xzf "$FILE"

chmod +x recall

# Install to /usr/local/bin (or fallback)
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

mv recall "$INSTALL_DIR"

echo "Installed to $INSTALL_DIR/recall"

# Add to PATH hint
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo "Add this to your shell:"
  echo "export PATH=\"$INSTALL_DIR:\$PATH\""
fi

echo "Done 🎉"