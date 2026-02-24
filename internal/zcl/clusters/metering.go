package clusters

import "zigbee-go-home/internal/zcl"

var Metering = zcl.ClusterDef{
	ID:   0x0702,
	Name: "Metering",
	Attributes: []zcl.AttributeDef{
		// Reading information set
		{ID: 0x0000, Name: "CurrentSummationDelivered", Type: zcl.TypeUint48, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "CurrentSummationReceived", Type: zcl.TypeUint48, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "CurrentMaxDemandDelivered", Type: zcl.TypeUint48, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "CurrentMaxDemandReceived", Type: zcl.TypeUint48, Access: zcl.AccessRead},
		{ID: 0x0006, Name: "PowerFactor", Type: zcl.TypeInt8, Access: zcl.AccessRead},
		// Meter status
		{ID: 0x0200, Name: "Status", Type: zcl.TypeBitmap8, Access: zcl.AccessRead},
		// Formatting
		{ID: 0x0300, Name: "UnitOfMeasure", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0301, Name: "Multiplier", Type: zcl.TypeUint24, Access: zcl.AccessRead},
		{ID: 0x0302, Name: "Divisor", Type: zcl.TypeUint24, Access: zcl.AccessRead},
		{ID: 0x0303, Name: "SummationFormatting", Type: zcl.TypeBitmap8, Access: zcl.AccessRead},
		{ID: 0x0306, Name: "MeteringDeviceType", Type: zcl.TypeBitmap8, Access: zcl.AccessRead},
		// Instantaneous demand
		{ID: 0x0400, Name: "InstantaneousDemand", Type: zcl.TypeInt24, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0401, Name: "CurrentDayConsumptionDelivered", Type: zcl.TypeUint24, Access: zcl.AccessRead},
		{ID: 0x0402, Name: "CurrentDayConsumptionReceived", Type: zcl.TypeUint24, Access: zcl.AccessRead},
		{ID: 0x0403, Name: "PreviousDayConsumptionDelivered", Type: zcl.TypeUint24, Access: zcl.AccessRead},
		{ID: 0x0404, Name: "PreviousDayConsumptionReceived", Type: zcl.TypeUint24, Access: zcl.AccessRead},
	},
}
