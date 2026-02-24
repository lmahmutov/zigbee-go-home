package coordinator

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"zigbee-go-home/internal/zcl"
)

func TestDeviceDBAddLookup(t *testing.T) {
	db := NewDeviceDB()

	db.Add(DeviceDefinition{
		Manufacturer: "IKEA of Sweden",
		Model:        "TRADFRI on/off switch",
		FriendlyName: "IKEA Switch",
		Bind:         []uint16{6, 8},
	})

	if db.Len() != 1 {
		t.Fatalf("len = %d, want 1", db.Len())
	}

	def := db.Lookup("IKEA of Sweden", "TRADFRI on/off switch")
	if def == nil {
		t.Fatal("lookup returned nil")
	}
	if def.FriendlyName != "IKEA Switch" {
		t.Errorf("friendly_name = %q, want %q", def.FriendlyName, "IKEA Switch")
	}
	if len(def.Bind) != 2 {
		t.Errorf("bind = %v, want [6 8]", def.Bind)
	}

	// Not found
	if db.Lookup("IKEA of Sweden", "unknown") != nil {
		t.Error("expected nil for unknown model")
	}
}

func TestLoadDeviceDir(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	registry := zcl.NewRegistry(logger)

	// Pre-register On/Off so we can test merge
	registry.Register(zcl.ClusterDef{
		ID:   6,
		Name: "On/Off",
		Attributes: []zcl.AttributeDef{
			{ID: 0, Name: "OnOff", Type: zcl.TypeBool, Access: zcl.AccessRead},
		},
	})

	dir := t.TempDir()

	// Write a device file with clusters + devices
	os.WriteFile(filepath.Join(dir, "test.json"), []byte(`{
		"clusters": [
			{
				"id": 64528,
				"name": "Tuya Private",
				"attributes": [
					{"id": 0, "name": "TuyaCmd", "type": 65, "access": 3}
				]
			},
			{
				"id": 6,
				"attributes": [
					{"id": 16387, "name": "IkeaStartup", "type": 48, "access": 3}
				]
			}
		],
		"devices": [
			{
				"manufacturer": "LUMI",
				"model": "lumi.sensor_ht",
				"friendly_name": "Aqara Temp",
				"bind": [],
				"reporting": [
					{"cluster": 1026, "attribute": 0, "type": 41, "min": 10, "max": 300, "change": 10}
				]
			}
		]
	}`), 0644)

	// Write a second file with only devices
	os.WriteFile(filepath.Join(dir, "ikea.json"), []byte(`{
		"devices": [
			{
				"manufacturer": "IKEA of Sweden",
				"model": "TRADFRI on/off switch",
				"friendly_name": "IKEA Switch",
				"bind": [6, 8]
			}
		]
	}`), 0644)

	db, err := LoadDeviceDir(dir, registry, logger)
	if err != nil {
		t.Fatal(err)
	}

	// Clusters merged into registry
	tuya := registry.Get(64528)
	if tuya == nil {
		t.Fatal("Tuya cluster not found in registry")
	}
	if tuya.Name != "Tuya Private" {
		t.Errorf("Tuya name = %q", tuya.Name)
	}

	onoff := registry.Get(6)
	if len(onoff.Attributes) != 2 {
		t.Errorf("On/Off attrs = %d, want 2", len(onoff.Attributes))
	}

	// Devices loaded
	if db.Len() != 2 {
		t.Fatalf("device count = %d, want 2", db.Len())
	}

	lumi := db.Lookup("LUMI", "lumi.sensor_ht")
	if lumi == nil {
		t.Fatal("LUMI device not found")
	}
	if lumi.FriendlyName != "Aqara Temp" {
		t.Errorf("friendly_name = %q", lumi.FriendlyName)
	}
	if len(lumi.Reporting) != 1 {
		t.Fatalf("reporting = %d, want 1", len(lumi.Reporting))
	}
	if lumi.Reporting[0].Cluster != 1026 {
		t.Errorf("reporting cluster = %d, want 1026", lumi.Reporting[0].Cluster)
	}

	ikea := db.Lookup("IKEA of Sweden", "TRADFRI on/off switch")
	if ikea == nil {
		t.Fatal("IKEA device not found")
	}
	if len(ikea.Bind) != 2 {
		t.Errorf("IKEA bind = %v, want [6 8]", ikea.Bind)
	}
}

func TestLoadDeviceDirMissing(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	registry := zcl.NewRegistry(logger)

	// Non-existent directory should return empty DB, no error.
	db, err := LoadDeviceDir("/nonexistent/dir", registry, logger)
	if err != nil {
		t.Fatal(err)
	}
	if db.Len() != 0 {
		t.Errorf("len = %d, want 0", db.Len())
	}
}
