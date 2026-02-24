package clusters

import "zigbee-go-home/internal/zcl"

var ElectricalMeasurement = zcl.ClusterDef{
	ID:   0x0B04,
	Name: "Electrical Measurement",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MeasurementType", Type: zcl.TypeBitmap32, Access: zcl.AccessRead},
		{ID: 0x0505, Name: "RMSVoltage", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0508, Name: "RMSCurrent", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x050B, Name: "ActivePower", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x050E, Name: "PowerFactor", Type: zcl.TypeInt8, Access: zcl.AccessRead},
		{ID: 0x0600, Name: "ACVoltageMultiplier", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0601, Name: "ACVoltageDivisor", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0602, Name: "ACCurrentMultiplier", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0603, Name: "ACCurrentDivisor", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0604, Name: "ACPowerMultiplier", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0605, Name: "ACPowerDivisor", Type: zcl.TypeUint16, Access: zcl.AccessRead},
	},
}
