package clusters

import "zigbee-go-home/internal/zcl"

var Commissioning = zcl.ClusterDef{
	ID:   0x0015,
	Name: "Commissioning",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "ShortAddress", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0001, Name: "ExtendedPANId", Type: zcl.TypeEUI64, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0002, Name: "PANId", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0003, Name: "ChannelMask", Type: zcl.TypeBitmap32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0004, Name: "ProtocolVersion", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0005, Name: "StackProfile", Type: zcl.TypeUint8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0006, Name: "StartupControl", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0010, Name: "TrustCenterAddress", Type: zcl.TypeEUI64, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0021, Name: "NetworkKeyType", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "RestartDevice", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "SaveStartupParameters", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "RestoreStartupParameters", Direction: zcl.DirectionToServer},
		{ID: 0x03, Name: "ResetStartupParameters", Direction: zcl.DirectionToServer},
	},
}
