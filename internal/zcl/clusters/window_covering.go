package clusters

import "zigbee-go-home/internal/zcl"

var WindowCovering = zcl.ClusterDef{
	ID:   0x0102,
	Name: "Window Covering",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "WindowCoveringType", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "CurrentPositionLiftPercent", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0004, Name: "CurrentPositionTiltPercent", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0007, Name: "ConfigStatus", Type: zcl.TypeBitmap8, Access: zcl.AccessRead},
		{ID: 0x0008, Name: "CurrentPositionLiftPercentage", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0009, Name: "CurrentPositionTiltPercentage", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0017, Name: "Mode", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "UpOpen", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "DownClose", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "Stop", Direction: zcl.DirectionToServer},
		{ID: 0x04, Name: "GoToLiftValue", Direction: zcl.DirectionToServer},
		{ID: 0x05, Name: "GoToLiftPercentage", Direction: zcl.DirectionToServer},
		{ID: 0x07, Name: "GoToTiltValue", Direction: zcl.DirectionToServer},
		{ID: 0x08, Name: "GoToTiltPercentage", Direction: zcl.DirectionToServer},
	},
}
