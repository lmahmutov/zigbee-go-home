package clusters

import "zigbee-go-home/internal/zcl"

var IASACE = zcl.ClusterDef{
	ID:   0x0501,
	Name: "IAS ACE",
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "Arm", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "Bypass", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "Emergency", Direction: zcl.DirectionToServer},
		{ID: 0x03, Name: "Fire", Direction: zcl.DirectionToServer},
		{ID: 0x04, Name: "Panic", Direction: zcl.DirectionToServer},
		{ID: 0x05, Name: "GetZoneIDMap", Direction: zcl.DirectionToServer},
		{ID: 0x06, Name: "GetZoneInformation", Direction: zcl.DirectionToServer},
		{ID: 0x07, Name: "GetPanelStatus", Direction: zcl.DirectionToServer},
		{ID: 0x08, Name: "GetBypassedZoneList", Direction: zcl.DirectionToServer},
		{ID: 0x09, Name: "GetZoneStatus", Direction: zcl.DirectionToServer},
		{ID: 0x00, Name: "ArmResponse", Direction: zcl.DirectionToClient},
		{ID: 0x01, Name: "GetZoneIDMapResponse", Direction: zcl.DirectionToClient},
		{ID: 0x02, Name: "GetZoneInformationResponse", Direction: zcl.DirectionToClient},
		{ID: 0x03, Name: "ZoneStatusChanged", Direction: zcl.DirectionToClient},
		{ID: 0x04, Name: "PanelStatusChanged", Direction: zcl.DirectionToClient},
		{ID: 0x05, Name: "GetPanelStatusResponse", Direction: zcl.DirectionToClient},
		{ID: 0x06, Name: "SetBypassedZoneList", Direction: zcl.DirectionToClient},
		{ID: 0x07, Name: "BypassResponse", Direction: zcl.DirectionToClient},
		{ID: 0x08, Name: "GetZoneStatusResponse", Direction: zcl.DirectionToClient},
	},
}

var IASWD = zcl.ClusterDef{
	ID:   0x0502,
	Name: "IAS Warning Device",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MaxDuration", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "StartWarning", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "Squawk", Direction: zcl.DirectionToServer},
	},
}
