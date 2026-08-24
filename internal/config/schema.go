package config

import "time"

// ProtocolType defines the supported tunneling or proxy protocol.
type ProtocolType string

const (
	ProtocolSOCKS5      ProtocolType = "socks5"
	ProtocolHTTP        ProtocolType = "http"
	ProtocolShadowsocks ProtocolType = "shadowsocks"
	ProtocolWireGuard   ProtocolType = "wireguard"
	ProtocolOpenVPN     ProtocolType = "openvpn"
)

// Profile represents a network proxy or VPN endpoint configuration.
type Profile struct {
	ID          string         `json:"id" yaml:"id"`
	Name        string         `json:"name" yaml:"name"`
	Protocol    ProtocolType   `json:"protocol" yaml:"protocol"`
	Endpoint    string         `json:"endpoint" yaml:"endpoint"` // e.g. "1.2.3.4:1080"
	Username    string         `json:"username,omitempty" yaml:"username,omitempty"`
	Password    string         `json:"password,omitempty" yaml:"password,omitempty"`
	DNS         string         `json:"dns,omitempty" yaml:"dns,omitempty"` // e.g. "1.1.1.1"
	KillSwitch  bool           `json:"kill_switch" yaml:"kill_switch"`     // fail-closed if tunnel drops
	WireGuard   *WGConfig      `json:"wireguard,omitempty" yaml:"wireguard,omitempty"`
	OpenVPN     *OpenVPNConfig `json:"openvpn,omitempty" yaml:"openvpn,omitempty"`
	LatencyMs   int64          `json:"latency_ms,omitempty" yaml:"latency_ms,omitempty"`
	LastTested  time.Time      `json:"last_tested,omitempty" yaml:"last_tested,omitempty"`
	CreatedAt   time.Time      `json:"created_at" yaml:"created_at"`
}

// WGConfig holds WireGuard interface and peer configuration.
type WGConfig struct {
	PrivateKey   string   `json:"private_key" yaml:"private_key"`
	PublicKey    string   `json:"public_key" yaml:"public_key"`
	PresharedKey string   `json:"preshared_key,omitempty" yaml:"preshared_key,omitempty"`
	Address      string   `json:"address" yaml:"address"`             // e.g. "10.14.0.2/32"
	AllowedIPs   []string `json:"allowed_ips" yaml:"allowed_ips"`     // e.g. ["0.0.0.0/0"]
	ConfigPath   string   `json:"config_path,omitempty" yaml:"config_path,omitempty"`
}

// OpenVPNConfig holds OpenVPN configuration and credentials.
type OpenVPNConfig struct {
	ConfigPath     string `json:"config_path" yaml:"config_path"` // Path to .ovpn file
	AuthUserPass   string `json:"auth_user_pass,omitempty" yaml:"auth_user_pass,omitempty"`
	Proto          string `json:"proto,omitempty" yaml:"proto,omitempty"` // "udp" or "tcp"
	RemoteHost     string `json:"remote_host,omitempty" yaml:"remote_host,omitempty"`
	RemotePort     string `json:"remote_port,omitempty" yaml:"remote_port,omitempty"`
	InlineCertData string `json:"inline_cert_data,omitempty" yaml:"inline_cert_data,omitempty"`
}

// Settings contains global application preferences.
type Settings struct {
	DefaultProfile  string `json:"default_profile" yaml:"default_profile"`
	RootlessMode    bool   `json:"rootless_mode" yaml:"rootless_mode"`
	DNSFallback     string `json:"dns_fallback" yaml:"dns_fallback"`
	PreferredTerm   string `json:"preferred_terminal" yaml:"preferred_terminal"` // e.g. "kitty", "alacritty", "auto"
	Theme           string `json:"theme" yaml:"theme"`
	LogLevel        string `json:"log_level" yaml:"log_level"`
	FailClosedAll   bool   `json:"fail_closed_all" yaml:"fail_closed_all"`
}

// Config is the root configuration structure.
type Config struct {
	Version  int        `json:"version" yaml:"version"`
	Settings Settings   `json:"settings" yaml:"settings"`
	Profiles []*Profile `json:"profiles" yaml:"profiles"`
}

// SessionInfo represents an active or past isolated sandbox session.
type SessionInfo struct {
	ID          string    `json:"id"`
	Command     string    `json:"command"`
	Args        []string  `json:"args"`
	ProfileID   string    `json:"profile_id"`
	ProfileName string    `json:"profile_name"`
	Protocol    string    `json:"protocol"`
	PID         int       `json:"pid"`
	Status      string    `json:"status"` // "running", "stopped", "failed"
	Namespace   string    `json:"namespace"`
	IPAddress   string    `json:"ip_address"`
	DNS         string    `json:"dns"`
	StartTime   time.Time `json:"start_time"`
	BytesRX     uint64    `json:"bytes_rx"`
	BytesTX     uint64    `json:"bytes_tx"`
	Detached    bool      `json:"detached"`
	LogFilePath string    `json:"log_file_path"`
}
