package clusters

import "zigbee-go-home/internal/zcl"

var DeviceTemperatureConfiguration = zcl.ClusterDef{
	ID:   0x0002,
	Name: "Device Temperature Configuration",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "CurrentTemperature", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "MinTempExperienced", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxTempExperienced", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "OverTempTotalDwell", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0010, Name: "DeviceTempAlarmMask", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0011, Name: "LowTempThreshold", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0012, Name: "HighTempThreshold", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0013, Name: "LowTempDwellTripPoint", Type: zcl.TypeUint24, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0014, Name: "HighTempDwellTripPoint", Type: zcl.TypeUint24, Access: zcl.AccessRead | zcl.AccessWrite},
	},
}
