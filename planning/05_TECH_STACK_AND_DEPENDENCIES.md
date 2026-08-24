# SkipTUI: Tech Stack & Dependencies

## 1. Programming Language & Runtime

| Technology | Selection | Rationale |
| :--- | :--- | :--- |
| **Language** | **Go (Golang 1.22+)** | - Direct access to Linux kernel syscalls (`unix.Unshare`, `unix.Setns`).<br>- Native goroutines for non-blocking TUI rendering and high-throughput TUN packet processing.<br>- Compiles to a single, statically linked binary without runtime dependencies.<br>- World-class TUI and networking libraries. |
| **Target OS** | **Linux (x86_64, aarch64)** | Relies on Linux-specific kernel features (`netns`, `veth`, `tun`, `unshare`, `wireguard`). |

---

## 2. Core Go Library Ecosystem

### 2.1 TUI & Visual Presentation
- **[`github.com/charmbracelet/bubbletea`](https://github.com/charmbracelet/bubbletea)**: The core reactive Elm-architecture framework for Go TUIs.
- **[`github.com/charmbracelet/lipgloss`](https://github.com/charmbracelet/lipgloss)**: Declarative terminal styling, borders, colors, layouts, and tables.
- **[`github.com/charmbracelet/bubbles`](https://github.com/charmbracelet/bubbles)**: Ready-made UI components (text inputs, lists, tables, spinners, viewports, file pickers).

### 2.2 Linux Networking, Namespaces & Syscalls
- **[`github.com/vishvananda/netlink`](https://github.com/vishvananda/netlink)**: Low-level Netlink communication for programmatic creation of Linux network interfaces, IP address assignment, routes, and virtual devices.
- **[`github.com/vishvananda/netns`](https://github.com/vishvananda/netns)**: Managing Linux network namespace file descriptors (`/proc/self/ns/net`).
- **[`golang.org/x/sys/unix`](https://pkg.go.dev/golang.org/x/sys/unix)**: Direct Linux system call wrappers (`unshare`, `setns`, `mount`).

### 2.3 Layer-3 / Layer-5 Tunneling & VPN
- **[`github.com/xjasonlyu/tun2socks/v2`](https://github.com/xjasonlyu/tun2socks)**: Translates raw Layer-3 IP packets from Linux TUN devices into SOCKS5/HTTP TCP streams and UDP datagrams in pure Go.
- **[`golang.zx2c4.com/wireguard/wgctrl`](https://pkg.go.dev/golang.zx2c4.com/wireguard/wgctrl)**: Cross-platform WireGuard device control library in Go.
- **OpenVPN Integration**: Native `.ovpn` parser and isolated OpenVPN worker management.

### 2.4 Process Lifecycle & IPC
- **Unix Domain Socket Server/Client**: Low-overhead IPC between TUI frontends, CLI commands, and the background session daemon.
- **Terminal Spawner**: Detection and invocation of external terminal emulators (`$TERMINAL`, `kitty`, `alacritty`, `wezterm`, `tmux`).

### 2.5 CLI & Configuration
- **[`github.com/spf13/cobra`](https://github.com/spf13/cobra)**: CLI command structure for headless scripting and flag parsing.
- **[`github.com/spf13/viper`](https://github.com/spf13/viper)**: Configuration management (JSON/YAML profile parsing and serialization).

---

## 3. System Requirements & Prerequisites

### 3.1 Linux Kernel Requirements
- Linux Kernel version `>= 5.4` (Kernel 6.x recommended).
- Kernel configuration options enabled:
  - `CONFIG_NET_NS=y` (Network Namespaces)
  - `CONFIG_USER_NS=y` (User Namespaces for Rootless mode)
  - `CONFIG_TUN=y` (Universal TUN/TAP device driver)
  - `CONFIG_WIREGUARD=m` or `=y` (for native WireGuard mode)

### 3.2 Optional External Utilities
- **`openvpn`** (Optional): For executing OpenVPN `.ovpn` profile sandboxes.
- **`slirp4netns`** (Optional): Fallback helper for rootless user namespace networking.
