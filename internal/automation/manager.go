//go:build !no_automation

package automation

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// validScriptID checks that a script ID is safe to use as a filename component.
func validScriptID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	if strings.ContainsAny(id, "/\\") || strings.Contains(id, "..") {
		return false
	}
	return true
}

var blocklyRe = regexp.MustCompile(`(?s)--\[\[BLOCKLY_XML\n(.*?)\nBLOCKLY_XML\]\]--`)

// Manager handles loading, saving, and listing automation scripts from disk.
type Manager struct {
	dir string
	mu  sync.RWMutex
}

// NewManager creates a new script manager rooted at dir.
// It ensures the directory exists.
func NewManager(dir string) (*Manager, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create scripts dir: %w", err)
	}
	return &Manager{dir: dir}, nil
}

// List returns all scripts found in the directory.
func (m *Manager) List() ([]*Script, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("read scripts dir: %w", err)
	}

	var scripts []*Script
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".lua") {
			continue
		}
		s, err := m.parseFile(filepath.Join(m.dir, e.Name()))
		if err != nil {
			continue // skip malformed scripts
		}
		scripts = append(scripts, s)
	}
	return scripts, nil
}

// Get returns a single script by ID (filename stem).
func (m *Manager) Get(id string) (*Script, error) {
	if !validScriptID(id) {
		return nil, fmt.Errorf("invalid script id: %q", id)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	path := filepath.Join(m.dir, id+".lua")
	return m.parseFile(path)
}

// Save writes a script to disk. If the script has no ID, one is generated
// from the name. Returns the (possibly updated) script.
func (m *Manager) Save(s *Script) (*Script, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s.ID == "" {
		s.ID = slugify(s.Meta.Name)
		if s.ID == "" {
			s.ID = "script"
		}
		// Ensure unique ID
		base := s.ID
		for i := 1; ; i++ {
			path := filepath.Join(m.dir, s.ID+".lua")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				break
			}
			s.ID = fmt.Sprintf("%s_%d", base, i)
		}
	}

	s.FilePath = filepath.Join(m.dir, s.ID+".lua")
	content := serializeScript(s)

	if err := os.WriteFile(s.FilePath, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("write script: %w", err)
	}
	return s, nil
}

// Delete removes a script file by ID.
func (m *Manager) Delete(id string) error {
	if !validScriptID(id) {
		return fmt.Errorf("invalid script id: %q", id)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	path := filepath.Join(m.dir, id+".lua")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete script: %w", err)
	}
	return nil
}

// parseFile reads and parses a .lua script file.
func (m *Manager) parseFile(path string) (*Script, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	s := &Script{
		ID:       strings.TrimSuffix(filepath.Base(path), ".lua"),
		FilePath: path,
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Parse JSON metadata from first line: -- {"name": "...", ...}
	if len(lines) > 0 && strings.HasPrefix(lines[0], "-- {") {
		jsonStr := strings.TrimPrefix(lines[0], "-- ")
		if err := json.Unmarshal([]byte(jsonStr), &s.Meta); err != nil {
			slog.Warn("script metadata parse error", "file", path, "err", err)
		}
	}

	// Extract Blockly XML from --[[BLOCKLY_XML ... BLOCKLY_XML]]--
	if match := blocklyRe.FindStringSubmatch(content); len(match) > 1 {
		s.BlocklyXML = match[1]
	}

	// Lua code is everything after the header comment lines and Blockly block
	luaLines := make([]string, 0, len(lines))
	inBlockly := false
	for i, line := range lines {
		if i == 0 && strings.HasPrefix(line, "-- {") {
			continue // skip metadata line
		}
		if strings.HasPrefix(line, "--[[BLOCKLY_XML") {
			inBlockly = true
			continue
		}
		if inBlockly {
			if strings.HasPrefix(line, "BLOCKLY_XML]]--") {
				inBlockly = false
			}
			continue
		}
		luaLines = append(luaLines, line)
	}

	if inBlockly {
		slog.Warn("unclosed BLOCKLY_XML block", "file", path)
	}

	// Trim leading empty lines from Lua code
	for len(luaLines) > 0 && strings.TrimSpace(luaLines[0]) == "" {
		luaLines = luaLines[1:]
	}
	s.LuaCode = strings.Join(luaLines, "\n")

	return s, nil
}

// serializeScript reassembles a script file from its parts.
func serializeScript(s *Script) string {
	var b strings.Builder

	// Metadata line
	meta, _ := json.Marshal(s.Meta)
	b.WriteString("-- ")
	b.Write(meta)
	b.WriteString("\n")

	// Blockly XML block
	if s.BlocklyXML != "" {
		b.WriteString("--[[BLOCKLY_XML\n")
		b.WriteString(s.BlocklyXML)
		b.WriteString("\nBLOCKLY_XML]]--\n")
	}

	// Lua code
	if s.LuaCode != "" {
		b.WriteString("\n")
		b.WriteString(s.LuaCode)
		if !strings.HasSuffix(s.LuaCode, "\n") {
			b.WriteString("\n")
		}
	}

	return b.String()
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugRe.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}
