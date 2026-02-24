package clusters

import "zigbee-go-home/internal/zcl"

var PowerConfiguration = zcl.ClusterDef{
	ID:   0x0001,
	Name: "Power Configuration",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MainsVoltage", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "MainsFrequency", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0020, Name: "BatteryVoltage", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0021, Name: "BatteryPercentageRemaining", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0030, Name: "BatteryManufacturer", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0031, Name: "BatterySize", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0033, Name: "BatteryQuantity", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0034, Name: "BatteryRatedVoltage", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0035, Name: "BatteryAlarmMask", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0036, Name: "BatteryVoltageMinThreshold", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
}
