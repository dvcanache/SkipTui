#!/usr/bin/env bash
set -euo pipefail

# Script to grant required Linux capabilities to the skiptui binary.
# This allows running high-performance network namespaces without full root/sudo access.

BINARY_PATH="${1:-./bin/skiptui}"

if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary not found at $BINARY_PATH"
    echo "Build the binary first using: make build"
    exit 1
fi

echo "Granting CAP_NET_ADMIN and CAP_SYS_ADMIN capabilities to $BINARY_PATH..."
sudo setcap 'cap_net_admin,cap_sys_admin+ep' "$BINARY_PATH"

echo "✓ Successfully configured Linux capabilities on $BINARY_PATH"
echo "You can now run SkipTUI in native NetNS mode without sudo!"
