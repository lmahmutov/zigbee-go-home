package store

import "time"

// Device represents a Zigbee device.
type Device struct {
	IEEEAddress  string            `json:"ieee_address"`
	ShortAddress uint16            `json:"short_address"`
	Manufacturer string            `json:"manufacturer,omitempty"`
	Model        string            `json:"model,omitempty"`
	FriendlyName string            `json:"friendly_name,omitempty"`
	Endpoints    []Endpoint        `json:"endpoints,omitempty"`
	Interviewed  bool              `json:"interviewed"`
	JoinedAt     time.Time         `json:"joined_at"`
	LastSeen     time.Time         `json:"last_seen"`
	LQI          uint8             `json:"lqi,omitempty"`
	RSSI         int8              `json:"rssi,omitempty"`
	Properties   map[string]any    `json:"properties,omitempty"`
}

// Endpoint represents a device endpoint.
type Endpoint struct {
	ID          uint8    `json:"id"`
	ProfileID   uint16   `json:"profile_id"`
	DeviceID    uint16   `json:"device_id"`
	InClusters  []uint16 `json:"in_clusters"`
	OutClusters []uint16 `json:"out_clusters"`
}

// NetworkState holds persisted network configuration.
// NetworkKey is hidden from API/JSON serialization via json:"-".
type NetworkState struct {
	Channel    uint8  `json:"channel"`
	PanID      uint16 `json:"pan_id"`
	ExtPanID   string `json:"ext_pan_id"`
	NetworkKey string `json:"-"`
	Formed     bool   `json:"formed"`
}

// networkStateStorage is the internal struct used for DB serialization,
// preserving the network key on disk.
type networkStateStorage struct {
	Channel    uint8  `json:"channel"`
	PanID      uint16 `json:"pan_id"`
	ExtPanID   string `json:"ext_pan_id"`
	NetworkKey string `json:"network_key,omitempty"`
	Formed     bool   `json:"formed"`
}
