package clusters

import "zigbee-go-home/internal/zcl"

var PowerProfile = zcl.ClusterDef{
	ID:   0x001A,
	Name: "Power Profile",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "TotalProfileNum", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "MultipleScheduling", Type: zcl.TypeBool, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "EnergyFormatting", Type: zcl.TypeBitmap8, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "EnergyRemote", Type: zcl.TypeBool, Access: zcl.AccessRead},
		{ID: 0x0004, Name: "ScheduleMode", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "PowerProfileRequest", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "PowerProfileStateRequest", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "GetPowerProfilePriceResponse", Direction: zcl.DirectionToServer},
		{ID: 0x03, Name: "GetOverallSchedulePriceResponse", Direction: zcl.DirectionToServer},
		{ID: 0x04, Name: "EnergyPhasesScheduleNotification", Direction: zcl.DirectionToServer},
		{ID: 0x05, Name: "EnergyPhasesScheduleResponse", Direction: zcl.DirectionToServer},
		{ID: 0x06, Name: "PowerProfileScheduleConstraintsRequest", Direction: zcl.DirectionToServer},
		{ID: 0x07, Name: "EnergyPhasesScheduleStateRequest", Direction: zcl.DirectionToServer},
		{ID: 0x00, Name: "PowerProfileNotification", Direction: zcl.DirectionToClient},
		{ID: 0x01, Name: "PowerProfileResponse", Direction: zcl.DirectionToClient},
		{ID: 0x02, Name: "PowerProfileStateResponse", Direction: zcl.DirectionToClient},
		{ID: 0x03, Name: "GetPowerProfilePrice", Direction: zcl.DirectionToClient},
		{ID: 0x04, Name: "PowerProfileStateNotification", Direction: zcl.DirectionToClient},
		{ID: 0x05, Name: "GetOverallSchedulePrice", Direction: zcl.DirectionToClient},
		{ID: 0x06, Name: "EnergyPhasesScheduleRequest", Direction: zcl.DirectionToClient},
		{ID: 0x07, Name: "EnergyPhasesScheduleStateResponse", Direction: zcl.DirectionToClient},
		{ID: 0x08, Name: "EnergyPhasesScheduleStateNotification", Direction: zcl.DirectionToClient},
		{ID: 0x09, Name: "PowerProfileScheduleConstraintsNotification", Direction: zcl.DirectionToClient},
		{ID: 0x0A, Name: "PowerProfileScheduleConstraintsResponse", Direction: zcl.DirectionToClient},
	},
}

var ApplianceControl = zcl.ClusterDef{
	ID:   0x001B,
	Name: "Appliance Control",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "StartTime", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "FinishTime", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "RemainingTime", Type: zcl.TypeUint16, Access: zcl.AccessRead},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "ExecutionOfACommand", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "SignalState", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "WriteFunctions", Direction: zcl.DirectionToServer},
		{ID: 0x03, Name: "OverloadPauseResume", Direction: zcl.DirectionToServer},
		{ID: 0x04, Name: "OverloadPause", Direction: zcl.DirectionToServer},
		{ID: 0x05, Name: "OverloadWarning", Direction: zcl.DirectionToServer},
		{ID: 0x00, Name: "SignalStateResponse", Direction: zcl.DirectionToClient},
		{ID: 0x01, Name: "SignalStateNotification", Direction: zcl.DirectionToClient},
	},
}
