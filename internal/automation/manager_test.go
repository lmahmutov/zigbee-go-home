//go:build !no_automation

package automation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "scripts")
	m, err := NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestManagerListEmpty(t *testing.T) {
	m := newTestManager(t)
	scripts, err := m.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) != 0 {
		t.Errorf("list count = %d, want 0", len(scripts))
	}
}

func TestManagerSaveAndGet(t *testing.T) {
	m := newTestManager(t)

	s := &Script{
		Meta: ScriptMeta{
			Name:        "Test Script",
			Description: "A test",
			Enabled:     true,
		},
		LuaCode:    `zigbee.log("hello")`,
		BlocklyXML: `<block type="zigbee_log"></block>`,
	}

	saved, err := m.Save(s)
	if err != nil {
		t.Fatal(err)
	}

	if saved.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if saved.ID != "test_script" {
		t.Errorf("id = %q, want test_script", saved.ID)
	}

	got, err := m.Get(saved.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.Meta.Name != "Test Script" {
		t.Errorf("name = %q, want Test Script", got.Meta.Name)
	}
	if got.Meta.Description != "A test" {
		t.Errorf("description = %q, want A test", got.Meta.Description)
	}
	if !got.Meta.Enabled {
		t.Error("enabled = false, want true")
	}
	if !strings.Contains(got.LuaCode, `zigbee.log("hello")`) {
		t.Errorf("lua_code = %q, want to contain zigbee.log", got.LuaCode)
	}
	if got.BlocklyXML != `<block type="zigbee_log"></block>` {
		t.Errorf("blockly_xml = %q", got.BlocklyXML)
	}
}

func TestManagerSaveExistingID(t *testing.T) {
	m := newTestManager(t)

	s := &Script{
		ID: "my_script",
		Meta: ScriptMeta{
			Name:    "My Script",
			Enabled: true,
		},
		LuaCode: `zigbee.log("v1")`,
	}

	saved, err := m.Save(s)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID != "my_script" {
		t.Errorf("id = %q, want my_script", saved.ID)
	}

	// Update same script
	saved.LuaCode = `zigbee.log("v2")`
	_, err = m.Save(saved)
	if err != nil {
		t.Fatal(err)
	}

	got, err := m.Get("my_script")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.LuaCode, `zigbee.log("v2")`) {
		t.Errorf("lua_code after update = %q", got.LuaCode)
	}
}

func TestManagerList(t *testing.T) {
	m := newTestManager(t)

	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		_, err := m.Save(&Script{
			Meta:    ScriptMeta{Name: name, Enabled: true},
			LuaCode: `zigbee.log("` + name + `")`,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	scripts, err := m.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) != 3 {
		t.Fatalf("list count = %d, want 3", len(scripts))
	}
}

func TestManagerDelete(t *testing.T) {
	m := newTestManager(t)

	saved, err := m.Save(&Script{
		Meta:    ScriptMeta{Name: "ToDelete", Enabled: true},
		LuaCode: `zigbee.log("bye")`,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Delete(saved.ID); err != nil {
		t.Fatal(err)
	}

	_, err = m.Get(saved.ID)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestManagerGetNotFound(t *testing.T) {
	m := newTestManager(t)

	_, err := m.Get("nonexistent")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestManagerUniqueID(t *testing.T) {
	m := newTestManager(t)

	s1, err := m.Save(&Script{
		Meta:    ScriptMeta{Name: "Dup", Enabled: true},
		LuaCode: `zigbee.log("1")`,
	})
	if err != nil {
		t.Fatal(err)
	}

	s2, err := m.Save(&Script{
		Meta:    ScriptMeta{Name: "Dup", Enabled: true},
		LuaCode: `zigbee.log("2")`,
	})
	if err != nil {
		t.Fatal(err)
	}

	if s1.ID == s2.ID {
		t.Errorf("expected unique IDs, got %q for both", s1.ID)
	}
}

func TestParseScriptFile(t *testing.T) {
	dir := t.TempDir()
	content := `-- {"name":"Bathroom Light","description":"Turn on when door opens","enabled":true}
--[[BLOCKLY_XML
<block type="zigbee_on_property"></block>
BLOCKLY_XML]]--

zigbee.on("property_update", {ieee="ABC123"}, function(event)
    zigbee.turn_on("DEF456")
end)
`
	path := filepath.Join(dir, "bathroom.lua")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	m := &Manager{dir: dir}
	s, err := m.parseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if s.ID != "bathroom" {
		t.Errorf("id = %q, want bathroom", s.ID)
	}
	if s.Meta.Name != "Bathroom Light" {
		t.Errorf("name = %q, want Bathroom Light", s.Meta.Name)
	}
	if s.Meta.Description != "Turn on when door opens" {
		t.Errorf("description = %q", s.Meta.Description)
	}
	if !s.Meta.Enabled {
		t.Error("enabled = false, want true")
	}
	if s.BlocklyXML != `<block type="zigbee_on_property"></block>` {
		t.Errorf("blockly_xml = %q", s.BlocklyXML)
	}
	if !strings.Contains(s.LuaCode, `zigbee.on("property_update"`) {
		t.Errorf("lua_code missing expected content: %q", s.LuaCode)
	}
	if !strings.Contains(s.LuaCode, `zigbee.turn_on("DEF456")`) {
		t.Errorf("lua_code missing turn_on: %q", s.LuaCode)
	}
}

func TestSerializeScript(t *testing.T) {
	s := &Script{
		ID: "test",
		Meta: ScriptMeta{
			Name:        "Test",
			Description: "desc",
			Enabled:     true,
		},
		LuaCode:    `zigbee.log("hi")`,
		BlocklyXML: `<block/>`,
	}

	content := serializeScript(s)

	if !strings.HasPrefix(content, "-- {") {
		t.Errorf("expected metadata line prefix, got: %q", content[:20])
	}
	if !strings.Contains(content, "--[[BLOCKLY_XML") {
		t.Error("missing BLOCKLY_XML block")
	}
	if !strings.Contains(content, "BLOCKLY_XML]]--") {
		t.Error("missing BLOCKLY_XML closing")
	}
	if !strings.Contains(content, `<block/>`) {
		t.Error("missing blockly content")
	}
	if !strings.Contains(content, `zigbee.log("hi")`) {
		t.Error("missing lua code")
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Bathroom Light", "bathroom_light"},
		{"hello world!", "hello_world"},
		{"", ""},
		{"  spaces  ", "spaces"},
		{"UPPER", "upper"},
	}
	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
