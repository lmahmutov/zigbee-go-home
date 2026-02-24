package clusters

import "zigbee-go-home/internal/zcl"

var Diagnostics = zcl.ClusterDef{
	ID:   0x0B05,
	Name: "Diagnostics",
	Attributes: []zcl.AttributeDef{
		// Hardware information
		{ID: 0x0000, Name: "NumberOfResets", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "PersistentMemoryWrites", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		// Stack/Network information
		{ID: 0x0100, Name: "MacRxBcast", Type: zcl.TypeUint32, Access: zcl.AccessRead},
		{ID: 0x0101, Name: "MacTxBcast", Type: zcl.TypeUint32, Access: zcl.AccessRead},
		{ID: 0x0102, Name: "MacRxUcast", Type: zcl.TypeUint32, Access: zcl.AccessRead},
		{ID: 0x0103, Name: "MacTxUcast", Type: zcl.TypeUint32, Access: zcl.AccessRead},
		{ID: 0x0104, Name: "MacTxUcastRetry", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0105, Name: "MacTxUcastFail", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0106, Name: "APSRxBcast", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0107, Name: "APSTxBcast", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0108, Name: "APSRxUcast", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0109, Name: "APSTxUcastSuccess", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x010A, Name: "APSTxUcastRetry", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x010B, Name: "APSTxUcastFail", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x010C, Name: "RouteDiscInitiated", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x010D, Name: "NeighborAdded", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x010E, Name: "NeighborRemoved", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x010F, Name: "NeighborStale", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0110, Name: "JoinIndication", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0111, Name: "ChildMoved", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0112, Name: "NWKFCFailure", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0113, Name: "APSFCFailure", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0114, Name: "APSUnauthorizedKey", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0115, Name: "NWKDecryptFailures", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0116, Name: "APSDecryptFailures", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0117, Name: "PacketBufferAllocateFailures", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0118, Name: "RelayedUcast", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0119, Name: "PhyToMACQueueLimitReached", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x011A, Name: "PacketValidateDropCount", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x011B, Name: "AverageMACRetryPerAPSMessageSent", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x011C, Name: "LastMessageLQI", Type: zcl.TypeUint8, Access: zcl.AccessRead},
		{ID: 0x011D, Name: "LastMessageRSSI", Type: zcl.TypeInt8, Access: zcl.AccessRead},
	},
}
