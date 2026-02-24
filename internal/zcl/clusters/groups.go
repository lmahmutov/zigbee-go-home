package clusters

import "zigbee-go-home/internal/zcl"

var Groups = zcl.ClusterDef{
	ID:   0x0004,
	Name: "Groups",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "NameSupport", Type: zcl.TypeBitmap8, Access: zcl.AccessRead},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "AddGroup", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "ViewGroup", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "GetGroupMembership", Direction: zcl.DirectionToServer},
		{ID: 0x03, Name: "RemoveGroup", Direction: zcl.DirectionToServer},
		{ID: 0x04, Name: "RemoveAllGroups", Direction: zcl.DirectionToServer},
		{ID: 0x05, Name: "AddGroupIfIdentifying", Direction: zcl.DirectionToServer},
	},
}
