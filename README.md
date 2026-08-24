# SkipTUI 🦘

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-Linux-FCC624?style=flat&logo=linux&logoColor=black)](https://kernel.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> **Process-Level Network Isolation & Proxy/VPN Tunneling TUI & CLI for Linux written in Go.**

SkipTUI lets you isolate any application, command, or interactive terminal shell through a dedicated proxy (**SOCKS5**, **HTTP**, **Shadowsocks**) or VPN (**WireGuard**, **OpenVPN**) while **leaving the rest of your host system on its regular network**.

---

## 🌟 Why SkipTUI?

Traditional VPNs force an "all-or-nothing" approach: once enabled, every application and background daemon on your machine is routed through the tunnel, often breaking local LAN devices, local dev servers (`localhost:3000`), games, or SSH connections.

Meanwhile, tools like `proxychains` rely on dynamic linker hooking (`LD_PRELOAD`), which silently fails on statically compiled binaries (Go, Rust, Zig) or direct Linux system calls.

**SkipTUI bridges this gap:**
- 🔒 **True Kernel-Level Sandboxes**: Uses native Linux Network Namespaces (`netns`) and User Namespaces (`unshare`).
- 🌐 **Zero Host Disruption**: Only the isolated process routes through the proxy/VPN.
- ⚡ **Works with All Binaries**: Go, Rust, C/C++, Java, Electron, Python, Node.js, and static binaries.
- 🛡️ **Leak-Proof DNS & KillSwitch**: Fail-closed architecture prevents DNS leaks or unencrypted packet escapes if the proxy disconnects.

---

## ✨ Features

- 🖥️ **Interactive Terminal UI**: Keyboard-driven dashboard built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).
- ⚡ **Quick Terminal Spawner (`[t]`)**: Interactive menu to choose any VPN/Proxy profile and spawn an isolated shell in a new window (Kitty, Alacritty, WezTerm, Ghostty, Foot, Tmux).
- 📝 **In-App Profile Form (`[a]` / `[e]`)**: Configure IP, Port, Credentials, Protocol, and DNS directly from the TUI.
- 🌐 **Multi-Protocol Support**:
  - **SOCKS5 / SOCKS5h** (with username/password & remote DNS)
  - **OpenVPN (`.ovpn`)** (native import with inline certificates & `auth-user-pass`)
  - **WireGuard (`.conf`)** (native interface migration & allowed IPs)
  - **HTTP / HTTPS Forward Proxy**
  - **Shadowsocks**
- 📶 **Live Latency & Bandwidth Monitor**: Real-time throughput (RX/TX bytes) and concurrent ping prober.
- 🔄 **Persistent Detachable Sessions**: Background daemon architecture keeps sessions running when the TUI closes.
- ⌨️ **Full CLI Parity**: Scriptable headless commands (`skiptui run`, `list`, `kill`, `import`, `test`) for window managers (i3, Hyprland, Rofi).

---

## 📸 TUI Interface Preview

### 1. Active Sessions Dashboard
```
┌─ SkipTUI v0.1.0 ────────────────────────────────────── [Daemon: Active | Root Mode] ─┐
│  [1] Sessions (3)  │  [2] Profiles (5)  │  [3] Traffic Logs  │  [4] Settings & Health │
├───────────────────────────────────────────────────────────────────────────────────────┤
│ ACTIVE ISOLATED SESSIONS                                                             │
│                                                                                       │
│  ID          NAME / CMD              PROFILE        PID     UPTIME    RX/TX     STATUS│
│  ● sb-90a1   firefox                 US-Residential 48102   00:14:22  14.2 MB   ● RUN │
│  ● sb-33fc   zsh (Kitty Terminal)    NL-WireGuard   48291   00:08:11   2.1 MB   ● RUN │
│  ● sb-77ab   openvpn (Corp-VPN)      Work-OpenVPN   48550   00:04:30  35.8 MB   ● RUN │
│                                                                                       │
├───────────────────────────────────────────────────────────────────────────────────────┤
│ [t] Spawn Terminal  [a] Add Profile  [e] Edit  [l] Launch App  [i] Import  [q] Quit   │
└───────────────────────────────────────────────────────────────────────────────────────┘
```

