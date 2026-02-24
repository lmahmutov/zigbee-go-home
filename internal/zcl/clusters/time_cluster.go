package clusters

import "zigbee-go-home/internal/zcl"

var Time = zcl.ClusterDef{
	ID:   0x000A,
	Name: "Time",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "Time", Type: zcl.TypeUTC, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0001, Name: "TimeStatus", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0002, Name: "TimeZone", Type: zcl.TypeInt32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0007, Name: "LocalTime", Type: zcl.TypeUint32, Access: zcl.AccessRead},
	},
}
