#!/usr/bin/env bash
set -euo pipefail

# Script to grant required Linux capabilities to the skiptui binary and OpenVPN.
# This allows running high-performance network namespaces and VPNs without full root/sudo access.

BINARY_PATH="${1:-./bin/skiptui}"

if [ ! -f "$BINARY_PATH" ]; then
    echo "Building binary first..."
    go build -o "$BINARY_PATH" skiptui/cmd/skiptui
fi

echo "Granting CAP_NET_ADMIN and CAP_SYS_ADMIN capabilities to $BINARY_PATH..."
sudo setcap 'cap_net_admin,cap_sys_admin+ep' "$BINARY_PATH"

OPENVPN_BIN="$(which openvpn 2>/dev/null || true)"
if [ -n "$OPENVPN_BIN" ]; then
    echo "Granting CAP_NET_ADMIN capability to $OPENVPN_BIN..."
    sudo setcap 'cap_net_admin+ep' "$OPENVPN_BIN"
fi

echo "✓ Successfully configured Linux capabilities!"
echo "You can now run SkipTUI and OpenVPN in native NetNS mode without sudo!"
