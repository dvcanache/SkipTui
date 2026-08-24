package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

var (
	defaultConfigMutex sync.RWMutex
	globalConfig       *Config
)

// GetConfigDir returns the XDG config directory for SkipTUI (~/.config/skiptui).
func GetConfigDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "/tmp/skiptui-config"
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "skiptui")
}

// GetProfilesDir returns the directory where imported profiles and configs are saved (~/.config/skiptui/profiles).
func GetProfilesDir() string {
	return filepath.Join(GetConfigDir(), "profiles")
}

// GetRuntimeDir returns the runtime directory for sockets and PID files (/run/user/<UID>/skiptui).
func GetRuntimeDir() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		uid := os.Getuid()
		runtimeDir = fmt.Sprintf("/tmp/skiptui-runtime-%d", uid)
	} else {
		runtimeDir = filepath.Join(runtimeDir, "skiptui")
	}
	_ = os.MkdirAll(runtimeDir, 0700)
	return runtimeDir
}

// GetSocketPath returns the path to the Unix Domain Socket for daemon IPC.
func GetSocketPath() string {
	return filepath.Join(GetRuntimeDir(), "skiptui.sock")
}

// InitDirs ensures that configuration and runtime directories exist with safe permissions.
func InitDirs() error {
	cfgDir := GetConfigDir()
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		return fmt.Errorf("failed to create config dir %s: %w", cfgDir, err)
	}

	profDir := GetProfilesDir()
	if err := os.MkdirAll(profDir, 0700); err != nil {
		return fmt.Errorf("failed to create profiles dir %s: %w", profDir, err)
	}

	runDir := GetRuntimeDir()
	if err := os.MkdirAll(runDir, 0700); err != nil {
		return fmt.Errorf("failed to create runtime dir %s: %w", runDir, err)
	}

	return nil
}

// LoadConfig reads the configuration file from ~/.config/skiptui/config.yaml.
func LoadConfig() (*Config, error) {
	defaultConfigMutex.Lock()
	defer defaultConfigMutex.Unlock()

	if err := InitDirs(); err != nil {
		return nil, err
	}

	configFile := filepath.Join(GetConfigDir(), "config.yaml")

	// If config file does not exist, initialize with default templates
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := getDefaultConfig()
		if err := saveConfigUnlocked(cfg); err != nil {
			return nil, fmt.Errorf("failed to write initial config: %w", err)
		}
		globalConfig = cfg
		return cfg, nil
	}

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Ensure profiles list is initialized
	if cfg.Profiles == nil {
		cfg.Profiles = make([]*Profile, 0)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// SaveConfig persists the current configuration to disk.
func SaveConfig(cfg *Config) error {
	defaultConfigMutex.Lock()
	defer defaultConfigMutex.Unlock()
	return saveConfigUnlocked(cfg)
}

func saveConfigUnlocked(cfg *Config) error {
	configFile := filepath.Join(GetConfigDir(), "config.yaml")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", configFile, err)
	}

	globalConfig = cfg
	return nil
}

// GetProfile finds a profile by ID or Name.
func (c *Config) GetProfile(idOrName string) *Profile {
	for _, p := range c.Profiles {
		if p.ID == idOrName || p.Name == idOrName {
			return p
		}
	}
	return nil
}

// AddProfile inserts or updates a profile in the configuration.
func (c *Config) AddProfile(p *Profile) error {
	if p.ID == "" {
		p.ID = "p-" + uuid.New().String()[:8]
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}

	for i, existing := range c.Profiles {
		if existing.ID == p.ID || existing.Name == p.Name {
			c.Profiles[i] = p
			return SaveConfig(c)
		}
	}

	c.Profiles = append(c.Profiles, p)
	return SaveConfig(c)
}

// DeleteProfile removes a profile from the configuration.
func (c *Config) DeleteProfile(idOrName string) error {
	newProfiles := make([]*Profile, 0, len(c.Profiles))
	found := false
	for _, p := range c.Profiles {
		if p.ID == idOrName || p.Name == idOrName {
			found = true
			continue
		}
		newProfiles = append(newProfiles, p)
	}

	if !found {
		return fmt.Errorf("profile '%s' not found", idOrName)
	}

	c.Profiles = newProfiles
	return SaveConfig(c)
}

// ImportFile imports an .ovpn or .conf file and copies it to the profiles directory.
func (c *Config) ImportFile(srcPath string, customName string) (*Profile, error) {
	if err := InitDirs(); err != nil {
		return nil, err
	}

	ext := stringsToLowerExt(srcPath)
	var profile *Profile
	var err error

	destFilename := filepath.Base(srcPath)
	destPath := filepath.Join(GetProfilesDir(), destFilename)

	// Copy file into ~/.config/skiptui/profiles/
	if err := copyFile(srcPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy profile file: %w", err)
	}

	switch ext {
	case ".ovpn":
		profile, err = ParseOVPN(destPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse OpenVPN file: %w", err)
		}
	case ".conf":
		profile, err = ParseWireGuard(destPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse WireGuard file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported profile extension '%s' (expected .ovpn or .conf)", ext)
	}

	if customName != "" {
		profile.Name = customName
	}

	if err := c.AddProfile(profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func stringsToLowerExt(path string) string {
	ext := filepath.Ext(path)
	return filepath.Clean(ext)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func getDefaultConfig() *Config {
	return &Config{
		Version: 1,
		Settings: Settings{
			DefaultProfile: "local-tor",
			RootlessMode:   false,
			DNSFallback:    "1.1.1.1",
			PreferredTerm:  "auto",
			Theme:          "nord",
			LogLevel:       "info",
			FailClosedAll:  true,
		},
		Profiles: []*Profile{
			{
				ID:         "p-local-tor",
				Name:       "Local-Tor-Socks",
				Protocol:   ProtocolSOCKS5,
				Endpoint:   "127.0.0.1:9050",
				DNS:        "1.1.1.1",
				KillSwitch: true,
				CreatedAt:  time.Now(),
			},
			{
				ID:         "p-sample-socks5",
				Name:       "Sample-SOCKS5",
				Protocol:   ProtocolSOCKS5,
				Endpoint:   "198.51.100.1:1080",
				Username:   "user",
				Password:   "password",
				DNS:        "1.1.1.1",
				KillSwitch: true,
				CreatedAt:  time.Now(),
			},
			{
				ID:         "p-sample-http",
				Name:       "Sample-HTTP-Proxy",
				Protocol:   ProtocolHTTP,
				Endpoint:   "198.51.100.2:8080",
				DNS:        "1.1.1.1",
				KillSwitch: true,
				CreatedAt:  time.Now(),
			},
		},
	}
}
