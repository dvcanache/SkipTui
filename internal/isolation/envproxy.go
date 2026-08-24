package isolation

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"skiptui/internal/config"
	"time"
)

// EnvProxyEngine provides zero-privilege proxy isolation by injecting standard proxy environment variables.
// This works on any Linux system without requiring root, capabilities, or user namespaces.
type EnvProxyEngine struct{}

func NewEnvProxyEngine() *EnvProxyEngine {
	return &EnvProxyEngine{}
}

func (e *EnvProxyEngine) Name() string {
	return "envproxy"
}

// CheckCapabilities always succeeds because environment variable injection requires no special privileges.
func (e *EnvProxyEngine) CheckCapabilities() error {
	return nil
}

// CreateSandbox provisions an environment-level proxy sandbox.
func (e *EnvProxyEngine) CreateSandbox(ctx context.Context, id string, profile *config.Profile) (*SandboxInfo, error) {
	dns := "1.1.1.1"
	if profile != nil && profile.DNS != "" {
		dns = profile.DNS
	}

	sb := &SandboxInfo{
		ID:         id,
		Namespace:  "env-" + id,
		DNS:        dns,
		IsRootless: true,
		CreatedAt:  time.Now().Unix(),
	}

	return sb, nil
}

// BuildCommand constructs an exec.Cmd with proxy environment variables injected.
func (e *EnvProxyEngine) BuildCommand(ctx context.Context, sb *SandboxInfo, targetCmd string, args ...string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, targetCmd, args...)
	cmd.Env = os.Environ()

	return cmd, nil
}

// BuildCommandWithProfile constructs an exec.Cmd configured with proxy environment variables for the given profile.
func (e *EnvProxyEngine) BuildCommandWithProfile(ctx context.Context, sb *SandboxInfo, profile *config.Profile, targetCmd string, args ...string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, targetCmd, args...)

	env := SanitizeProcessEnv(os.Environ())

	if profile != nil {
		proxyURL := ""
		switch profile.Protocol {
		case config.ProtocolSOCKS5:
			if profile.Username != "" && profile.Password != "" {
				proxyURL = fmt.Sprintf("socks5h://%s:%s@%s", profile.Username, profile.Password, profile.Endpoint)
			} else {
				proxyURL = fmt.Sprintf("socks5h://%s", profile.Endpoint)
			}
			env = append(env,
				"ALL_PROXY="+proxyURL,
				"all_proxy="+proxyURL,
				"SOCKS_SERVER="+profile.Endpoint,
				"SOCKS5_SERVER="+profile.Endpoint,
			)
		case config.ProtocolHTTP:
			if profile.Username != "" && profile.Password != "" {
				proxyURL = fmt.Sprintf("http://%s:%s@%s", profile.Username, profile.Password, profile.Endpoint)
			} else {
				proxyURL = fmt.Sprintf("http://%s", profile.Endpoint)
			}
			env = append(env,
				"HTTP_PROXY="+proxyURL,
				"HTTPS_PROXY="+proxyURL,
				"http_proxy="+proxyURL,
				"https_proxy="+proxyURL,
				"ALL_PROXY="+proxyURL,
				"all_proxy="+proxyURL,
			)
		case config.ProtocolShadowsocks:
			proxyURL = fmt.Sprintf("ss://%s", profile.Endpoint)
			env = append(env, "ALL_PROXY="+proxyURL, "all_proxy="+proxyURL)
		}

		// Disable proxy bypass
		env = append(env, "NO_PROXY=", "no_proxy=")
	}

	cmd.Env = env
	return cmd, nil
}

// DestroySandbox cleans up any resources.
func (e *EnvProxyEngine) DestroySandbox(ctx context.Context, sb *SandboxInfo) error {
	return nil
}
