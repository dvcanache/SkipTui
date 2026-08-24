# SkipTUI: TUI UX, Interface Design & Workflow

## 1. Design Philosophy & Technology

SkipTUI's interface is built on the **Elm Architecture** via [Bubble Tea](https://github.com/charmbracelet/bubbletea), with styling and layout powered by [Lip Gloss](https://github.com/charmbracelet/lipgloss) and standard component widgets from [Bubbles](https://github.com/charmbracelet/bubbles).

### Core UX Principles:
1. **Live Dashboard Visibility**: Keep the SkipTUI monitoring dashboard open while spawning terminal shells in external windows or tmux panes.
2. **Persistent Detached Sessions**: Closing SkipTUI does not kill active downloads/browsing sessions unless explicitly requested.
3. **Multi-Protocol Wizard**: Native import for `.ovpn` files, WireGuard `.conf`, and SOCKS5 URIs.
4. **Keyboard-First Navigation**: Full support for Vim keys (`h/j/k/l`), Tab switching, and global hotkeys.

---

## 2. Main Interface Layout & Screen Mockups

### 2.1 Dashboard & Sessions Screen (Main View)

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
│  ● sb-12eb   curl -s api.ipify.org   SG-Socks5      48900   00:00:02   4.5 KB   ● DONE│
│                                                                                       │
├───────────────────────────────────────────────────────────────────────────────────────┤
│ SELECTED SESSION DETAILS (sb-33fc: zsh in Kitty Window)                               │
│ ┌───────────────────────────┐  ┌────────────────────────────────────────────────────┐ │
│ │ Namespace: skiptui-33fc   │  │ Upstream: 185.220.101.5:51820 (WireGuard)          │ │
│ │ Assigned IP: 10.14.0.2    │  │ Latency: 18ms (Loss: 0%)                           │ │
│ │ DNS: 10.14.0.1 (VPN Pushed│  │ Current Speed: ▲ 12.1 KB/s   ▼ 184.2 KB/s          │ │
│ └───────────────────────────┘  └────────────────────────────────────────────────────┘ │
├───────────────────────────────────────────────────────────────────────────────────────┤
│ [l] Launch App   [t] Spawn Terminal   [k] Terminate   [d] Detach   [n] New Profile    │
│ [Tab] Switch Tab [q] Quit             [?] Help Modal                                  │
└───────────────────────────────────────────────────────────────────────────────────────┘
```

---

### 2.2 Profile Manager Screen

```
┌─ SkipTUI: Profiles ──────────────────────────────────────────────────────────────────┐
│  [1] Sessions (3)  │  [2] Profiles (5)  │  [3] Traffic Logs  │  [4] Settings & Health │
├───────────────────────────────────────────────────────────────────────────────────────┤
│ CONFIGURED PROXY & VPN PROFILES                                                       │
│                                                                                       │
│  NAME               TYPE        ENDPOINT                  AUTH        LATENCY  STATUS │
│ > US-Residential    SOCKS5      proxy-us.smartproxy:1080  user:***    45ms     ● OK   │
│   NL-WireGuard      WireGuard   185.220.101.5:51820       PubKey      18ms     ● OK   │
│   Corp-OpenVPN      OpenVPN     vpn.company.com:1194      Cert+Pass   32ms     ● OK   │
│   SG-Shadowsocks    SS          sg01.node.com:8388        aes-256-gcm 165ms    ● OK   │
│   Local-Tor-Socks   SOCKS5      127.0.0.1:9050            None        88ms     ● OK   │
│                                                                                       │
├───────────────────────────────────────────────────────────────────────────────────────┤
│ PROFILE PREVIEW: Corp-OpenVPN                                                         │
│ - Source: ~/.config/skiptui/profiles/corp.ovpn                                        │
│ - Protocol: OpenVPN UDP (AES-256-GCM / SHA256)                                       │
│ - Fail-Closed KillSwitch: Enabled                                                     │
├───────────────────────────────────────────────────────────────────────────────────────┤
│ [Enter] Launch Terminal  [l] Launch Custom App  [i] Import (.ovpn/.conf)  [e] Edit    │
│ [d] Delete Profile       [t] Test Latency       [T] Test All Profiles                 │
└───────────────────────────────────────────────────────────────────────────────────────┘
```

---

### 2.3 Quick Launcher Modal Popup (`[l]` or `[t]` key)

```
┌─────────────────────────── Launch Isolated Process ───────────────────────────┐
│                                                                               │
│  Target Command / Binary:                                                     │
│  [ zsh                                                                      ] │
│                                                                               │
│  Execution Target:                                                            │
│  (•) External Terminal Emulator (Kitty / Alacritty / Tmux)                    │
│  ( ) Background GUI Application (Firefox, Chromium, etc.)                     │
│                                                                               │
│  Select Network Profile:                                                      │
│  (•) Corp-OpenVPN (OpenVPN - 32ms)                                            │
│  ( ) NL-WireGuard (WireGuard - 18ms)                                          │
│  ( ) US-Residential (SOCKS5 - 45ms)                                           │
│                                                                               │
│  Options:                                                                     │
│  [X] Persistent Session (keep running after SkipTUI exit)                     │
│  [X] Strict DNS Leak Killswitch                                               │
│                                                                               │
│             [ <Enter> Launch Session ]       [ <Esc> Cancel ]                 │
└───────────────────────────────────────────────────────────────────────────────┘
```

---

### 2.4 Exit Confirmation & Detach Modal (`[q]` key)

```
┌──────────────────────── Active Sessions Detected ─────────────────────────┐
│                                                                           │
│  There are currently 3 isolated sessions running in the background.       │
│                                                                           │
│  What would you like to do?                                               │
│                                                                           │
│  [D] Detach & Keep Running in Background (Recommended)                    │
│      Sessions and proxy tunnels will continue running. You can reconnect  │
│      by opening SkipTUI again.                                            │
│                                                                           │
│  [K] Terminate All Sessions & Clean Up Namespaces                         │
│      Stops all processes and destroys all temporary network sandboxes.    │
│                                                                           │
│  [Esc] Cancel / Return to Dashboard                                       │
│                                                                           │
└───────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Keyboard Navigation & Hotkey Map

### Global Navigation
| Key | Action |
| :--- | :--- |
| `1` - `4` | Directly jump to Tab (Sessions, Profiles, Logs, Settings) |
| `Tab` / `Shift+Tab` | Cycle through tabs sequentially |
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `?` | Toggle Help & Keybinding cheat-sheet |
| `q` / `Ctrl+C` | Open Exit Dialog (Detach vs Kill All) |

### Sessions View
| Key | Action |
| :--- | :--- |
| `l` | Open Quick Launcher Modal |
| `t` | Quick-spawn isolated terminal shell in new window |
| `k` / `x` | Send SIGTERM / Kill selected isolated session |
| `r` | Restart session |
| `c` | Clear finished/dead sessions from table |

### Profiles View
| Key | Action |
| :--- | :--- |
| `Enter` | Launch isolated terminal shell with selected profile |
| `l` | Launch custom app with selected profile |
| `i` | Import `.ovpn` or WireGuard `.conf` file |
| `a` | Add new proxy profile wizard |
| `e` | Edit selected profile |
| `d` | Delete profile |
| `t` | Run live latency ping test |
| `T` | Test all profiles concurrently |

---

## 4. External Terminal Spawner Integration

When launching an isolated shell:
1. **Terminal Detection**: SkipTUI checks `$TERMINAL`, followed by:
   `kitty`, `alacritty`, `wezterm`, `ghostty`, `foot`, `gnome-terminal`, `xterm`.
2. **Tmux Detection**: If running inside `tmux`, SkipTUI can open a new split pane or window (`tmux split-window "skiptui exec <id>"`).
3. **Execution Command**: Spawns the terminal with `skiptui exec --session <id> -- <shell>`.
4. **Dashboard State**: The SkipTUI main TUI dashboard remains open, live, and updating in real-time.
