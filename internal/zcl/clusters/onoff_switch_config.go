package clusters

import "zigbee-go-home/internal/zcl"

var OnOffSwitchConfiguration = zcl.ClusterDef{
	ID:   0x0007,
	Name: "On/Off Switch Configuration",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "SwitchType", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0010, Name: "SwitchActions", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
}
