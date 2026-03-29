#!/usr/bin/env bash
set -e

REPO="Sadham-Hussian/recall"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')

ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
  ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
elif [ "$ARCH" = "arm64" ]; then
  ARCH="arm64"
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

VERSION=$(curl -fsSL https://api.github.com/repos/$REPO/releases/latest | grep tag_name | cut -d '"' -f 4)

FILE="recall_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/${VERSION}/${FILE}"

echo "Installing recall ${VERSION} for ${OS}_${ARCH}..."
echo "Downloading: $URL"

# Try download
if ! curl -fLO "$URL"; then
  echo "⚠️ Failed for ${ARCH}, falling back to amd64..."
  FILE="recall_${OS}_amd64.tar.gz"
  URL="https://github.com/$REPO/releases/download/${VERSION}/${FILE}"
  curl -fLO "$URL"
fi

tar -xzf "$FILE"
chmod +x recall

INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

mv recall "$INSTALL_DIR"

echo "Installed to $INSTALL_DIR/recall"

if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo ""
  echo "Add this to your shell:"
  echo "export PATH=\"$INSTALL_DIR:\$PATH\""
fi

echo "Done 🎉"