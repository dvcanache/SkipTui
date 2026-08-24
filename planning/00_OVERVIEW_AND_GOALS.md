# SkipTUI: Overview & Project Goals

## 1. Executive Summary

**SkipTUI** is a modern Terminal User Interface (TUI) and Command-Line tool built with Golang designed for **granular, per-process network isolation and proxy/VPN tunneling on Linux**. 

Traditional VPNs and system-wide proxy clients force an "all-or-nothing" approach: once active, all network traffic from the entire operating system is redirected through the tunnel. This often breaks local LAN services, slows down unaffected tasks (e.g., local development servers, streaming, gaming, SSH sessions), and creates unnecessary bandwidth overhead or privacy conflicts.

SkipTUI solves this problem by allowing users to launch specific applications, CLI tools, or interactive terminal shells in **fully isolated network sandboxes** (using Linux Network Namespaces, User Namespaces, `tun2socks`, WireGuard, and OpenVPN). The host operating system remains connected to the default local network without disruption, while the isolated process communicates exclusively through the designated proxy (SOCKS5, HTTP, Shadowsocks) or VPN (WireGuard, OpenVPN).

---

## 2. Core Problem & Value Proposition

### 2.1 The Problem
- **System-Wide Hijacking**: Activating a system VPN routes background sync services, package managers, local database connections, and unrelated apps through the VPN.
- **Flaky Proxy SOCKSifiers**: Tools like `proxychains` rely on dynamic linker hooking (`LD_PRELOAD`), which:
  - Fails completely with statically compiled binaries (e.g., Golang, Rust, Zig).
  - Fails with binaries making direct Linux system calls (bypassing `libc`).
  - Leaks DNS queries or UDP traffic if not meticulously configured.
- **Complex Manual Namespace Setup**: Manually configuring Linux `ip netns`, `veth` pairs, routing tables, NAT/iptables rules, and `/etc/netns/` resolver configs is tedious, error-prone, and requires root permissions.
- **Lack of Visibility**: Users cannot easily see active isolated sessions, real-time throughput, connection health, or process logs in an intuitive interface.

### 2.2 The SkipTUI Solution
| Feature | Traditional System VPN | `proxychains` / `LD_PRELOAD` | **SkipTUI** |
| :--- | :--- | :--- | :--- |
| **Isolation Granularity** | Entire Host OS | Process-level (dynamic binaries only) | **True Process & Container-grade Sandbox** |
| **Go / Rust / Static Binary Support** | Yes | No (fails silently or bypasses proxy) | **Yes (Kernel-level network namespace)** |
| **Host Network Disruption** | High (all traffic rerouted) | None | **Zero (Host stays on default network)** |
| **Protocol Support** | VPN only | TCP SOCKS/HTTP (often no UDP) | **SOCKS5, HTTP, Shadowsocks, WireGuard, OpenVPN** |
| **DNS Leak Mitigation** | Varies | Poor (prone to glibc leaks) | **Strict (Isolated resolv.conf / DNS proxy)** |
| **User Experience** | Desktop GUI / CLI daemon | Pure CLI wrapper | **Interactive TUI + Full Headless CLI Parity** |
| **Process Model** | Global | Inline execution | **External Terminal Spawner + Detachable Daemons** |
| **Rootless Execution** | No | Yes | **Yes (via User Namespaces + slirp4netns/gVisor)** |

---

## 3. Primary Use Cases

1. **Isolated Browsing & Testing**:
   Launch a browser instance (Chromium, Firefox) routed through a residential proxy, OpenVPN config, or foreign WireGuard server while maintaining local access to `localhost:3000` on the main desktop.
2. **Terminal Sandboxing via External Window/Pane**:
   Spawn an interactive shell (`bash`, `zsh`, `fish`) in a dedicated terminal emulator window (e.g. Alacritty, Kitty, WezTerm, Tmux) where all commands (`curl`, `git`, `ssh`, `pacman`, `apt`) automatically route through a secure SOCKS5/WireGuard/OpenVPN tunnel while SkipTUI monitors bandwidth in real-time.
3. **OpenVPN Profile Isolation**:
   Import and run standard `.ovpn` configuration files (corporate VPNs, ProtonVPN, Mullvad, etc.) isolated strictly to specific applications without overriding host default routes.
4. **Geo-Location & Network Testing for Developers**:
   Test how web applications, APIs, or scrapers behave across multiple proxies or geolocations simultaneously by running separate isolated windows.
5. **Security & Anonymity Tools**:
   Run tools through Tor, SOCKS5, or WireGuard tunnels without leaking host IP or DNS.
6. **Multi-Identity / Multi-Account Environments**:
   Run multiple instances of desktop apps (e.g., Telegram, Discord, Steam) with each assigned to a different proxy profile.

---

## 4. Key Goals & Success Criteria

### 4.1 Functional Goals
- [ ] **TUI Dashboard**: Interactive, keyboard-driven UI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).
- [ ] **Full CLI Parity**: Headless subcommands (`skiptui run`, `skiptui list`, `skiptui kill`, `skiptui import`, `skiptui test`) for scripting and window-manager shortcuts (i3/Hyprland/Rofi).
- [ ] **Multi-Protocol Engine**: Support for **SOCKS5/SOCKS5h**, **WireGuard**, **OpenVPN (`.ovpn`)**, **HTTP/HTTPS**, and **Shadowsocks**.
- [ ] **External Terminal Spawner**: Automatically detect and launch isolated shells in user-preferred terminal emulators or Tmux panes.
- [ ] **Persistent & Detachable Sessions**: Background daemon architecture allowing sessions to persist when SkipTUI closes and reconnect seamlessly upon restart.
- [ ] **Zero DNS Leaking**: Enforce strict per-namespace DNS routing and D-Bus/systemd-resolved leak guards.

### 4.2 Technical & Performance Goals
- **Minimal Resource Footprint**: Written in Go with minimal background overhead (< 30MB RAM idle).
- **Fast Startup**: Spawn an isolated sandbox environment in < 200ms.
- **Robust Cleanup**: Guarantee cleanup of network namespaces, virtual interfaces, routing rules, and background daemons on process termination or exit.
- **Dual Mode (Privileged & Rootless)**:
  - *Privileged Mode*: Linux Network Namespaces (`netns`) + `veth` / `tun2socks` / kernel WireGuard / OpenVPN.
  - *Rootless Mode*: User Namespaces (`unshare -U -n`) + `slirp4netns` or embedded user-space TCP/IP stack (gVisor `netstack`).