### 2. Quick Terminal Profile Selector (`[t]`)
```
┌────────────── ⚡ Select Profile for Isolated Terminal ──────────────┐
│                                                                     │
│  Spawning interactive 'zsh' terminal shell in isolated network:      │
│                                                                     │
│  ▶ [1] VPNBook-UK          [openvpn]   uk205.vpnbook.com:443   42ms │
│    [2] US-Residential      [socks5]    198.51.100.22:1080      38ms │
│    [3] Tor-Network         [socks5]    127.0.0.1:9050           1ms │
│    [4] Corporate-HTTP      [http]      corp.internal:8080        -- │
│                                                                     │
│  [↑ / ↓ / j / k] Select Profile    [Enter] Launch Terminal   [Esc]  │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 🚀 Installation & Build

### Prerequisites
- Linux Kernel `>= 5.4`
- Go `1.22+`
- (Optional) `openvpn` package for `.ovpn` profiles

### Build from Source
```bash
# Clone repository
git clone https://github.com/dvcanache/SkipTui.git
cd SkipTui

# Build binary
make build

# (Recommended) Grant Linux capabilities once for native NetNS performance:
sudo make setcap

# Run SkipTUI
./bin/skiptui
```

---

## 📖 Usage Guide

### 1. Interactive TUI
```bash
skiptui
```

| Key | Action |
| :--- | :--- |
| **`t`** | **Quick-spawn isolated terminal shell** (Opens profile selector) |
| **`a`** | **Add new proxy/VPN profile** (Interactive form) |
| **`e`** | **Edit selected profile** |
| **`i`** | **Import `.ovpn` or WireGuard `.conf`** with credentials |
| **`l`** | **Launch custom app / binary** in isolated sandbox |
| **`p` / `T`** | **Test latency** on selected or all profiles |
| **`k` / `x`** | **Terminate** selected session |
| **`1` - `4`** | Jump to Tab (Sessions, Profiles, Logs, Settings) |
| **`q`** | Quit (Prompts to Detach vs Terminate) |

---

### 2. Command-Line Interface (CLI)

```bash
# 1. Spawn an isolated terminal in a new Kitty/Alacritty window
skiptui run --profile VPNBook-UK --terminal -- zsh

# 2. Run a command inside an isolated proxy sandbox
skiptui run --profile US-Residential -- curl -s https://api.ipify.org

# 3. Launch an isolated browser instance
skiptui run --profile Tor-Network -- firefox --new-instance -P "tor-session"

# 4. Import OpenVPN (.ovpn) with credentials
skiptui import ~/Downloads/vpn.ovpn --name "My-VPN" -u "username" -p "password"

# 5. Import WireGuard (.conf)
skiptui import ~/Downloads/wg0.conf --name "Mullvad-WG"

# 6. Test latency across all profiles
skiptui test

# 7. List active sessions & profiles
skiptui list
```

---

## 📚 Architecture & Design Documentation

Comprehensive specifications and internal documentation are located in [`planning/`](planning/):

- [**00. Overview & Goals**](planning/00_OVERVIEW_AND_GOALS.md) - Problem statement & value proposition.
- [**01. Architecture & Design**](planning/01_ARCHITECTURE_AND_DESIGN.md) - Client/daemon architecture & IPC flow.
- [**02. Networking & Isolation Strategies**](planning/02_NETWORKING_AND_ISOLATION_STRATEGIES.md) - NetNS, Tun2Socks, WireGuard, OpenVPN, and DNS isolation.
- [**03. TUI UX & Workflow**](planning/03_TUI_UX_AND_WORKFLOW.md) - Screen layouts, modals, and keybindings.
- [**04. Security & Permissions**](planning/04_SECURITY_AND_PERMISSIONS.md) - Capabilities, threat model, and leak mitigation.
- [**05. Tech Stack & Dependencies**](planning/05_TECH_STACK_AND_DEPENDENCIES.md) - Go packages & kernel requirements.
- [**06. Implementation Roadmap**](planning/06_IMPLEMENTATION_ROADMAP.md) - Phased development milestones.
- [**07. Project Structure & Code Layout**](planning/07_PROJECT_STRUCTURE_AND_CODE_LAYOUT.md) - Go packages & struct models.

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
