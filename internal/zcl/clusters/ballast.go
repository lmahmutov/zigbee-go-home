package clusters

import "zigbee-go-home/internal/zcl"

var BallastConfiguration = zcl.ClusterDef{
	ID:   0x0301,
	Name: "Ballast Configuration",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "PhysicalMinLevel", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "PhysicalMaxLevel", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "BallastStatus", Type: zcl.TypeBitmap8, Access: zcl.AccessRead},
		{ID: 0x0010, Name: "MinLevel", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0011, Name: "MaxLevel", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0014, Name: "PowerOnLevel", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0015, Name: "PowerOnFadeTime", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0016, Name: "IntrinsicBallastFactor", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0017, Name: "BallastFactorAdjustment", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0020, Name: "LampQuantity", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0030, Name: "LampType", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0031, Name: "LampManufacturer", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0032, Name: "LampRatedHours", Type: zcl.TypeUint24, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0033, Name: "LampBurnHours", Type: zcl.TypeUint24, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0034, Name: "LampAlarmMode", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0035, Name: "LampBurnHoursTripPoint", Type: zcl.TypeUint24, Access: zcl.AccessRead | zcl.AccessWrite},
	},
}
