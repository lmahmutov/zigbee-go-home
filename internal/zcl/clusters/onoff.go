package clusters

import "zigbee-go-home/internal/zcl"

var OnOff = zcl.ClusterDef{
	ID:   0x0006,
	Name: "On/Off",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "OnOff", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x4000, Name: "GlobalSceneControl", Type: zcl.TypeBool, Access: zcl.AccessRead},
		{ID: 0x4001, Name: "OnTime", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x4002, Name: "OffWaitTime", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x4003, Name: "StartUpOnOff", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "Off", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "On", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "Toggle", Direction: zcl.DirectionToServer},
		{ID: 0x40, Name: "OffWithEffect", Direction: zcl.DirectionToServer},
		{ID: 0x41, Name: "OnWithRecallGlobalScene", Direction: zcl.DirectionToServer},
		{ID: 0x42, Name: "OnWithTimedOff", Direction: zcl.DirectionToServer},
	},
}
