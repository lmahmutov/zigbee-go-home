package clusters

import "zigbee-go-home/internal/zcl"

var OTAUpgrade = zcl.ClusterDef{
	ID:   0x0019,
	Name: "OTA Upgrade",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "UpgradeServerID", Type: zcl.TypeEUI64, Access: zcl.AccessRead},
		{ID: 0x0002, Name: "CurrentFileVersion", Type: zcl.TypeUint32, Access: zcl.AccessRead},
		{ID: 0x0004, Name: "DownloadedFileVersion", Type: zcl.TypeUint32, Access: zcl.AccessRead},
		{ID: 0x0006, Name: "ImageUpgradeStatus", Type: zcl.TypeEnum8, Access: zcl.AccessRead},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x01, Name: "QueryNextImageRequest", Direction: zcl.DirectionToServer},
		{ID: 0x03, Name: "ImageBlockRequest", Direction: zcl.DirectionToServer},
		{ID: 0x06, Name: "UpgradeEndRequest", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "QueryNextImageResponse", Direction: zcl.DirectionToClient},
		{ID: 0x05, Name: "ImageBlockResponse", Direction: zcl.DirectionToClient},
		{ID: 0x07, Name: "UpgradeEndResponse", Direction: zcl.DirectionToClient},
	},
}
