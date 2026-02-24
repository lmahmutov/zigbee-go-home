package ncp

import (
	"bytes"
	"testing"
)

func TestTypeSizeFixedTypes(t *testing.T) {
	tests := []struct {
		typeID uint8
		name   string
		want   int
	}{
		// Data types (0x08-0x0F)
		{0x08, "data8", 1},
		{0x09, "data16", 2},
		{0x0A, "data24", 3},
		{0x0B, "data32", 4},
		{0x0C, "data40", 5},
		{0x0D, "data48", 6},
		{0x0E, "data56", 7},
		{0x0F, "data64", 8},

		// Bool
		{0x10, "bool", 1},

		// Bitmap types
		{0x18, "map8", 1},
		{0x19, "map16", 2},
		{0x1A, "map24", 3},
		{0x1B, "map32", 4},

		// Unsigned integers
		{0x20, "uint8", 1},
		{0x21, "uint16", 2},
		{0x22, "uint24", 3},
		{0x23, "uint32", 4},
		{0x24, "uint40", 5},
		{0x25, "uint48", 6},
		{0x26, "uint56", 7},
		{0x27, "uint64", 8},

		// Signed integers
		{0x28, "int8", 1},
		{0x29, "int16", 2},
		{0x2A, "int24", 3},
		{0x2B, "int32", 4},
		{0x2C, "int40", 5},
		{0x2D, "int48", 6},
		{0x2E, "int56", 7},
		{0x2F, "int64", 8},

		// Enum types
		{0x30, "enum8", 1},
		{0x31, "enum16", 2},

		// Float types
		{0x38, "float16", 2},
		{0x39, "float32", 4},
		{0x3A, "float64", 8},

		// Time types
		{0xE0, "ToD", 4},
		{0xE1, "Date", 4},
		{0xE2, "UTC", 4},

		// Identifier types
		{0xE8, "ClusterID", 2},
		{0xE9, "AttrID", 2},

		// EUI64
		{0xF0, "EUI64", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeSize(tt.typeID)
			if got != tt.want {
				t.Errorf("typeSize(0x%02X) = %d, want %d", tt.typeID, got, tt.want)
			}
		})
	}
}

func TestTypeSizeVariableTypes(t *testing.T) {
	tests := []struct {
		typeID uint8
		name   string
		want   int
	}{
		{0x41, "octstr", typeSizeVariable},
		{0x42, "string", typeSizeVariable},
		{0x43, "octstr16", typeSizeVariable16},
		{0x44, "string16", typeSizeVariable16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeSize(tt.typeID)
			if got != tt.want {
				t.Errorf("typeSize(0x%02X) = %d, want %d", tt.typeID, got, tt.want)
			}
		})
	}
}

func TestTypeSizeUnknown(t *testing.T) {
	// Type 0x00 (nodata) and other unhandled types should be unknown
	if got := typeSize(0x50); got != typeSizeUnknown {
		t.Errorf("typeSize(0x50) = %d, want %d (unknown)", got, typeSizeUnknown)
	}
}

func TestParseAttributeResponsesSingle(t *testing.T) {
	// AttrID=0x0000, Status=0x00, DataType=0x10 (bool), Value=0x01
	data := []byte{0x00, 0x00, 0x00, 0x10, 0x01}

	results := parseAttributeResponses(data)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	r := results[0]
	if r.AttrID != 0x0000 {
		t.Errorf("AttrID = 0x%04X, want 0x0000", r.AttrID)
	}
	if r.Status != 0x00 {
		t.Errorf("Status = 0x%02X, want 0x00", r.Status)
	}
	if r.DataType != 0x10 {
		t.Errorf("DataType = 0x%02X, want 0x10", r.DataType)
	}
	if !bytes.Equal(r.Value, []byte{0x01}) {
		t.Errorf("Value = %X, want [01]", r.Value)
	}
}

