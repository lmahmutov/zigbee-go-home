package clusters

import "zigbee-go-home/internal/zcl"

var TemperatureMeasurement = zcl.ClusterDef{
	ID:   0x0402,
	Name: "Temperature Measurement",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MeasuredValue", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "MinMeasuredValue", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxMeasuredValue", Type: zcl.TypeInt16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "Tolerance", Type: zcl.TypeUint16, Access: zcl.AccessRead},
	},
}
