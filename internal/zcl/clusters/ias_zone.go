package clusters

import "zigbee-go-home/internal/zcl"

var IASZone = zcl.ClusterDef{
	ID:   0x0500,
	Name: "IAS Zone",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "ZoneState", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "ZoneType", Type: zcl.TypeEnum16, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "ZoneStatus", Type: zcl.TypeBitmap16, Access: zcl.AccessRead | zcl.AccessReport},
		{ID: 0x0010, Name: "IASCIEAddress", Type: zcl.TypeEUI64, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0011, Name: "ZoneID", Type: zcl.TypeUint8, Access: zcl.AccessRead},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "ZoneEnrollResponse", Direction: zcl.DirectionToServer},
		{ID: 0x00, Name: "ZoneStatusChangeNotification", Direction: zcl.DirectionToClient},
		{ID: 0x01, Name: "ZoneEnrollRequest", Direction: zcl.DirectionToClient},
	},
}
