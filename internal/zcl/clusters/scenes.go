package clusters

import "zigbee-go-home/internal/zcl"

var Scenes = zcl.ClusterDef{
	ID:   0x0005,
	Name: "Scenes",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "SceneCount", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "CurrentScene", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "CurrentGroup", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0003, Name: "SceneValid", Type: zcl.TypeBool, Access: zcl.AccessRead},
		{ID: 0x0004, Name: "NameSupport", Type: zcl.TypeBitmap8, Access: zcl.AccessRead},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "AddScene", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "ViewScene", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "RemoveScene", Direction: zcl.DirectionToServer},
		{ID: 0x03, Name: "RemoveAllScenes", Direction: zcl.DirectionToServer},
		{ID: 0x04, Name: "StoreScene", Direction: zcl.DirectionToServer},
		{ID: 0x05, Name: "RecallScene", Direction: zcl.DirectionToServer},
		{ID: 0x06, Name: "GetSceneMembership", Direction: zcl.DirectionToServer},
	},
}
