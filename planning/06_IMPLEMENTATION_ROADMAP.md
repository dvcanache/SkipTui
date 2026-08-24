# SkipTUI: Implementation Roadmap & Milestones

## 1. Roadmap Overview

```mermaid
gantt
    title SkipTUI Implementation Roadmap
    dateFormat  YYYY-MM-DD
    section Phase 1: Foundation
    Scaffolding & Go Modules          :p1_1, 2026-09-01, 3d
    Core NetNS Isolation Engine       :p1_2, after p1_1, 5d
    Process Execution & Signal Trap   :p1_3, after p1_2, 4d
    section Phase 2: Tunneling & VPN
    TUN Device & Tun2Socks Core       :p2_1, after p1_3, 6d
    WireGuard & OpenVPN Engines       :p2_2, after p2_1, 5d
    XDG Profile Store & Parser        :p2_3, after p2_2, 4d
    DNS Isolation & Leak Killer       :p2_4, after p2_3, 3d
    section Phase 3: Daemon & IPC
    Unix Socket Daemon & Supervisor   :p3_1, after p2_4, 5d
    Detachable Session State          :p3_2, after p3_1, 3d
    Concurrent Latency Prober         :p3_3, after p3_2, 3d
    section Phase 4: Bubble Tea TUI
    Dashboard & Session Table         :p4_1, after p3_3, 5d
    Profile Manager & Import Wizard   :p4_2, after p4_1, 4d
    Terminal Spawner & Launch Modal   :p4_3, after p4_2, 4d
    section Phase 5: CLI & Hardening
    Headless CLI Parity (Cobra)       :p5_1, after p4_3, 4d
    Rootless UserNS Engine            :p5_2, after p5_1, 5d
    Packaging & Release Automation    :p5_3, after p5_2, 3d
```

---

## 2. Milestone Breakdown & Deliverables

### Phase 1: Core Foundation & Linux Namespace Engine
- [ ] Initialize Go repository, `go.mod`, directory layout, and CI linters.
- [ ] Implement `internal/isolation/netns`:
  - Create isolated network namespace (`ip netns add ...` via netlink).
  - Bring up `lo` interface.
  - Spawn target process inside namespace (`unix.Setns`).
  - Cleanup namespace on process exit.
- [ ] Implement robust Signal Handler (`SIGINT`, `SIGTERM`, `SIGHUP`) and stale namespace sweeper.
- **Acceptance Criteria**: Running a test binary spawns a process with an isolated `lo` interface and no access to host `eth0` / WAN.

### Phase 2: Tunneling Engines (Tun2Socks, WireGuard, OpenVPN)
- [ ] Integrate embedded `tun2socks` engine:
  - Allocate TUN interface `tun0` inside namespace.
  - Connect TUN raw packets to upstream SOCKS5 / HTTP proxy.
- [ ] Implement WireGuard netns migration adapter (`wgctrl`).
- [ ] Implement OpenVPN parser (`internal/config/ovpn_parser.go`) and isolated OpenVPN worker.
- [ ] Implement isolated DNS resolver (`resolv.conf` binding & D-Bus blocking).
- **Acceptance Criteria**: Running `curl ifconfig.me` inside SOCKS5, WireGuard, or OpenVPN sandbox returns the proxy/VPN IP while host IP remains unchanged.

### Phase 3: Daemon, IPC & Session Persistence
- [ ] Implement Unix Domain Socket Server (`/run/user/<UID>/skiptui/skiptui.sock`).
- [ ] Build Session Supervisor to track running sandboxes, PIDs, and RX/TX traffic.
- [ ] Implement session persistence across frontend disconnects.
- [ ] Implement concurrent latency prober (`internal/netprobe`).
- **Acceptance Criteria**: Closing the CLI or frontend leaves sandboxed workloads running; reconnecting restores live session state.

### Phase 4: Bubble Tea TUI & External Terminal Spawner
- [ ] Implement `internal/tui/app.go` (Main Elm architecture loop).
- [ ] Build **Sessions View**: live table of active sandboxes, PIDs, RX/TX traffic, and status badges.
- [ ] Build **Profiles View**: list profiles, run live latency tests, and view details.
- [ ] Build **Import Wizard**: file picker modal for `.ovpn` and `.conf` files.
- [ ] Implement **External Terminal Spawner** (`internal/terminal/spawner.go`): detects `$TERMINAL`, `kitty`, `alacritty`, `wezterm`, or `tmux` and spawns isolated shell sessions in separate windows.
- [ ] Implement **Exit Confirmation Dialog**: prompt to Detach vs Terminate All.
- **Acceptance Criteria**: User can browse profiles, test latency, launch terminal shells in external windows, and monitor bandwidth in real-time.

### Phase 5: CLI Parity, Rootless Mode & Packaging
- [ ] Implement full Cobra CLI subcommands:
  - `skiptui run --profile <name> [--terminal] -- <command>`
  - `skiptui list`, `skiptui kill <id>`, `skiptui import <file>`, `skiptui test`
- [ ] Implement `internal/isolation/rootless` via User Namespaces (`CLONE_NEWUSER`) + `slirp4netns` or gVisor netstack.
- [ ] Configure `GoReleaser` for building binaries, `.deb`, `.rpm`, and Arch Linux PKGBUILDs.
- **Acceptance Criteria**: Complete CLI and TUI feature parity with 100% test coverage for core namespace operations.
