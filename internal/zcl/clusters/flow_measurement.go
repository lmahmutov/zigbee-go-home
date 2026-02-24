package clusters

import "zigbee-go-home/internal/zcl"

var FlowMeasurement = zcl.ClusterDef{
	ID:   0x0404,
	Name: "Flow Measurement",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "MinMeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxMeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "Tolerance", Type: zcl.TypeUint16, Access: zcl.AccessRead},
	},
}

var SoilMoisture = zcl.ClusterDef{
	ID:   0x0408,
	Name: "Soil Moisture",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "MinMeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxMeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "Tolerance", Type: zcl.TypeUint16, Access: zcl.AccessRead},
	},
}

var PHMeasurement = zcl.ClusterDef{
	ID:   0x0409,
	Name: "pH Measurement",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "MinMeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxMeasuredValue", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "Tolerance", Type: zcl.TypeUint16, Access: zcl.AccessRead},
	},
}

var CarbonMonoxide = zcl.ClusterDef{
	ID:   0x040C,
	Name: "Carbon Monoxide (CO) Measurement",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "MinMeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxMeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "Tolerance", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
	},
}

var CarbonDioxide = zcl.ClusterDef{
	ID:   0x040D,
	Name: "Carbon Dioxide (CO2) Measurement",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "MinMeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxMeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "Tolerance", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
	},
}

var PM25Measurement = zcl.ClusterDef{
	ID:   0x042A,
	Name: "PM2.5 Measurement",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "MinMeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxMeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "Tolerance", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
	},
}

var FormaldehydeMeasurement = zcl.ClusterDef{
	ID:   0x042B,
	Name: "Formaldehyde Measurement",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "MinMeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "MaxMeasuredValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "Tolerance", Type: zcl.TypeFloat32, Access: zcl.AccessRead},
	},
}
