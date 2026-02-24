package clusters

import "zigbee-go-home/internal/zcl"

var BinaryInput = zcl.ClusterDef{
	ID:   0x000F,
	Name: "Binary Input (Basic)",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0004, Name: "ActiveText", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x001C, Name: "Description", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x002E, Name: "InactiveText", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0051, Name: "OutOfService", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0054, Name: "Polarity", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0055, Name: "PresentValue", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite | zcl.AccessReport},
		{ID: 0x0067, Name: "Reliability", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x006F, Name: "StatusFlags", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0100, Name: "ApplicationType", Type: zcl.TypeUint32, Access: zcl.AccessRead},
	},
}

var BinaryOutput = zcl.ClusterDef{
	ID:   0x0010,
	Name: "Binary Output (Basic)",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0004, Name: "ActiveText", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x001C, Name: "Description", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x002E, Name: "InactiveText", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0042, Name: "MinimumOffTime", Type: zcl.TypeUint32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0043, Name: "MinimumOnTime", Type: zcl.TypeUint32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0051, Name: "OutOfService", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0054, Name: "Polarity", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0055, Name: "PresentValue", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite | zcl.AccessReport},
		{ID: 0x0067, Name: "Reliability", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0068, Name: "RelinquishDefault", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x006F, Name: "StatusFlags", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0100, Name: "ApplicationType", Type: zcl.TypeUint32, Access: zcl.AccessRead},
	},
}

var BinaryValue = zcl.ClusterDef{
	ID:   0x0011,
	Name: "Binary Value (Basic)",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0004, Name: "ActiveText", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x001C, Name: "Description", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x002E, Name: "InactiveText", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0042, Name: "MinimumOffTime", Type: zcl.TypeUint32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0043, Name: "MinimumOnTime", Type: zcl.TypeUint32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0051, Name: "OutOfService", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0055, Name: "PresentValue", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite | zcl.AccessReport},
		{ID: 0x0067, Name: "Reliability", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0068, Name: "RelinquishDefault", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x006F, Name: "StatusFlags", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0100, Name: "ApplicationType", Type: zcl.TypeUint32, Access: zcl.AccessRead},
	},
}
