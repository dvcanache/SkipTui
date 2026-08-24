package isolation

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"skiptui/internal/config"
	"time"
)

// RootlessEngine implements Engine using unprivileged Linux User Namespaces.
type RootlessEngine struct{}

func NewRootlessEngine() *RootlessEngine {
	return &RootlessEngine{}
}

func (r *RootlessEngine) Name() string {
	return "rootless"
}

// CheckCapabilities checks if unprivileged user namespaces are enabled on the system.
func (r *RootlessEngine) CheckCapabilities() error {
	cmd := exec.Command("unshare", "--user", "--net", "--map-root-user", "true")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("user namespaces not permitted: %s\nRun with 'sudo' or grant capabilities using 'sudo make setcap'", string(out))
	}
	return nil
}

// CreateSandbox registers a rootless sandbox.
func (r *RootlessEngine) CreateSandbox(ctx context.Context, id string, profile *config.Profile) (*SandboxInfo, error) {
	if err := r.CheckCapabilities(); err != nil {
		return nil, err
	}

	dns := "1.1.1.1"
	if profile != nil && profile.DNS != "" {
		dns = profile.DNS
	}

	sb := &SandboxInfo{
		ID:         id,
		Namespace:  "rootless-" + id,
		DNS:        dns,
		IsRootless: true,
		CreatedAt:  time.Now().Unix(),
	}

	return sb, nil
}

// BuildCommand creates an unshare command that enters a fresh unprivileged user & network namespace.
func (r *RootlessEngine) BuildCommand(ctx context.Context, sb *SandboxInfo, targetCmd string, args ...string) (*exec.Cmd, error) {
	var cmdArgs []string
	cmdArgs = append(cmdArgs, "--user", "--net", "--map-root-user", "--", targetCmd)
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, "unshare", cmdArgs...)
	cmd.Env = SanitizeProcessEnv(os.Environ())

	return cmd, nil
}

// DestroySandbox cleans up rootless sandbox.
func (r *RootlessEngine) DestroySandbox(ctx context.Context, sb *SandboxInfo) error {
	return nil
}
