package clusters

import "zigbee-go-home/internal/zcl"

var DoorLock = zcl.ClusterDef{
	ID:   0x0101,
	Name: "Door Lock",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "LockState", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0001, Name: "LockType", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "ActuatorEnabled", Type: zcl.TypeBool, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "DoorState", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessReport},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "LockDoor", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "UnlockDoor", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "Toggle", Direction: zcl.DirectionToServer},
	},
}
