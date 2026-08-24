package isolation

import (
	"context"
	"os/exec"
	"skiptui/internal/config"
)

// SandboxInfo holds runtime metadata about an active isolated network sandbox.
type SandboxInfo struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	TunDevice string `json:"tun_device,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	DNS       string `json:"dns,omitempty"`
	IsRootless bool   `json:"is_rootless"`
	CreatedAt int64  `json:"created_at"`
}

// Engine defines the contract for initializing, executing commands within, and destroying sandboxes.
type Engine interface {
	// Name returns the identifier of this isolation engine (e.g. "netns", "rootless")
	Name() string

	// CreateSandbox provisions an isolated network sandbox environment.
	CreateSandbox(ctx context.Context, id string, profile *config.Profile) (*SandboxInfo, error)

	// BuildCommand prepares an exec.Cmd to run a binary inside the sandbox.
	BuildCommand(ctx context.Context, sb *SandboxInfo, targetCmd string, args ...string) (*exec.Cmd, error)

	// DestroySandbox tears down the network sandbox, removing interfaces and releasing namespaces.
	DestroySandbox(ctx context.Context, sb *SandboxInfo) error

	// CheckCapabilities verifies if the host system and current user permissions support this engine.
	CheckCapabilities() error
}
