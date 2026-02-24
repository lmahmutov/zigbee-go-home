package clusters

import "zigbee-go-home/internal/zcl"

var LevelControl = zcl.ClusterDef{
	ID:   0x0008,
	Name: "Level Control",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "CurrentLevel", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "RemainingTime", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x000F, Name: "Options", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0010, Name: "OnOffTransitionTime", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0011, Name: "OnLevel", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x4000, Name: "StartUpCurrentLevel", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "MoveToLevel", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "Move", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "Step", Direction: zcl.DirectionToServer},
		{ID: 0x03, Name: "Stop", Direction: zcl.DirectionToServer},
		{ID: 0x04, Name: "MoveToLevelWithOnOff", Direction: zcl.DirectionToServer},
		{ID: 0x05, Name: "MoveWithOnOff", Direction: zcl.DirectionToServer},
		{ID: 0x06, Name: "StepWithOnOff", Direction: zcl.DirectionToServer},
		{ID: 0x07, Name: "StopWithOnOff", Direction: zcl.DirectionToServer},
	},
}
