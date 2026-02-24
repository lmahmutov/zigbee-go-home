package clusters

import "zigbee-go-home/internal/zcl"

var TouchlinkCommissioning = zcl.ClusterDef{
	ID:   0x1000,
	Name: "Touchlink Commissioning",
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "ScanRequest", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "DeviceInformationRequest", Direction: zcl.DirectionToServer},
		{ID: 0x06, Name: "IdentifyRequest", Direction: zcl.DirectionToServer},
		{ID: 0x07, Name: "ResetToFactoryNewRequest", Direction: zcl.DirectionToServer},
		{ID: 0x10, Name: "NetworkStartRequest", Direction: zcl.DirectionToServer},
		{ID: 0x12, Name: "NetworkJoinRouterRequest", Direction: zcl.DirectionToServer},
		{ID: 0x14, Name: "NetworkJoinEndDeviceRequest", Direction: zcl.DirectionToServer},
		{ID: 0x16, Name: "NetworkUpdateRequest", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "ScanResponse", Direction: zcl.DirectionToClient},
		{ID: 0x03, Name: "DeviceInformationResponse", Direction: zcl.DirectionToClient},
		{ID: 0x11, Name: "NetworkStartResponse", Direction: zcl.DirectionToClient},
		{ID: 0x13, Name: "NetworkJoinRouterResponse", Direction: zcl.DirectionToClient},
		{ID: 0x15, Name: "NetworkJoinEndDeviceResponse", Direction: zcl.DirectionToClient},
	},
}
