// Package ncp defines the interface for the Zigbee Network Co-Processor backend.
// Backend: nRF52840 (ZBOSS NCP over USB CDC ACM).
package ncp

import "context"

// NCP is the abstract interface for a Zigbee NCP device.
type NCP interface {
	// Network management
	Reset(ctx context.Context) error
	FactoryReset(ctx context.Context) error
	Init(ctx context.Context) error
	FormNetwork(ctx context.Context, cfg NetworkConfig) error
	StartNetwork(ctx context.Context) error
	PermitJoin(ctx context.Context, duration uint8) error
	NetworkInfo(ctx context.Context) (*NetworkInfo, error)
	NetworkScan(ctx context.Context) ([]NetworkScanResult, error)
	GetLocalIEEE(ctx context.Context) ([8]byte, error)

	// ZDO
	ActiveEndpoints(ctx context.Context, shortAddr uint16) ([]uint8, error)
	SimpleDescriptor(ctx context.Context, shortAddr uint16, endpoint uint8) (*SimpleDescriptor, error)
	Bind(ctx context.Context, req BindRequest) error
	Unbind(ctx context.Context, req BindRequest) error
	MgmtLeave(ctx context.Context, shortAddr uint16, ieeeAddr [8]byte) error

	// ZCL
	ReadAttributes(ctx context.Context, req ReadAttributesRequest) ([]AttributeResponse, error)
	WriteAttributes(ctx context.Context, req WriteAttributesRequest) error
	SendCommand(ctx context.Context, req ClusterCommandRequest) error
	ConfigureReporting(ctx context.Context, req ConfigureReportingRequest) error

	// Indication callbacks
	OnDeviceJoined(handler func(DeviceJoinedEvent))
	OnDeviceLeft(handler func(DeviceLeftEvent))
	OnDeviceAnnounce(handler func(DeviceAnnounceEvent))
	OnAttributeReport(handler func(AttributeReportEvent))
	OnClusterCommand(handler func(ClusterCommandEvent))
	OnNwkAddrUpdate(handler func(uint16))

	// Info
	GetNCPInfo() *NCPInfo

	// Lifecycle
	Close() error
}

// NCPInfo holds firmware/stack version information from the NCP.
type NCPInfo struct {
	FWVersion       uint32
	StackVersion    string // e.g. "3.11.3.0"
	ProtocolVersion uint32
	NetworkKey      []byte // 16-byte network key, set during FormNetwork
}

// NetworkConfig holds parameters for network formation.
type NetworkConfig struct {
	Channel  uint8
	PanID    uint16
	ExtPanID [8]byte
}

// NetworkInfo holds current network state.
type NetworkInfo struct {
	Channel  uint8
	PanID    uint16
	ExtPanID [8]byte
	State    uint8
}

// NetworkScanResult holds one discovered network from an active scan.
type NetworkScanResult struct {
	ExtPanID    [8]byte `json:"ext_pan_id"`
	PanID       uint16  `json:"pan_id"`
	UpdateID    uint8   `json:"update_id"`
	Channel     uint8   `json:"channel"`
	StackProfile uint8  `json:"stack_profile"`
	PermitJoin  bool    `json:"permit_join"`
	RouterCap   bool    `json:"router_capacity"`
	EDCap       bool    `json:"end_device_capacity"`
	LQI         uint8   `json:"lqi"`
	RSSI        int8    `json:"rssi"`
}

// SimpleDescriptor describes an endpoint.
type SimpleDescriptor struct {
	Endpoint    uint8
	ProfileID   uint16
	DeviceID    uint16
	InClusters  []uint16
	OutClusters []uint16
}

// BindRequest is a ZDO bind/unbind request.
type BindRequest struct {
	TargetShortAddr uint16
	SrcIEEE         [8]byte
	SrcEP           uint8
	ClusterID       uint16
	DstIEEE         [8]byte
	DstEP           uint8
}

// ReadAttributesRequest specifies which attributes to read.
type ReadAttributesRequest struct {
	DstAddr   uint16
	DstEP     uint8
	ClusterID uint16
	AttrIDs   []uint16
}

// AttributeResponse holds a single attribute read result.
type AttributeResponse struct {
	AttrID   uint16
	Status   uint8
	DataType uint8
	Value    []byte
}

// WriteAttributesRequest specifies attributes to write.
type WriteAttributesRequest struct {
	DstAddr   uint16
	DstEP     uint8
	ClusterID uint16
	Records   []WriteRecord
}

// WriteRecord is a single attribute write.
type WriteRecord struct {
	AttrID   uint16
	DataType uint8
	Value    []byte
}

// ClusterCommandRequest sends a cluster-specific command.
type ClusterCommandRequest struct {
	DstAddr   uint16
	DstEP     uint8
	ClusterID uint16
	CommandID uint8
	Payload   []byte
}

// ConfigureReportingRequest sets up attribute reporting.
type ConfigureReportingRequest struct {
	DstAddr      uint16
	DstEP        uint8
	ClusterID    uint16
	AttrID       uint16
	DataType     uint8
	MinInterval  uint16
	MaxInterval  uint16
	ReportChange []byte
}

// DeviceJoinedEvent is emitted when a device joins the network.
type DeviceJoinedEvent struct {
	ShortAddr uint16
	IEEEAddr  [8]byte
}

// DeviceLeftEvent is emitted when a device leaves.
type DeviceLeftEvent struct {
	ShortAddr uint16
	IEEEAddr  [8]byte
}

// DeviceAnnounceEvent is emitted on device announce.
type DeviceAnnounceEvent struct {
	ShortAddr  uint16
	IEEEAddr   [8]byte
	Capability uint8
}

// AttributeReportEvent is emitted for unsolicited attribute reports.
type AttributeReportEvent struct {
	SrcAddr   uint16
	SrcEP     uint8
	ClusterID uint16
	AttrID    uint16
	DataType  uint8
	Value     []byte
	LQI       uint8
	RSSI      int8
}

// ClusterCommandEvent is emitted for incoming cluster-specific commands (e.g., Tuya DP).
type ClusterCommandEvent struct {
	SrcAddr   uint16
	SrcEP     uint8
	ClusterID uint16
	CommandID uint8
	Payload   []byte
	LQI       uint8
	RSSI      int8
}
