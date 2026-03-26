package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configFileName  = "config.json"
	sessionFileName = "session.json"
)

type AppConfig struct {
	DefaultClub string `json:"default_club"`
}

type Session struct {
	Token    string `json:"token"`
	DeviceID string `json:"device_id"`
	HostID   string `json:"host_id"`
}

type Manager struct {
	rootDir string
}

func NewManager(rootDir string) (*Manager, error) {
	if rootDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home directory: %w", err)
		}
		rootDir = filepath.Join(home, ".swims")
	}
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolving config directory: %w", err)
	}
	return &Manager{rootDir: rootDir}, nil
}

func (m *Manager) RootDir() string {
	return m.rootDir
}

func (m *Manager) DataDir() string {
	return m.rootDir
}

func (m *Manager) ConfigPath() string {
	return filepath.Join(m.rootDir, configFileName)
}

func (m *Manager) SessionPath() string {
	return filepath.Join(m.rootDir, sessionFileName)
}

func (m *Manager) EnsureDir() error {
	if err := os.MkdirAll(m.rootDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	return nil
}

func (m *Manager) Load() (*AppConfig, error) {
	var cfg AppConfig
	if err := readJSONFile(m.ConfigPath(), &cfg); err != nil {
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return &cfg, nil
}

func (m *Manager) Save(cfg *AppConfig) error {
	if err := m.EnsureDir(); err != nil {
		return err
	}
	if cfg == nil {
		cfg = &AppConfig{}
	}
	if err := writeJSONFile(m.ConfigPath(), cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	return nil
}

func (m *Manager) LoadSession() (*Session, error) {
	var session Session
	if err := readJSONFile(m.SessionPath(), &session); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("loading session: %w", err)
	}
	if session.Token == "" {
		return nil, nil
	}
	return &session, nil
}

func (m *Manager) SaveSession(session *Session) error {
	if err := m.EnsureDir(); err != nil {
		return err
	}
	if session == nil || session.Token == "" {
		return m.ClearSession()
	}
	if err := writeJSONFile(m.SessionPath(), session); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}
	return nil
}

func (m *Manager) ClearSession() error {
	err := os.Remove(m.SessionPath())
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clearing session: %w", err)
	}
	return nil
}

func readJSONFile(path string, out interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}
	return nil
}

func writeJSONFile(path string, value interface{}) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0644); err != nil {
		return err
	}
	return nil
}
