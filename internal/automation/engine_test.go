//go:build !no_automation

package automation

import (
	"testing"

	"zigbee-go-home/internal/coordinator"
	"zigbee-go-home/internal/store"

	lua "github.com/yuin/gopher-lua"
)

func TestGoToLua(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name string
		val  interface{}
		want lua.LValueType
	}{
		{"nil", nil, lua.LTNil},
		{"bool true", true, lua.LTBool},
		{"bool false", false, lua.LTBool},
		{"string", "hello", lua.LTString},
		{"int", 42, lua.LTNumber},
		{"int64", int64(99), lua.LTNumber},
		{"float64", 3.14, lua.LTNumber},
		{"uint8", uint8(255), lua.LTNumber},
		{"uint16", uint16(1024), lua.LTNumber},
		{"uint32", uint32(100000), lua.LTNumber},
		{"int8", int8(-10), lua.LTNumber},
		{"map", map[string]interface{}{"a": 1}, lua.LTTable},
		{"slice", []interface{}{1, 2, 3}, lua.LTTable},
		{"unknown", struct{}{}, lua.LTString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := goToLua(L, tt.val)
			if result.Type() != tt.want {
				t.Errorf("goToLua(%v) type = %v, want %v", tt.val, result.Type(), tt.want)
			}
		})
	}
}

func TestGoToLuaBoolValues(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	if v := goToLua(L, true); v != lua.LTrue {
		t.Errorf("goToLua(true) = %v, want LTrue", v)
	}
	if v := goToLua(L, false); v != lua.LFalse {
		t.Errorf("goToLua(false) = %v, want LFalse", v)
	}
}

func TestGoToLuaNumberValues(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	v := goToLua(L, 42)
	if n, ok := v.(lua.LNumber); !ok || float64(n) != 42 {
		t.Errorf("goToLua(42) = %v, want LNumber(42)", v)
	}
}

func TestGoToLuaStringValue(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	v := goToLua(L, "hello")
	if s, ok := v.(lua.LString); !ok || string(s) != "hello" {
		t.Errorf("goToLua(hello) = %v, want LString(hello)", v)
	}
}

func TestGoToLuaMap(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	m := map[string]interface{}{"key": "value", "num": 10}
	v := goToLua(L, m)
	tbl, ok := v.(*lua.LTable)
	if !ok {
		t.Fatal("expected LTable")
	}

	keyVal := tbl.RawGetString("key")
	if s, ok := keyVal.(lua.LString); !ok || string(s) != "value" {
		t.Errorf("map[key] = %v, want value", keyVal)
	}

	numVal := tbl.RawGetString("num")
	if n, ok := numVal.(lua.LNumber); !ok || float64(n) != 10 {
		t.Errorf("map[num] = %v, want 10", numVal)
	}
}

func TestGoToLuaSlice(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	s := []interface{}{"a", "b", "c"}
	v := goToLua(L, s)
	tbl, ok := v.(*lua.LTable)
	if !ok {
		t.Fatal("expected LTable")
	}

	if tbl.Len() != 3 {
		t.Errorf("table len = %d, want 3", tbl.Len())
	}

	first := tbl.RawGetInt(1)
	if str, ok := first.(lua.LString); !ok || string(str) != "a" {
		t.Errorf("slice[1] = %v, want a", first)
	}
}

func TestMatchesHandler(t *testing.T) {
	tests := []struct {
		name    string
		handler luaEventHandler
		evType  string
		evData  map[string]interface{}
		want    bool
	}{
		{
			"exact match",
			luaEventHandler{eventType: "property_update", ieee: "AABB", property: "contact"},
			"property_update",
			map[string]interface{}{"ieee": "AABB", "property": "contact"},
			true,
		},
		{
			"wrong event type",
			luaEventHandler{eventType: "property_update"},
			"attribute_report",
			map[string]interface{}{},
			false,
		},
		{
			"ieee filter mismatch",
			luaEventHandler{eventType: "property_update", ieee: "AABB"},
			"property_update",
			map[string]interface{}{"ieee": "CCDD"},
			false,
		},
		{
			"property filter mismatch",
			luaEventHandler{eventType: "property_update", property: "contact"},
			"property_update",
			map[string]interface{}{"property": "battery"},
			false,
		},
		{
			"no filters match any",
			luaEventHandler{eventType: "property_update"},
			"property_update",
			map[string]interface{}{"ieee": "AABB", "property": "contact"},
			true,
		},
		{
			"ieee filter only",
			luaEventHandler{eventType: "property_update", ieee: "AABB"},
			"property_update",
			map[string]interface{}{"ieee": "AABB", "property": "anything"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesHandler(tt.handler, coordinator.Event{
				Type: tt.evType,
				Data: tt.evData,
			})
			if got != tt.want {
				t.Errorf("matchesHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"00158D00012A3B4C", true},
		{"abcdef0123456789", true},
		{"ABCDEF0123456789", true},
		{"not_hex!", false},
		{"", true},
		{"0123456789abcdefg", false},
	}

	for _, tt := range tests {
		got := isHexString(tt.input)
		if got != tt.want {
			t.Errorf("isHexString(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestFindEndpointWithCluster(t *testing.T) {
	dev := &store.Device{
		Endpoints: []store.Endpoint{
			{ID: 1, InClusters: []uint16{0x0000, 0x0006}},
			{ID: 2, InClusters: []uint16{0x0008, 0x0300}},
		},
	}

	if ep := findEndpointWithCluster(dev, 0x0006); ep != 1 {
		t.Errorf("findEndpointWithCluster(0x0006) = %d, want 1", ep)
	}
	if ep := findEndpointWithCluster(dev, 0x0008); ep != 2 {
		t.Errorf("findEndpointWithCluster(0x0008) = %d, want 2", ep)
	}
	// Not found — should return first endpoint
	if ep := findEndpointWithCluster(dev, 0x9999); ep != 1 {
		t.Errorf("findEndpointWithCluster(0x9999) = %d, want 1 (fallback)", ep)
	}

	// Empty endpoints — should return 1
	empty := &store.Device{}
	if ep := findEndpointWithCluster(empty, 0x0006); ep != 1 {
		t.Errorf("findEndpointWithCluster on empty = %d, want 1", ep)
	}
}
