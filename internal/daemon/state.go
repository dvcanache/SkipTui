package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"skiptui/internal/config"
	"sync"
	"syscall"
)

var stateMutex sync.Mutex

// SaveSessionState writes the active session list to /run/user/<UID>/skiptui/sessions.json.
func SaveSessionState(sessions []*config.SessionInfo) error {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	stateFile := filepath.Join(config.GetRuntimeDir(), "sessions.json")
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session state: %w", err)
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

// GetSessionByID retrieves a session by its ID from disk state.
func GetSessionByID(id string) (*config.SessionInfo, error) {
	sessions, err := LoadSessionState()
	if err != nil {
		return nil, err
	}
	for _, s := range sessions {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, fmt.Errorf("session '%s' not found", id)
}

// UpdateSession updates or appends a session in the persistent state file.
func UpdateSession(sess *config.SessionInfo) error {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	stateFile := filepath.Join(config.GetRuntimeDir(), "sessions.json")
	var sessions []*config.SessionInfo
	if data, err := os.ReadFile(stateFile); err == nil {
		_ = json.Unmarshal(data, &sessions)
	}

	found := false
	for i, s := range sessions {
		if s.ID == sess.ID {
			sessions[i] = sess
			found = true
			break
		}
	}
	if !found {
		sessions = append(sessions, sess)
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile, data, 0600)
}

// RemoveSession deletes a session by ID from persistent state.
func RemoveSession(id string) error {
	stateMutex.Lock()
	defer stateMutex.Unlock()

	stateFile := filepath.Join(config.GetRuntimeDir(), "sessions.json")
	var sessions []*config.SessionInfo
	if data, err := os.ReadFile(stateFile); err == nil {
		_ = json.Unmarshal(data, &sessions)
	}

	newSessions := make([]*config.SessionInfo, 0, len(sessions))
	for _, s := range sessions {
		if s.ID != id {
			newSessions = append(newSessions, s)
		}
	}

	data, err := json.MarshalIndent(newSessions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile, data, 0600)
}

// IsPIDAlive checks if a process with the given PID is currently active.
func IsPIDAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}
