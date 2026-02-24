package clusters

import "zigbee-go-home/internal/zcl"

var FanControl = zcl.ClusterDef{
	ID:   0x0202,
	Name: "Fan Control",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "FanMode", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0001, Name: "FanModeSequence", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
}
