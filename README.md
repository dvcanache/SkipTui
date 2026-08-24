# SkipTUI 🦘

> **Process-Level Network Isolation & Proxy/VPN Tunneling TUI & CLI for Linux written in Go.**

SkipTUI lets you isolate any application, command, or interactive terminal shell through a dedicated proxy (SOCKS5, HTTP, Shadowsocks) or VPN (WireGuard, OpenVPN) while **leaving the rest of your host system on its regular network**.

---

## ✨ Key Features

- 🔒 **Process-Level Network Isolation**: Runs applications inside dedicated Linux Network Namespaces (`netns`) or Rootless User Namespaces (`unshare`).
- 🌐 **Zero Host Disruption**: Only the target application's traffic is proxied. Your host's default routes, local LAN, SSH connections, and other desktop apps stay 100% untouched.
- ⚡ **Multi-Protocol Support**: First-class support for **SOCKS5/SOCKS5h**, **WireGuard**, **OpenVPN (`.ovpn`)**, **HTTP/HTTPS**, and **Shadowsocks**.
- 📟 **External Terminal Spawner**: Automatically detects and spawns isolated shell sessions in your preferred terminal emulator (Kitty, Alacritty, WezTerm, Tmux) while keeping the SkipTUI dashboard visible.
- 🔄 **Persistent & Detachable Sessions**: Background daemon architecture allows sessions to persist when the TUI closes and reconnect seamlessly upon restart.
- 🛡️ **Leak-Proof DNS & KillSwitch**: Fail-closed architecture prevents DNS leaks or unencrypted packet escapes if the proxy disconnects.
- 💻 **Modern Terminal UI**: Interactive keyboard-driven dashboard built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).
- ⌨️ **Full CLI Parity**: Scriptable headless CLI (`skiptui run`, `skiptui list`, `skiptui kill`, `skiptui import`, `skiptui test`) for window-manager keybindings (i3, Hyprland, Rofi).

---

## 📚 Comprehensive Planning & Architecture Docs

Detailed engineering specifications and roadmaps are organized in the [`planning/`](file:///home/dvcanache/Workspaces/skiptui/planning/) directory:

| Document | Description |
| :--- | :--- |
| 📄 [**00. Overview & Goals**](file:///home/dvcanache/Workspaces/skiptui/planning/00_OVERVIEW_AND_GOALS.md) | Executive summary, problem definition, value proposition, and success criteria. |
| 📄 [**01. Architecture & Design**](file:///home/dvcanache/Workspaces/skiptui/planning/01_ARCHITECTURE_AND_DESIGN.md) | Client/daemon architecture, Unix socket IPC, sequence diagrams, and lifecycle flow. |
| 📄 [**02. Networking & Isolation Strategies**](file:///home/dvcanache/Workspaces/skiptui/planning/02_NETWORKING_AND_ISOLATION_STRATEGIES.md) | Deep dive on Linux NetNS, Tun2Socks, WireGuard, OpenVPN `.ovpn` isolation, rootless mode, and DNS leak prevention. |
| 📄 [**03. TUI UX, Interface & Workflow**](file:///home/dvcanache/Workspaces/skiptui/planning/03_TUI_UX_AND_WORKFLOW.md) | Screen layouts, ASCII mockups, keybinding maps, external terminal spawner, and exit dialog. |
| 📄 [**04. Security, Permissions & Threat Model**](file:///home/dvcanache/Workspaces/skiptui/planning/04_SECURITY_AND_PERMISSIONS.md) | Linux capabilities (`setcap`), rootless user namespaces, credential management, and crash garbage collection. |
| 📄 [**05. Tech Stack & Dependencies**](file:///home/dvcanache/Workspaces/skiptui/planning/05_TECH_STACK_AND_DEPENDENCIES.md) | Golang libraries (`bubbletea`, `netlink`, `tun2socks`, `cobra`, `viper`) and Linux kernel requirements. |
| 📄 [**06. Implementation Roadmap**](file:///home/dvcanache/Workspaces/skiptui/planning/06_IMPLEMENTATION_ROADMAP.md) | Phased milestones from core NetNS engine to full TUI, testing strategies, and risk mitigations. |
| 📄 [**07. Project Structure & Code Layout**](file:///home/dvcanache/Workspaces/skiptui/planning/07_PROJECT_STRUCTURE_AND_CODE_LAYOUT.md) | Standard Go project directory layout, key interfaces, struct models, and XDG configuration format. |

---

## 🚀 Quick Usage Concept

```bash
# Launch the interactive TUI Dashboard
skiptui

# Or use direct headless CLI commands:
# Launch an isolated terminal in a new Kitty/Alacritty window via OpenVPN
skiptui run --profile work-vpn --terminal -- zsh

# Launch Firefox through a SOCKS5 proxy sandbox
skiptui run --profile us-residential -- firefox

# Import an OpenVPN configuration file
skiptui import ~/Downloads/client.ovpn --name "Client-VPN"

# Test latency across all profiles
skiptui test
```
