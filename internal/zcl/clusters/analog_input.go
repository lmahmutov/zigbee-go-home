package clusters

import "zigbee-go-home/internal/zcl"

var AnalogInput = zcl.ClusterDef{
	ID:   0x000C,
	Name: "Analog Input (Basic)",
	Attributes: []zcl.AttributeDef{
		{ID: 0x001C, Name: "Description", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0041, Name: "MaxPresentValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0045, Name: "MinPresentValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0051, Name: "OutOfService", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0055, Name: "PresentValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessWrite | zcl.AccessReport},
		{ID: 0x0067, Name: "Reliability", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x006A, Name: "Resolution", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x006F, Name: "StatusFlags", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0075, Name: "EngineeringUnits", Type: zcl.TypeEnum16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0100, Name: "ApplicationType", Type: zcl.TypeUint32, Access: zcl.AccessRead},
	},
}

var AnalogOutput = zcl.ClusterDef{
	ID:   0x000D,
	Name: "Analog Output (Basic)",
	Attributes: []zcl.AttributeDef{
		{ID: 0x001C, Name: "Description", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0041, Name: "MaxPresentValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0045, Name: "MinPresentValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0051, Name: "OutOfService", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0055, Name: "PresentValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessWrite | zcl.AccessReport},
		{ID: 0x0067, Name: "Reliability", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0068, Name: "RelinquishDefault", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x006A, Name: "Resolution", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x006F, Name: "StatusFlags", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0075, Name: "EngineeringUnits", Type: zcl.TypeEnum16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0100, Name: "ApplicationType", Type: zcl.TypeUint32, Access: zcl.AccessRead},
	},
}

var AnalogValue = zcl.ClusterDef{
	ID:   0x000E,
	Name: "Analog Value (Basic)",
	Attributes: []zcl.AttributeDef{
		{ID: 0x001C, Name: "Description", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0051, Name: "OutOfService", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0055, Name: "PresentValue", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessWrite | zcl.AccessReport},
		{ID: 0x0067, Name: "Reliability", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0068, Name: "RelinquishDefault", Type: zcl.TypeFloat32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x006F, Name: "StatusFlags", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0075, Name: "EngineeringUnits", Type: zcl.TypeEnum16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0100, Name: "ApplicationType", Type: zcl.TypeUint32, Access: zcl.AccessRead},
	},
}
