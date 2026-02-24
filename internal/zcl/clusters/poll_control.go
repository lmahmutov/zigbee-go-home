package clusters

import "zigbee-go-home/internal/zcl"

var PollControl = zcl.ClusterDef{
	ID:   0x0020,
	Name: "Poll Control",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "CheckInInterval", Type: zcl.TypeUint32, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0001, Name: "LongPollInterval", Type: zcl.TypeUint32, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "ShortPollInterval", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "FastPollTimeout", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0004, Name: "CheckInIntervalMin", Type: zcl.TypeUint32, Access: zcl.AccessRead},
		{ID: 0x0005, Name: "LongPollIntervalMin", Type: zcl.TypeUint32, Access: zcl.AccessRead},
		{ID: 0x0006, Name: "FastPollTimeoutMax", Type: zcl.TypeUint16, Access: zcl.AccessRead},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "CheckInResponse", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "FastPollStop", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "SetLongPollInterval", Direction: zcl.DirectionToServer},
		{ID: 0x03, Name: "SetShortPollInterval", Direction: zcl.DirectionToServer},
		{ID: 0x00, Name: "CheckIn", Direction: zcl.DirectionToClient},
	},
}
