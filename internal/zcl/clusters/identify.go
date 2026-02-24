package clusters

import "zigbee-go-home/internal/zcl"

var Identify = zcl.ClusterDef{
	ID:   0x0003,
	Name: "Identify",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "IdentifyTime", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "Identify", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "IdentifyQuery", Direction: zcl.DirectionToServer},
		{ID: 0x40, Name: "TriggerEffect", Direction: zcl.DirectionToServer},
	},
}
