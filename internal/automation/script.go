//go:build !no_automation

package automation

// ScriptMeta holds user-editable metadata for a script.
type ScriptMeta struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// Script represents a single automation script stored on disk.
type Script struct {
	ID         string     `json:"id"`          // filename stem (no .lua)
	Meta       ScriptMeta `json:"meta"`
	LuaCode    string     `json:"lua_code"`    // raw Lua source (without header)
	BlocklyXML string     `json:"blockly_xml"` // Blockly workspace XML
	FilePath   string     `json:"-"`           // absolute path on disk
}
