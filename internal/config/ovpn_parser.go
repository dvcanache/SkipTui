package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ParseOVPN parses an OpenVPN configuration file and extracts critical routing/endpoint metadata.
func ParseOVPN(path string) (*Profile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open ovpn file: %w", err)
	}
	defer file.Close()

	profile := &Profile{
		ID:         "ovpn-" + uuid.New().String()[:8],
		Name:       strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Protocol:   ProtocolOpenVPN,
		KillSwitch: true,
		CreatedAt:  time.Now(),
		OpenVPN: &OpenVPNConfig{
			ConfigPath: path,
			Proto:      "udp",
			RemotePort: "1194",
		},
	}

	scanner := bufio.NewScanner(file)
	var dnsServers []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		key := strings.ToLower(fields[0])

		switch key {
		case "remote":
			if len(fields) >= 2 {
				profile.OpenVPN.RemoteHost = fields[1]
				if len(fields) >= 3 {
					profile.OpenVPN.RemotePort = fields[2]
				}
				if len(fields) >= 4 {
					profile.OpenVPN.Proto = strings.ToLower(fields[3])
				}
				profile.Endpoint = fmt.Sprintf("%s:%s", profile.OpenVPN.RemoteHost, profile.OpenVPN.RemotePort)
			}
		case "proto":
			if len(fields) >= 2 {
				profile.OpenVPN.Proto = strings.ToLower(fields[1])
			}
		case "port":
			if len(fields) >= 2 {
				profile.OpenVPN.RemotePort = fields[1]
				if profile.OpenVPN.RemoteHost != "" {
					profile.Endpoint = fmt.Sprintf("%s:%s", profile.OpenVPN.RemoteHost, profile.OpenVPN.RemotePort)
				}
			}
		case "dhcp-option":
			if len(fields) >= 3 && strings.ToUpper(fields[1]) == "DNS" {
				dnsServers = append(dnsServers, fields[2])
			}
		case "auth-user-pass":
			if len(fields) >= 2 {
				profile.OpenVPN.AuthUserPass = fields[1]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading ovpn file: %w", err)
	}

	if len(dnsServers) > 0 {
		profile.DNS = strings.Join(dnsServers, ", ")
	} else {
		profile.DNS = "1.1.1.1" // Fallback safe DNS
	}

	if profile.Endpoint == "" {
		profile.Endpoint = filepath.Base(path)
	}

	return profile, nil
}
