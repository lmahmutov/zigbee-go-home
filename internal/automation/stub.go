//go:build no_automation

package automation

import (
	"log/slog"
	"time"

	"zigbee-go-home/internal/coordinator"
)

// ScriptMeta holds user-editable metadata for a script.
type ScriptMeta struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// Script represents a single automation script stored on disk.
type Script struct {
	ID         string     `json:"id"`
	Meta       ScriptMeta `json:"meta"`
	LuaCode    string     `json:"lua_code"`
	BlocklyXML string     `json:"blockly_xml"`
	FilePath   string     `json:"-"`
}

// RunResult is the result of a one-shot script execution.
type RunResult struct {
	OK       bool     `json:"ok"`
	Error    string   `json:"error,omitempty"`
	Logs     []string `json:"logs"`
	Duration string   `json:"duration"`
}

// SystemConfig holds system exec settings (stub).
type SystemConfig struct {
	ExecAllowlist []string
	ExecTimeout   time.Duration
}

// TelegramConfig holds Telegram bot settings (stub).
type TelegramConfig struct {
	BotToken string
	ChatIDs  []string
}

// Manager is a no-op stub when automation is disabled.
type Manager struct{}

// NewManager returns nil manager when automation is disabled.
func NewManager(_ string) (*Manager, error) { return nil, nil }

// List returns nil.
func (m *Manager) List() ([]*Script, error) { return nil, nil }

// Get returns nil.
func (m *Manager) Get(_ string) (*Script, error) { return nil, nil }

// Save returns the script unchanged.
func (m *Manager) Save(s *Script) (*Script, error) { return s, nil }

// Delete is a no-op.
func (m *Manager) Delete(_ string) error { return nil }

// Engine is a no-op stub when automation is disabled.
type Engine struct{}

// NewEngine returns a no-op engine when automation is disabled.
func NewEngine(_ *coordinator.Coordinator, _ *Manager, _ *slog.Logger, _ SystemConfig, _ TelegramConfig) *Engine {
	return &Engine{}
}

// Start is a no-op.
func (e *Engine) Start() {}

// Stop is a no-op.
func (e *Engine) Stop() {}

// ReloadScript is a no-op.
func (e *Engine) ReloadScript(_ string) error { return nil }

// StopScript is a no-op.
func (e *Engine) StopScript(_ string) {}

// RunScript returns a stub result.
func (e *Engine) RunScript(_ string) *RunResult {
	return &RunResult{OK: false, Error: "automation disabled"}
}

// RunLuaCode returns a stub result.
func (e *Engine) RunLuaCode(_ string) *RunResult {
	return &RunResult{OK: false, Error: "automation disabled"}
}
