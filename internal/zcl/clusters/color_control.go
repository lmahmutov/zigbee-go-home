package clusters

import "zigbee-go-home/internal/zcl"

var ColorControl = zcl.ClusterDef{
	ID:   0x0300,
	Name: "Color Control",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "CurrentHue", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "CurrentSaturation", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0002, Name: "RemainingTime", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "CurrentX", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0004, Name: "CurrentY", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0007, Name: "ColorTemperatureMireds", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0008, Name: "ColorMode", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x000F, Name: "Options", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x4001, Name: "EnhancedCurrentHue", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x4002, Name: "EnhancedColorMode", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x400A, Name: "ColorCapabilities", Type: zcl.TypeBitmap16, Access: zcl.AccessRead},
		{ID: 0x400B, Name: "ColorTempPhysicalMinMireds", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x400C, Name: "ColorTempPhysicalMaxMireds", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x400D, Name: "CoupleColorTempToLevelMinMireds", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x4010, Name: "StartUpColorTemperatureMireds", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "MoveToHue", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "MoveHue", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "StepHue", Direction: zcl.DirectionToServer},
		{ID: 0x03, Name: "MoveToSaturation", Direction: zcl.DirectionToServer},
		{ID: 0x04, Name: "MoveSaturation", Direction: zcl.DirectionToServer},
		{ID: 0x05, Name: "StepSaturation", Direction: zcl.DirectionToServer},
		{ID: 0x06, Name: "MoveToHueAndSaturation", Direction: zcl.DirectionToServer},
		{ID: 0x07, Name: "MoveToColor", Direction: zcl.DirectionToServer},
		{ID: 0x08, Name: "MoveColor", Direction: zcl.DirectionToServer},
		{ID: 0x09, Name: "StepColor", Direction: zcl.DirectionToServer},
		{ID: 0x0A, Name: "MoveToColorTemperature", Direction: zcl.DirectionToServer},
		{ID: 0x47, Name: "StopMoveStep", Direction: zcl.DirectionToServer},
		{ID: 0x4B, Name: "MoveColorTemperature", Direction: zcl.DirectionToServer},
		{ID: 0x4C, Name: "StepColorTemperature", Direction: zcl.DirectionToServer},
	},
}
