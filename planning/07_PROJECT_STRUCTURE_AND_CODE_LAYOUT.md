# SkipTUI: Project Structure & Code Layout

## 1. Directory Tree

SkipTUI adheres to standard Go project conventions ([golang-standards/project-layout](https://github.com/golang-standards/project-layout)):

```
skiptui/
├── cmd/
│   └── skiptui/
│       ├── main.go                 # Application entrypoint
│       ├── root.go                 # Cobra root command & TUI launcher
│       ├── run.go                  # CLI command: skiptui run --profile <p> -- <cmd>
│       ├── list.go                 # CLI command: skiptui list
│       ├── kill.go                 # CLI command: skiptui kill <id>
│       ├── import.go               # CLI command: skiptui import <file.ovpn|conf>
│       ├── test.go                 # CLI command: skiptui test [profile]
│       └── exec.go                 # Internal worker entrypoint for spawned terminals
├── internal/
│   ├── app/
│   │   └── app.go                  # Core application coordinator & client
│   ├── daemon/
│   │   ├── server.go               # Unix domain socket server (/run/user/<UID>/skiptui/skiptui.sock)
│   │   ├── client.go               # IPC client communicating with daemon
│   │   └── state.go                # Active session registry & state serialization
│   ├── config/
│   │   ├── config.go               # Viper configuration loader (XDG standard)
│   │   ├── schema.go               # Config struct models (Profile, Settings, Session)
│   │   ├── ovpn_parser.go          # Native .ovpn profile & certificate parser
│   │   └── wg_parser.go            # WireGuard .conf profile parser
│   ├── isolation/
│   │   ├── engine.go               # IsolationEngine interface
│   │   ├── netns.go                # Linux NetNS implementation (vishvananda/netns)
│   │   ├── rootless.go             # Rootless UserNS + slirp4netns implementation
│   │   ├── dns.go                  # Isolated DNS / resolv.conf injector
│   │   └── cleanup.go              # Stale namespace detection & signal-based GC
│   ├── tunnel/
│   │   ├── tunnel.go               # TunnelEngine interface
│   │   ├── tun2socks.go            # Tun2Socks Layer-3 to Layer-5 adapter
│   │   ├── wireguard.go            # WireGuard NetNS interface migration adapter
│   │   ├── openvpn.go              # OpenVPN isolated namespace adapter
│   │   └── tun_device.go           # Linux TUN device allocator (/dev/net/tun)
│   ├── terminal/
│   │   ├── spawner.go              # External terminal emulator launcher ($TERMINAL/kitty/alacritty/tmux)
│   │   └── detector.go             # Available terminal & tmux environment detector
│   ├── session/
│   │   ├── supervisor.go           # Session registry & active sandbox tracker
│   │   ├── process.go              # Child process runner, PTY allocation, stdio piping
│   │   └── stats.go                # RX/TX byte counters and PID metrics collector
│   ├── netprobe/
│   │   ├── latency.go              # Profile ping & RTT latency prober
│   │   └── health.go               # Killswitch health check watchdog
│   └── tui/
│       ├── tui.go                  # Bubble Tea main program wrapper
│       ├── model.go                # Root Bubble Tea Model (Elm architecture)
│       ├── update.go               # Global keybindings & message dispatcher
│       ├── view.go                 # Top-level UI layout & tab orchestrator
│       ├── styles/
│       │   ├── theme.go            # Lip Gloss colors, borders, and styles
│       │   └── icons.go            # Nerd Font / Unicode icon definitions
│       ├── views/
│       │   ├── dashboard.go        # Active sessions table & summary view
│       │   ├── profiles.go         # Profile list & latency manager view
│       │   ├── logs.go             # Real-time stdout/stderr log stream view
│       │   └── settings.go         # System capabilities & config view
│       └── modals/
│           ├── launcher.go         # Quick-launch app popup modal
│           ├── profile_form.go     # Add/Edit profile interactive form
│           ├── import_wizard.go    # File picker modal for .ovpn and .conf files
│           ├── exit_dialog.go      # Detach vs Kill confirmation dialog
│           └── help.go             # Keyboard shortcut cheat-sheet modal
├── pkg/
│   └── version/
│       └── version.go              # Build version, git commit, and release metadata
├── planning/                       # Architectural & Planning documentation
│   ├── 00_OVERVIEW_AND_GOALS.md
│   ├── 01_ARCHITECTURE_AND_DESIGN.md
│   ├── 02_NETWORKING_AND_ISOLATION_STRATEGIES.md
│   ├── 03_TUI_UX_AND_WORKFLOW.md
│   ├── 04_SECURITY_AND_PERMISSIONS.md
│   ├── 05_TECH_STACK_AND_DEPENDENCIES.md
│   ├── 06_IMPLEMENTATION_ROADMAP.md
│   └── 07_PROJECT_STRUCTURE_AND_CODE_LAYOUT.md
├── scripts/
│   ├── setup_caps.sh               # Helper script to set Linux capabilities on binary
│   └── test_netns.sh               # Integration test helper for network namespaces
├── Makefile                        # Build, test, lint, and run targets
├── go.mod                          # Go module definition
└── README.md                       # Project landing page & quick start guide
```

---

## 2. Key Go Interfaces & Struct Definitions

### 2.1 Profile Configuration Schema (`internal/config/schema.go`)
```go
package config

type ProtocolType string

const (
    ProtocolSOCKS5      ProtocolType = "socks5"
    ProtocolHTTP        ProtocolType = "http"
    ProtocolShadowsocks ProtocolType = "shadowsocks"
    ProtocolWireGuard   ProtocolType = "wireguard"
    ProtocolOpenVPN     ProtocolType = "openvpn"
)

type Profile struct {
    ID          string          `json:"id" yaml:"id"`
    Name        string          `json:"name" yaml:"name"`
    Protocol    ProtocolType    `json:"protocol" yaml:"protocol"`
    Endpoint    string          `json:"endpoint" yaml:"endpoint"`       // "1.2.3.4:1080"
    Username    string          `json:"username,omitempty" yaml:"username,omitempty"`
    Password    string          `json:"password,omitempty" yaml:"password,omitempty"`
    DNS         string          `json:"dns" yaml:"dns"`                 // e.g. "1.1.1.1"
    KillSwitch  bool            `json:"kill_switch" yaml:"kill_switch"` // fail-closed
    WireGuard   *WGConfig       `json:"wireguard,omitempty" yaml:"wireguard,omitempty"`
    OpenVPN     *OpenVPNConfig  `json:"openvpn,omitempty" yaml:"openvpn,omitempty"`
}

type WGConfig struct {
    PrivateKey string   `json:"private_key" yaml:"private_key"`
    PublicKey  string   `json:"public_key" yaml:"public_key"`
    Address    string   `json:"address" yaml:"address"`       // "10.14.0.2/32"
    AllowedIPs []string `json:"allowed_ips" yaml:"allowed_ips"` // ["0.0.0.0/0"]
}

type OpenVPNConfig struct {
    ConfigPath     string `json:"config_path" yaml:"config_path"`         // ~/.config/skiptui/profiles/corp.ovpn
    AuthUserPass   string `json:"auth_user_pass,omitempty" yaml:"auth_user_pass,omitempty"`
    InlineCertData string `json:"inline_cert_data,omitempty" yaml:"inline_cert_data,omitempty"`
}
```

### 2.2 Isolation Engine Interface (`internal/isolation/engine.go`)
```go
package isolation

import (
    "context"
    "os/exec"
    "skiptui/internal/config"
)

type SandboxInfo struct {
    ID          string
    Namespace   string
    TunDevice   string
    IPAddress   string
    DNS         string
    CreatedAt   int64
}

type IsolationEngine interface {
    // Initialize a new isolated network sandbox with routing
    CreateSandbox(ctx context.Context, id string, profile *config.Profile) (*SandboxInfo, error)
    
    // Execute a command inside the isolated sandbox
    BuildCommand(ctx context.Context, sb *SandboxInfo, targetCmd string, args ...string) *exec.Cmd
    
    // Teardown the sandbox, remove virtual devices, and release namespace
    DestroySandbox(ctx context.Context, sb *SandboxInfo) error
}
```

### 2.3 External Terminal Spawner Interface (`internal/terminal/spawner.go`)
```go
package terminal

import (
    "context"
)

type TerminalType string

const (
    TermKitty     TerminalType = "kitty"
    TermAlacritty TerminalType = "alacritty"
    TermWezterm   TerminalType = "wezterm"
    TermTmux      TerminalType = "tmux"
    TermGeneric   TerminalType = "generic"
)

type Spawner interface {
    DetectPreferredTerminal() TerminalType
    SpawnSessionInTerminal(ctx context.Context, sessionID string, cmd string, args []string) error
}
```

---

## 3. Configuration Directory Layout (XDG Standard)

```
~/.config/skiptui/
├── config.yaml                     # Global settings
└── profiles/                       # Profile storage
    ├── work-corp.ovpn              # Imported OpenVPN profile
    ├── proton-nl.conf              # Imported WireGuard profile
    └── proxies.json                # SOCKS5/HTTP profile definitions

/run/user/1000/skiptui/             # Runtime volatile files (RAM/tmpfs)
├── skiptui.sock                    # Unix domain socket for daemon IPC
└── sessions.json                   # Active session state metadata
```
