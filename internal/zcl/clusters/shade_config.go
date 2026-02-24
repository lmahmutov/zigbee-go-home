package clusters

import "zigbee-go-home/internal/zcl"

var ShadeConfiguration = zcl.ClusterDef{
	ID:   0x0100,
	Name: "Shade Configuration",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "PhysicalClosedLimit", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "MotorStepSize", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "Status", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0010, Name: "ClosedLimit", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0011, Name: "Mode", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
}

var BarrierControl = zcl.ClusterDef{
	ID:   0x0103,
	Name: "Barrier Control",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0001, Name: "MovingState", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "SafetyStatus", Type: zcl.TypeBitmap16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "Capabilities", Type: zcl.TypeBitmap8, Access: zcl.AccessRead},
		{ID: 0x000A, Name: "BarrierPosition", Type: zcl.TypeUint8, Access: zcl.AccessRead},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "GoToPercent", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "Stop", Direction: zcl.DirectionToServer},
	},
}
