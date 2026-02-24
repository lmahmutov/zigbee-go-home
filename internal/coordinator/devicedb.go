package coordinator

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"zigbee-go-home/internal/zcl"
)

// PropertySource describes a proprietary attribute that contains multiple sub-values.
type PropertySource struct {
	Cluster   uint16        `json:"cluster"`
	Attribute uint16        `json:"attribute"`
	Decoder   string        `json:"decoder"` // "xiaomi_tlv"
	Values    []PropertyDef `json:"values"`
}

// PropertyDef describes a single named property extracted from a decoded attribute.
type PropertyDef struct {
	Tag       int    `json:"tag"`
	Name      string `json:"name"`
	Transform string `json:"transform,omitempty"`
}

// ManufacturerGroup groups device models under one manufacturer name.
type ManufacturerGroup struct {
	Name   string             `json:"name"`
	Models []DeviceDefinition `json:"models"`
}

// DeviceDefinition describes how to configure a specific device model.
type DeviceDefinition struct {
	Manufacturer string           `json:"manufacturer"`
	Model        string           `json:"model"`
	FriendlyName string           `json:"friendly_name,omitempty"`
	Bind         []uint16         `json:"bind"`
	Reporting    []ReportingEntry `json:"reporting,omitempty"`
	Properties   []PropertySource `json:"properties,omitempty"`
}

// ReportingEntry specifies attribute reporting configuration for a cluster.
type ReportingEntry struct {
	Cluster   uint16 `json:"cluster"`
	Attribute uint16 `json:"attribute"`
	Type      uint8  `json:"type"`
	Min       uint16 `json:"min"`
	Max       uint16 `json:"max"`
	Change    int    `json:"change"`
}

// DeviceDB holds device definitions keyed by manufacturer+model.
type DeviceDB struct {
	defs map[string]*DeviceDefinition
}

func deviceKey(manufacturer, model string) string {
	return manufacturer + "\x00" + model
}

// NewDeviceDB creates an empty device database.
func NewDeviceDB() *DeviceDB {
	return &DeviceDB{defs: make(map[string]*DeviceDefinition)}
}

// Add inserts a device definition into the database.
func (db *DeviceDB) Add(def DeviceDefinition) {
	cp := def
	db.defs[deviceKey(def.Manufacturer, def.Model)] = &cp
}

// Lookup finds a device definition by manufacturer and model.
func (db *DeviceDB) Lookup(manufacturer, model string) *DeviceDefinition {
	return db.defs[deviceKey(manufacturer, model)]
}

// Len returns the number of device definitions.
func (db *DeviceDB) Len() int {
	return len(db.defs)
}

// deviceFile is the JSON structure for files in the devices directory.
type deviceFile struct {
	Clusters      []zcl.ClusterDef    `json:"clusters,omitempty"`
	Devices       []DeviceDefinition  `json:"devices,omitempty"`
	Manufacturers []ManufacturerGroup `json:"manufacturers,omitempty"`
}

// LoadDeviceDir reads all *.json files from a directory, registering custom
// clusters into the ZCL registry and loading device definitions into a DeviceDB.
// Returns an empty DeviceDB (not an error) if the directory doesn't exist or is empty.
func LoadDeviceDir(dir string, registry *zcl.Registry, logger *slog.Logger) (*DeviceDB, error) {
	db := NewDeviceDB()

	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return db, fmt.Errorf("glob devices dir: %w", err)
	}
	if len(matches) == 0 {
		logger.Info("no device definition files found", "dir", dir)
		return db, nil
	}

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return db, fmt.Errorf("read %s: %w", path, err)
		}

		var df deviceFile
		if err := json.Unmarshal(data, &df); err != nil {
			return db, fmt.Errorf("parse %s: %w", path, err)
		}

		for _, c := range df.Clusters {
			registry.Register(c)
		}
		for _, d := range df.Devices {
			db.Add(d)
		}
		for _, mg := range df.Manufacturers {
			for _, d := range mg.Models {
				d.Manufacturer = mg.Name
				db.Add(d)
			}
		}

		deviceCount := len(df.Devices)
		for _, mg := range df.Manufacturers {
			deviceCount += len(mg.Models)
		}
		logger.Info("loaded device file", "path", filepath.Base(path),
			"clusters", len(df.Clusters), "devices", deviceCount)
	}

	logger.Info("device database loaded", "files", len(matches), "devices", db.Len())
	return db, nil
}
