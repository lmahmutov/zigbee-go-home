package clusters

import "zigbee-go-home/internal/zcl"

var OccupancySensing = zcl.ClusterDef{
	ID:   0x0406,
	Name: "Occupancy Sensing",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "Occupancy", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "OccupancySensorType", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "OccupancySensorTypeBitmap", Type: zcl.TypeBitmap8, Access: zcl.AccessRead},
		{ID: 0x0010, Name: "PIROccupiedToUnoccupiedDelay", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0011, Name: "PIRUnoccupiedToOccupiedDelay", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0012, Name: "PIRUnoccupiedToOccupiedThreshold", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
}
