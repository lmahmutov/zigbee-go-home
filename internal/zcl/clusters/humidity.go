package clusters

import "zigbee-go-home/internal/zcl"

var RelativeHumidity = zcl.ClusterDef{
	ID:   0x0405,
	Name: "Relative Humidity",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "MinMeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxMeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "Tolerance", Type: zcl.TypeUint16, Access: zcl.AccessRead},
	},
}
