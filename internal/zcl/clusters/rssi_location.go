package clusters

import "zigbee-go-home/internal/zcl"

var RSSILocation = zcl.ClusterDef{
	ID:   0x000B,
	Name: "RSSI Location",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "LocationType", Type: zcl.TypeBitmap8, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "LocationMethod", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "LocationAge", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "QualityMeasure", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0004, Name: "NumberOfDevices", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0010, Name: "Coordinate1", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0011, Name: "Coordinate2", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0012, Name: "Coordinate3", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0013, Name: "Power", Type: zcl.TypeInt16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0014, Name: "PathLossExponent", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0015, Name: "ReportingPeriod", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0016, Name: "CalculationPeriod", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0017, Name: "NumberRSSIMeasurements", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
}
