package clusters

import "zigbee-go-home/internal/zcl"

var MultistateInput = zcl.ClusterDef{
	ID:   0x0012,
	Name: "Multistate Input (Basic)",
	Attributes: []zcl.AttributeDef{
		{ID: 0x000E, Name: "StateText", Type: zcl.TypeArray, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x001C, Name: "Description", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x004A, Name: "NumberOfStates", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0051, Name: "OutOfService", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0055, Name: "PresentValue", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite | zcl.AccessReport},
		{ID: 0x0067, Name: "Reliability", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x006F, Name: "StatusFlags", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0100, Name: "ApplicationType", Type: zcl.TypeUint32, Access: zcl.AccessRead},
	},
}

var MultistateOutput = zcl.ClusterDef{
	ID:   0x0013,
	Name: "Multistate Output (Basic)",
	Attributes: []zcl.AttributeDef{
		{ID: 0x000E, Name: "StateText", Type: zcl.TypeArray, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x001C, Name: "Description", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x004A, Name: "NumberOfStates", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0051, Name: "OutOfService", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0055, Name: "PresentValue", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite | zcl.AccessReport},
		{ID: 0x0067, Name: "Reliability", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0068, Name: "RelinquishDefault", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x006F, Name: "StatusFlags", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0100, Name: "ApplicationType", Type: zcl.TypeUint32, Access: zcl.AccessRead},
	},
}

var MultistateValue = zcl.ClusterDef{
	ID:   0x0014,
	Name: "Multistate Value (Basic)",
	Attributes: []zcl.AttributeDef{
		{ID: 0x000E, Name: "StateText", Type: zcl.TypeArray, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x001C, Name: "Description", Type: zcl.TypeCharStr, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x004A, Name: "NumberOfStates", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0051, Name: "OutOfService", Type: zcl.TypeBool, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0055, Name: "PresentValue", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite | zcl.AccessReport},
		{ID: 0x0067, Name: "Reliability", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0068, Name: "RelinquishDefault", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x006F, Name: "StatusFlags", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0100, Name: "ApplicationType", Type: zcl.TypeUint32, Access: zcl.AccessRead},
	},
}
