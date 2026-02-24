package clusters

import "zigbee-go-home/internal/zcl"

var Basic = zcl.ClusterDef{
	ID:   0x0000,
	Name: "Basic",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "ZCLVersion", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "ApplicationVersion", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "StackVersion", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "HWVersion", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0004, Name: "ManufacturerName", Type: zcl.TypeCharStr, Access: zcl.AccessRead},
		{ID: 0x0005, Name: "ModelIdentifier", Type: zcl.TypeCharStr, Access: zcl.AccessRead},
		{ID: 0x0006, Name: "DateCode", Type: zcl.TypeCharStr, Access: zcl.AccessRead},
		{ID: 0x0007, Name: "PowerSource", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x4000, Name: "SWBuildID", Type: zcl.TypeCharStr, Access: zcl.AccessRead},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "ResetToFactoryDefaults", Direction: zcl.DirectionToServer},
	},
}
