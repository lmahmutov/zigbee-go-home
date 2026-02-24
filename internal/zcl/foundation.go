package zcl

// Foundation ZCL command IDs (global, not cluster-specific).
const (
	FoundationReadAttributes         uint8 = 0x00
	FoundationReadAttributesResponse uint8 = 0x01
	FoundationWriteAttributes        uint8 = 0x02
	FoundationWriteAttributesResp    uint8 = 0x04
	FoundationConfigReporting        uint8 = 0x06
	FoundationConfigReportingResp    uint8 = 0x07
	FoundationReadReportingConfig    uint8 = 0x08
	FoundationReportAttributes       uint8 = 0x0A
	FoundationDefaultResponse        uint8 = 0x0B
	FoundationDiscoverAttributes     uint8 = 0x0C
	FoundationDiscoverAttributesResp uint8 = 0x0D
)

// ZCL status codes
const (
	ZCLStatusSuccess            uint8 = 0x00
	ZCLStatusFailure            uint8 = 0x01
	ZCLStatusUnsupportedAttr    uint8 = 0x86
	ZCLStatusInvalidDataType    uint8 = 0x8D
	ZCLStatusReadOnly           uint8 = 0x88
	ZCLStatusNotFound           uint8 = 0x8B
	ZCLStatusUnreportable       uint8 = 0x8C
	ZCLStatusInvalidValue       uint8 = 0x87
)
