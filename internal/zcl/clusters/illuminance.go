package clusters

import "zigbee-go-home/internal/zcl"

var IlluminanceMeasurement = zcl.ClusterDef{
	ID:   0x0400,
	Name: "Illuminance Measurement",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "MinMeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxMeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "Tolerance", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0004, Name: "LightSensorType", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
	},
}

var IlluminanceLevelSensing = zcl.ClusterDef{
	ID:   0x0401,
	Name: "Illuminance Level Sensing",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "LevelStatus", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "LightSensorType", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0010, Name: "IlluminanceTargetLevel", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
	},
}
