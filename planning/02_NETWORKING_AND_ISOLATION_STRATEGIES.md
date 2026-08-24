# SkipTUI: Networking & Isolation Strategies

## 1. Deep Dive: Linux Network Namespaces (`netns`)

Linux Network Namespaces provide complete virtualization of the network stack:
- Independent network interfaces (e.g. `lo`, `tun0`, `wg0`, `tap0`)
- Independent routing tables (`ip route show`)
- Independent firewall and packet filtering rules (`nftables` / `iptables`)
- Independent socket lists and port binding spaces (`ss -tuln`)

Processes executing inside a network namespace have zero visibility into host network interfaces (`eth0`, `wlan0`, `docker0`) unless explicitly bridged or routed.

---

## 2. Supported Isolation & Tunneling Strategies

SkipTUI provides dedicated backends for all major proxy and VPN protocols:

```
+-----------------------------------------------------------------------------------------+
|                                  ISOLATION STRATEGIES                                   |
+----------------------+-----------------------+--------------------+---------------------+
| 1. NetNS + Tun2Socks | 2. NetNS + WireGuard  | 3. NetNS + OpenVPN | 4. Rootless UserNS  |
+----------------------+-----------------------+--------------------+---------------------+
| - SOCKS5 / SOCKS5h   | - Kernel WireGuard    | - Native `.ovpn`   | - Zero root/sudo    |
| - HTTP / Shadowsocks | - Extreme speed (L3)  |   profiles         | - `unshare -U -n`   |
| - L3 to L5 pure Go   | - Moves `wg0` to ns   | - Isolated daemon  | - User-space TCP/IP |
|   packet translator  | - Zero host leakage   | - Inline certs/keys|   (slirp4netns)     |
+----------------------+-----------------------+--------------------+---------------------+
```

---

## 3. Strategy 1: Network Namespace + `tun2socks` (SOCKS5 & HTTP)

For generic proxies (SOCKS5, SOCKS5h, HTTP, Shadowsocks), the sandbox traffic is captured at Layer-3 (IP) and forwarded at Layer-5 (TCP/UDP proxy streams):

```
[ Target Process inside Sandbox (e.g. curl/firefox) ]
                      |
                      v (generates TCP/UDP/DNS packets)
               [ Virtual TUN (tun0) ]
                      |
                      v (raw IP packet buffer)
     [ Go Embedded Tun2Socks Engine (in Host or NS) ]
                      |
                      v (converts IP packet to SOCKS5 TCP connect / UDP associate)
      [ Host Physical Network Stack (eth0/wlan0) ]
                      |
                      v (encrypted / proxied TCP socket)
              [ Remote Proxy Server ]
```

---

## 4. Strategy 2: Network Namespace + WireGuard (Native VPN)

WireGuard interfaces can be created in the host namespace (inheriting the host's physical network routing for UDP peer handshakes) and then moved directly into the target namespace:

1. Create WireGuard link in host: `ip link add wg-skiptui type wireguard`
2. Configure WireGuard private keys, endpoints, and allowed IPs on host.
3. Move link to target namespace: `ip link set wg-skiptui netns skiptui-<id>`
4. Switch into namespace context: assign IP address, bring `wg-skiptui` UP, and set default route `default dev wg-skiptui`.
5. **Result**: Host routing is untouched; sandbox processes can only communicate through the WireGuard link.

---

## 5. Strategy 3: Network Namespace + OpenVPN (`.ovpn` Profiles)

OpenVPN is a widely used VPN protocol in corporate networks, ProtonVPN, Mullvad, etc. SkipTUI provides native support for `.ovpn` files.

### 5.1 OpenVPN Namespace Architecture
```
HOST NAMESPACE:
- Default network (eth0 / wlan0 untouched)
- OpenVPN socket connects to remote OpenVPN server (UDP/TCP:1194) via host stack or veth gateway

ISOLATED NAMESPACE:
- lo (127.0.0.1)
- tun0 (Assigned VPN IP, e.g., 10.8.0.5)
- Routing: default via 10.8.0.1 dev tun0
- Process: Target app (Browser / Terminal) + OpenVPN worker process
```

### 5.2 Execution Flow for `.ovpn` Profiles
1. **Profile Parsing**: SkipTUI parses `.ovpn` configuration files, extracting remote hosts, ports, protocols (UDP/TCP), certificates (`<ca>`, `<cert>`, `<key>`, `<tls-auth>`), and cipher settings.
2. **Credential Injection**: Supports inline credentials or external `auth-user-pass` files securely referenced in the profile.
3. **Namespace Spawning**: OpenVPN is launched with `--dev tun0 --redirect-gateway def1` strictly inside the target network namespace, preventing any host-level gateway override.
4. **DNS Extraction**: Automatically extracts pushed DNS options (`dhcp-option DNS ...`) and writes them to the isolated namespace's `/etc/resolv.conf`.

---

## 6. Strategy 4: User Namespaces + `slirp4netns` (Rootless Mode)

For unprivileged environments where the user has no `sudo` access or `CAP_NET_ADMIN`:
- Uses Linux User Namespaces (`unshare -U -n --map-root-user`).
- Spawns `slirp4netns` or an embedded **gVisor `netstack`** user-space virtual network interface.
- Translates sandbox traffic into user-space outbound connections without requiring root privileges.

---

## 7. Strict DNS Leak Prevention & Resolution Strategy

SkipTUI enforces three layers of defense against DNS leaks:

### 7.1 Isolated Mount Namespaces & `/etc/resolv.conf`
When launching a process, SkipTUI spawns a private Mount Namespace (`CLONE_NEWNS`):
- Bind-mounts an isolated `resolv.conf` (e.g. `/run/user/<UID>/skiptui/<id>/resolv.conf`) over `/etc/resolv.conf` inside the sandbox only.
- Directs all DNS lookups to:
  - The WireGuard / OpenVPN pushed DNS server, or
  - User-configured DNS (e.g. `1.1.1.1`, `9.9.9.9`), or
  - An embedded local DNS proxy running on `10.0.0.1:53` that forwards DNS over SOCKS5.

### 7.2 Host Systemd-Resolved & D-Bus Blocking
- Modern Linux distributions route DNS over D-Bus to `systemd-resolved` on the host.
- SkipTUI unsets and sanitizes `DBUS_SESSION_BUS_ADDRESS` and `DBUS_SYSTEM_BUS_ADDRESS` inside the sandbox to prevent applications from bypassing the namespace's network stack via IPC to the host resolver.

---

## 8. Strategy Comparison Matrix

| Criteria | NetNS + Tun2Socks | NetNS + WireGuard | NetNS + OpenVPN | UserNS + slirp4netns |
| :--- | :--- | :--- | :--- | :--- |
| **Privileges Required** | `CAP_NET_ADMIN` / sudo | `CAP_NET_ADMIN` / sudo | `CAP_NET_ADMIN` / sudo | **None (100% Rootless)** |
| **Supported Protocols** | SOCKS5, HTTP, Shadowsocks | WireGuard | OpenVPN (`.ovpn`) | SOCKS5, HTTP |
| **Throughput & Latency** | High (~1-2 Gbps) | Maximum (Kernel WireGuard) | High (Native OpenVPN) | Moderate (~500 Mbps) |
| **UDP & ICMP Support** | Full | Full | Full | Full |
| **DNS Leak Resistance** | Guaranteed (Fail-closed)| Guaranteed (Fail-closed) | Guaranteed (Fail-closed)| High |
| **Static Binary Support**| 100% | 100% | 100% | 100% |
