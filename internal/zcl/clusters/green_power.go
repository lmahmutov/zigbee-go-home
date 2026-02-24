package clusters

import "zigbee-go-home/internal/zcl"

var GreenPower = zcl.ClusterDef{
	ID:   0x0021,
	Name: "Green Power",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "MaxSinkTableEntries", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "SinkTable", Type: zcl.TypeOctetStr, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "CommunicationMode", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0003, Name: "CommissioningExitMode", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0004, Name: "CommissioningWindow", Type: zcl.TypeUint16, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0005, Name: "SecurityLevel", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0006, Name: "Functionality", Type: zcl.TypeBitmap24, Access: zcl.AccessRead},
		{ID: 0x0007, Name: "ActiveFunctionality", Type: zcl.TypeBitmap24, Access: zcl.AccessRead},
		// Proxy side
		{ID: 0x0010, Name: "MaxProxyTableEntries", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x0011, Name: "ProxyTable", Type: zcl.TypeOctetStr, Access: zcl.AccessRead},
		{ID: 0x0016, Name: "SharedSecurityKeyType", Type: zcl.TypeBitmap8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0022, Name: "LinkKey", Type: zcl.TypeOctetStr, Access: zcl.AccessRead | zcl.AccessWrite},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "GPNotification", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "GPPairingSearch", Direction: zcl.DirectionToServer},
		{ID: 0x04, Name: "GPCommissioningNotification", Direction: zcl.DirectionToServer},
		{ID: 0x00, Name: "GPNotificationResponse", Direction: zcl.DirectionToClient},
		{ID: 0x01, Name: "GPPairing", Direction: zcl.DirectionToClient},
		{ID: 0x02, Name: "GPProxyCommissioningMode", Direction: zcl.DirectionToClient},
		{ID: 0x06, Name: "GPResponse", Direction: zcl.DirectionToClient},
	},
}
