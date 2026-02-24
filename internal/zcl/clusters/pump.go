package clusters

import "zigbee-go-home/internal/zcl"

var PumpConfigurationAndControl = zcl.ClusterDef{
	ID:   0x0200,
	Name: "Pump Configuration and Control",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MaxPressure", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "MaxSpeed", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxFlow", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "MinConstPressure", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0004, Name: "MaxConstPressure", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0005, Name: "MinCompPressure", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0006, Name: "MaxCompPressure", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0007, Name: "MinConstSpeed", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0008, Name: "MaxConstSpeed", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0009, Name: "MinConstFlow", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x000A, Name: "MaxConstFlow", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x000B, Name: "MinConstTemp", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x000C, Name: "MaxConstTemp", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0010, Name: "PumpStatus", Type: zcl.TypeBitmap16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0011, Name: "EffectiveOperationMode", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0012, Name: "EffectiveControlMode", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0013, Name: "Capacity", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0014, Name: "Speed", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0020, Name: "OperationMode", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0021, Name: "ControlMode", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0022, Name: "AlarmMask", Type: zcl.TypeBitmap16, Access: zcl.AccessRead},
	},
}
