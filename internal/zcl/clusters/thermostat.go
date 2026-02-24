package clusters

import "zigbee-go-home/internal/zcl"

var Thermostat = zcl.ClusterDef{
	ID:   0x0201,
	Name: "Thermostat",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "LocalTemperature", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0003, Name: "AbsMinHeatSetpointLimit", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0004, Name: "AbsMaxHeatSetpointLimit", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0005, Name: "AbsMinCoolSetpointLimit", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0006, Name: "AbsMaxCoolSetpointLimit", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0011, Name: "OccupiedCoolingSetpoint", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0012, Name: "OccupiedHeatingSetpoint", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x001B, Name: "ControlSequenceOfOperation", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x001C, Name: "SystemMode", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x001E, Name: "RunningMode", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0029, Name: "RunningState", Type: zcl.TypeBitmap16, Access: zcl.AccessRead},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "SetpointRaiseLower", Direction: zcl.DirectionToServer},
	},
}
