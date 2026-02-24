package coordinator

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"zigbee-go-home/internal/zcl"
)

// Real Xiaomi TLV payload captured from lumi.sensor_magnet.aq2 attribute 0xFF01.
// Tags: 1=battery(uint16 3055), 3=temperature(int8 31), 4=unknown(uint16 23085),
// 5=power_outage_count(uint16 2), 6=trigger_count(uint40 2), 100=contact(bool true)
var xiaomiTestPayload = []byte{
	0x01, 0x21, 0xEF, 0x0B, // tag=1, type=uint16, value=3055
	0x03, 0x28, 0x1F, // tag=3, type=int8, value=31
	0x04, 0x21, 0x2D, 0x5A, // tag=4, type=uint16, value=23085
	0x05, 0x21, 0x02, 0x00, // tag=5, type=uint16, value=2
	0x06, 0x24, 0x02, 0x00, 0x00, 0x00, 0x00, // tag=6, type=uint40, value=2
	0x64, 0x10, 0x01, // tag=100, type=bool, value=true
}

func TestDecodeXiaomiTLV(t *testing.T) {
	result, err := decodeXiaomiTLV(xiaomiTestPayload)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		tag  int
		want interface{}
	}{
		{1, uint16(3055)},
		{3, int8(31)},
		{4, uint16(23085)},
		{5, uint16(2)},
		{6, uint64(2)},
		{100, true},
	}

	for _, tt := range tests {
		got, ok := result[tt.tag]
		if !ok {
			t.Errorf("tag %d not found", tt.tag)
			continue
		}
		if got != tt.want {
			t.Errorf("tag %d = %v (%T), want %v (%T)", tt.tag, got, got, tt.want, tt.want)
		}
	}
}

func TestDecodeXiaomiTLVEmpty(t *testing.T) {
	result, err := decodeXiaomiTLV(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestDecodeXiaomiTLVTruncated(t *testing.T) {
	// Just a tag byte, no type — should return empty map, no error
	result, err := decodeXiaomiTLV([]byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestApplyTransformLumiBattery(t *testing.T) {
	tests := []struct {
		input interface{}
		want  int
	}{
		{uint16(3055), 100}, // above max, clamped
		{uint16(3000), 100},
		{uint16(2925), 50},
		{uint16(2850), 0},
		{uint16(2700), 0}, // below min, clamped
	}

	for _, tt := range tests {
		got := applyTransform("lumi_battery", tt.input)
		if got != tt.want {
			t.Errorf("lumi_battery(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestApplyTransformMinusOne(t *testing.T) {
	got := applyTransform("minus_one", uint16(2))
	if got != int64(1) {
		t.Errorf("minus_one(2) = %v (%T), want 1", got, got)
	}
}

func TestApplyTransformLumiTrigger(t *testing.T) {
	// uint40 value = 0x0000000002 → lower 16 bits = 2 → minus 1 = 1
	got := applyTransform("lumi_trigger", uint64(2))
	if got != int64(1) {
		t.Errorf("lumi_trigger(2) = %v (%T), want 1", got, got)
	}

	// Larger value: 0x0001000A → lower 16 = 10 → minus 1 = 9
	got = applyTransform("lumi_trigger", uint64(0x0001000A))
	if got != int64(9) {
		t.Errorf("lumi_trigger(0x0001000A) = %v, want 9", got)
	}
}

func TestApplyTransformBoolInvert(t *testing.T) {
	got := applyTransform("bool_invert", true)
	if got != false {
		t.Errorf("bool_invert(true) = %v, want false", got)
	}

	got = applyTransform("bool_invert", false)
	if got != true {
		t.Errorf("bool_invert(false) = %v, want true", got)
	}

	got = applyTransform("bool_invert", uint8(0))
	if got != true {
		t.Errorf("bool_invert(uint8(0)) = %v, want true", got)
	}

	got = applyTransform("bool_invert", uint8(1))
	if got != false {
		t.Errorf("bool_invert(uint8(1)) = %v, want false", got)
	}
}

func TestApplyTransformUnknown(t *testing.T) {
	got := applyTransform("nonexistent", uint16(42))
	if got != uint16(42) {
		t.Errorf("unknown transform changed value: %v", got)
	}
}

func TestLoadDeviceDirManufacturers(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	registry := zcl.NewRegistry(logger)

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "xiaomi.json"), []byte(`{
		"manufacturers": [
			{
				"name": "LUMI",
				"models": [
					{
						"model": "lumi.sensor_magnet.aq2",
						"friendly_name": "Aqara Door Sensor",
						"bind": [6],
						"properties": [
							{
								"cluster": 0,
								"attribute": 65281,
								"decoder": "xiaomi_tlv",
								"values": [
									{"tag": 1, "name": "battery_voltage"}
								]
							}
						]
					}
				]
			}
		]
	}`), 0644)

	db, err := LoadDeviceDir(dir, registry, logger)
	if err != nil {
		t.Fatal(err)
	}

	if db.Len() != 1 {
		t.Fatalf("device count = %d, want 1", db.Len())
	}

	def := db.Lookup("LUMI", "lumi.sensor_magnet.aq2")
	if def == nil {
		t.Fatal("lookup returned nil")
	}
	if def.Manufacturer != "LUMI" {
		t.Errorf("manufacturer = %q, want LUMI", def.Manufacturer)
	}
	if def.FriendlyName != "Aqara Door Sensor" {
		t.Errorf("friendly_name = %q", def.FriendlyName)
	}
	if len(def.Properties) != 1 {
		t.Fatalf("properties = %d, want 1", len(def.Properties))
	}
	if def.Properties[0].Decoder != "xiaomi_tlv" {
		t.Errorf("decoder = %q", def.Properties[0].Decoder)
	}
	if len(def.Properties[0].Values) != 1 {
		t.Errorf("values = %d, want 1", len(def.Properties[0].Values))
	}
}
