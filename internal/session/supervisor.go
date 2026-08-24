package session

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"skiptui/internal/config"
	"skiptui/internal/isolation"
	"skiptui/internal/terminal"
	"skiptui/internal/tunnel"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// ActiveSession wraps the runtime process, sandbox, and tunnel worker for a session.
type ActiveSession struct {
	Info     *config.SessionInfo
	Sandbox  *isolation.SandboxInfo
	Tunnel   tunnel.TunnelInstance
	Cmd      *exec.Cmd
	Cancel   context.CancelFunc
	Engine   isolation.Engine
	mu       sync.RWMutex
}

// Supervisor coordinates active sandboxed sessions and lifecycles.
type Supervisor struct {
	sessions map[string]*ActiveSession
	mu       sync.RWMutex
	cfg      *config.Config
	netnsEng *isolation.NetnsEngine
	rootless *isolation.RootlessEngine
	envProxy *isolation.EnvProxyEngine
}

func NewSupervisor(cfg *config.Config) *Supervisor {
	return &Supervisor{
		sessions: make(map[string]*ActiveSession),
		cfg:      cfg,
		netnsEng: isolation.NewNetnsEngine(),
		rootless: isolation.NewRootlessEngine(),
		envProxy: isolation.NewEnvProxyEngine(),
	}
}

// SelectBestEngine picks the highest-privilege working isolation engine available.
func (s *Supervisor) SelectBestEngine() isolation.Engine {
	if !s.cfg.Settings.RootlessMode && s.netnsEng.CheckCapabilities() == nil {
		return s.netnsEng
	}
	if s.rootless.CheckCapabilities() == nil {
		return s.rootless
	}
	return s.envProxy
}

// LaunchSession starts an isolated sandbox, connects the proxy/VPN tunnel, and runs the target command.
func (s *Supervisor) LaunchSession(ctx context.Context, profile *config.Profile, targetCmd string, args []string, inTerminal bool) (*config.SessionInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := "sb-" + uuid.New().String()[:8]

	// 1. Choose best working isolation engine
	engine := s.SelectBestEngine()

	// 2. Create sandbox (with fallback if engine fails)
	sb, err := engine.CreateSandbox(ctx, sessionID, profile)
	if err != nil {
		// Fallback to EnvProxyEngine if kernel restrictions block namespace creation
		engine = s.envProxy
		sb, err = engine.CreateSandbox(ctx, sessionID, profile)
		if err != nil {
			return nil, fmt.Errorf("failed to create sandbox: %w", err)
		}
	}

	// 3. Initialize tunnel adapter
	tunWorker, err := tunnel.CreateTunnel(sb, profile)
	if err != nil {
		_ = engine.DestroySandbox(ctx, sb)
		return nil, fmt.Errorf("failed to create tunnel adapter: %w", err)
	}

	if err := tunWorker.Start(ctx); err != nil {
		_ = engine.DestroySandbox(ctx, sb)
		return nil, fmt.Errorf("failed to start tunnel worker: %w", err)
	}

	// 4. Setup logs path
	logDir := filepath.Join(config.GetRuntimeDir(), "logs")
	_ = os.MkdirAll(logDir, 0700)
	logFile := filepath.Join(logDir, sessionID+".log")

	info := &config.SessionInfo{
		ID:          sessionID,
		Command:     targetCmd,
		Args:        args,
		ProfileID:   profile.ID,
		ProfileName: profile.Name,
		Protocol:    string(profile.Protocol),
		Status:      "running",
		Namespace:   sb.Namespace,
		IPAddress:   sb.IPAddress,
		DNS:         sb.DNS,
		StartTime:   time.Now(),
		Detached:    true,
		LogFilePath: logFile,
	}

	sessionCtx, sessionCancel := context.WithCancel(context.Background())

	active := &ActiveSession{
		Info:    info,
		Sandbox: sb,
		Tunnel:  tunWorker,
		Cancel:  sessionCancel,
		Engine:  engine,
	}

	// 5. Execute Command
	if inTerminal {
		// Spawn in external terminal window
		spawner := terminal.NewSpawner(s.cfg.Settings.PreferredTerm)
		if err := spawner.SpawnInExternalTerminal(sessionCtx, sessionID, targetCmd, targetCmd, args); err != nil {
			_ = tunWorker.Stop()
			_ = engine.DestroySandbox(ctx, sb)
			sessionCancel()
			return nil, fmt.Errorf("failed to spawn external terminal: %w", err)
		}
	} else {
		// Spawn background process
		var cmd *exec.Cmd
		if envEng, ok := engine.(*isolation.EnvProxyEngine); ok {
			cmd, err = envEng.BuildCommandWithProfile(sessionCtx, sb, profile, targetCmd, args...)
		} else {
			cmd, err = engine.BuildCommand(sessionCtx, sb, targetCmd, args...)
		}

		if err != nil {
			_ = tunWorker.Stop()
			_ = engine.DestroySandbox(ctx, sb)
			sessionCancel()
			return nil, fmt.Errorf("failed to build command: %w", err)
		}

		// Direct log output
		if logHandle, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
			cmd.Stdout = logHandle
			cmd.Stderr = logHandle
		}

		if err := cmd.Start(); err != nil {
			_ = tunWorker.Stop()
			_ = engine.DestroySandbox(ctx, sb)
			sessionCancel()
			return nil, fmt.Errorf("failed to start process: %w", err)
		}

		active.Cmd = cmd
		info.PID = cmd.Process.Pid

		// Monitor process exit in background
		go func() {
			_ = cmd.Wait()
			active.mu.Lock()
			info.Status = "stopped"
			active.mu.Unlock()
		}()
	}

	s.sessions[sessionID] = active
	return info, nil
}

// ListSessions returns a slice of all tracked sessions with up-to-date metrics.
func (s *Supervisor) ListSessions() []*config.SessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*config.SessionInfo
	for _, sess := range s.sessions {
		sess.mu.RLock()
		if sess.Tunnel != nil {
			rx, tx := sess.Tunnel.GetMetrics()
			sess.Info.BytesRX = rx
			sess.Info.BytesTX = tx
		}
		list = append(list, sess.Info)
		sess.mu.RUnlock()
	}
	return list
}

// KillSession terminates a running session and releases resources.
func (s *Supervisor) KillSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session '%s' not found", sessionID)
	}

	sess.mu.Lock()
	defer sess.mu.Unlock()

	if sess.Cancel != nil {
		sess.Cancel()
	}

	if sess.Cmd != nil && sess.Cmd.Process != nil {
		_ = syscall.Kill(-sess.Cmd.Process.Pid, syscall.SIGTERM)
		_ = sess.Cmd.Process.Kill()
	}

	if sess.Tunnel != nil {
		_ = sess.Tunnel.Stop()
	}

	if sess.Engine != nil && sess.Sandbox != nil {
		_ = sess.Engine.DestroySandbox(ctx, sess.Sandbox)
	}

	sess.Info.Status = "stopped"
	delete(s.sessions, sessionID)

	return nil
}

// CleanupAll terminates all active sessions during shutdown.
func (s *Supervisor) CleanupAll(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, sess := range s.sessions {
		if sess.Cancel != nil {
			sess.Cancel()
		}
		if sess.Cmd != nil && sess.Cmd.Process != nil {
			_ = sess.Cmd.Process.Kill()
		}
		if sess.Tunnel != nil {
			_ = sess.Tunnel.Stop()
		}
		if sess.Engine != nil && sess.Sandbox != nil {
			_ = sess.Engine.DestroySandbox(ctx, sess.Sandbox)
		}
		delete(s.sessions, id)
	}
}
