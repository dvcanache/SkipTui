package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"skiptui/internal/config"
	"sync"
)

var stateMutex sync.Mutex

// SaveSessionState writes the active session list to /run/user/<UID>/skiptui/sessions.json.
func SaveSessionState(sessions []*config.SessionInfo) error {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	stateFile := filepath.Join(config.GetRuntimeDir(), "sessions.json")
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0600)
}

// LoadSessionState reads saved session state from /run/user/<UID>/skiptui/sessions.json.
func LoadSessionState() ([]*config.SessionInfo, error) {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	stateFile := filepath.Join(config.GetRuntimeDir(), "sessions.json")
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return []*config.SessionInfo{}, nil
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	var sessions []*config.SessionInfo
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}