func TestParseAttributeResponsesMultiple(t *testing.T) {
	// Two attributes: uint8 and uint16
	data := []byte{
		// Attr 0x0000, status 0, type uint8 (0x20), value 0x42
		0x00, 0x00, 0x00, 0x20, 0x42,
		// Attr 0x0001, status 0, type uint16 (0x21), value 0x3412
		0x01, 0x00, 0x00, 0x21, 0x12, 0x34,
	}

	results := parseAttributeResponses(data)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].AttrID != 0x0000 || !bytes.Equal(results[0].Value, []byte{0x42}) {
		t.Errorf("attr[0]: id=0x%04X, value=%X", results[0].AttrID, results[0].Value)
	}
	if results[1].AttrID != 0x0001 || !bytes.Equal(results[1].Value, []byte{0x12, 0x34}) {
		t.Errorf("attr[1]: id=0x%04X, value=%X", results[1].AttrID, results[1].Value)
	}
}

func TestParseAttributeResponsesErrorStatus(t *testing.T) {
	// Attr with error status (no data type or value follows)
	data := []byte{
		// Attr 0x0005, status 0x86 (unsupported)
		0x05, 0x00, 0x86,
		// Attr 0x0004, status 0, type string (0x42), value "Hi"
		0x04, 0x00, 0x00, 0x42, 0x02, 'H', 'i',
	}

	results := parseAttributeResponses(data)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].AttrID != 0x0005 || results[0].Status != 0x86 {
		t.Errorf("attr[0]: id=0x%04X, status=0x%02X", results[0].AttrID, results[0].Status)
	}
	if results[0].Value != nil {
		t.Errorf("attr[0] with error should have nil value, got %X", results[0].Value)
	}
	if results[1].AttrID != 0x0004 || !bytes.Equal(results[1].Value, []byte{0x02, 'H', 'i'}) {
		t.Errorf("attr[1]: id=0x%04X, value=%X", results[1].AttrID, results[1].Value)
	}
}

func TestParseAttributeResponsesVariableLength16(t *testing.T) {
	// Attr with octstr16 (0x43) type: 2-byte length prefix
	data := []byte{
		0x00, 0x00, 0x00, 0x43, // AttrID=0, status=0, type=octstr16
		0x03, 0x00, // length = 3
		0xAA, 0xBB, 0xCC, // value
	}

	results := parseAttributeResponses(data)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	expected := []byte{0x03, 0x00, 0xAA, 0xBB, 0xCC}
	if !bytes.Equal(results[0].Value, expected) {
		t.Errorf("Value = %X, want %X", results[0].Value, expected)
	}
}

func TestParseAttributeResponsesFloat16(t *testing.T) {
	// This tests the bug fix: float16 (0x38) was previously unhandled
	data := []byte{
		0x00, 0x00, 0x00, 0x38, // AttrID=0, status=0, type=float16
		0x00, 0x3C, // value (some float16)
		// Second attr should still parse
		0x01, 0x00, 0x00, 0x20, 0xFF, // AttrID=1, status=0, type=uint8, value=0xFF
	}

	results := parseAttributeResponses(data)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2 (float16 must not stop parsing)", len(results))
	}
	if !bytes.Equal(results[0].Value, []byte{0x00, 0x3C}) {
		t.Errorf("float16 value = %X, want [003C]", results[0].Value)
	}
	if results[1].AttrID != 0x0001 {
		t.Errorf("second attr should parse after float16, got id=0x%04X", results[1].AttrID)
	}
}

func TestParseAttributeResponsesData8(t *testing.T) {
	// This tests the bug fix: data8 (0x08) was previously unhandled
	data := []byte{
		0x00, 0x00, 0x00, 0x08, 0xAB, // data8
		0x01, 0x00, 0x00, 0x20, 0x42, // uint8
	}

	results := parseAttributeResponses(data)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2 (data8 must not stop parsing)", len(results))
	}
}

func TestParseAttributeResponsesEmpty(t *testing.T) {
	results := parseAttributeResponses(nil)
	if len(results) != 0 {
		t.Errorf("got %d results for nil data, want 0", len(results))
	}
	results = parseAttributeResponses([]byte{})
	if len(results) != 0 {
		t.Errorf("got %d results for empty data, want 0", len(results))
	}
}
