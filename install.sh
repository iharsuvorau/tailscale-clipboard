#!/usr/bin/env bash
set -euo pipefail

BINARY=clipboard-server
INSTALL_DIR="$HOME/.local/bin"
SERVICE=clipboard-server.service
SERVICE_DIR="$HOME/.config/systemd/user"

cd "$(dirname "$0")"

if systemctl --user is-active --quiet "$BINARY"; then
    echo "Stopping $BINARY..."
    systemctl --user stop "$BINARY"
fi

echo "Building..."
go build -o "$BINARY" .

echo "Installing binary to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
mv "$BINARY" "$INSTALL_DIR/"

echo "Installing service..."
mkdir -p "$SERVICE_DIR"
cp "$SERVICE" "$SERVICE_DIR/"
systemctl --user daemon-reload
systemctl --user enable --now "$BINARY"

echo "Done. Status:"
systemctl --user status "$BINARY" --no-pager
