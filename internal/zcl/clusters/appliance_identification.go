package clusters

import "zigbee-go-home/internal/zcl"

var ApplianceIdentification = zcl.ClusterDef{
	ID:   0x0B00,
	Name: "Appliance Identification",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "BasicIdentification", Type: zcl.TypeUint48, Access: zcl.AccessRead},
		{ID: 0x0010, Name: "CompanyName", Type: zcl.TypeCharStr, Access: zcl.AccessRead},
		{ID: 0x0011, Name: "CompanyId", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0012, Name: "BrandName", Type: zcl.TypeCharStr, Access: zcl.AccessRead},
		{ID: 0x0013, Name: "BrandId", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0014, Name: "Model", Type: zcl.TypeOctetStr, Access: zcl.AccessRead},
		{ID: 0x0015, Name: "PartNumber", Type: zcl.TypeOctetStr, Access: zcl.AccessRead},
		{ID: 0x0016, Name: "ProductRevision", Type: zcl.TypeOctetStr, Access: zcl.AccessRead},
		{ID: 0x0017, Name: "SoftwareRevision", Type: zcl.TypeOctetStr, Access: zcl.AccessRead},
		{ID: 0x0018, Name: "ProductTypeName", Type: zcl.TypeOctetStr, Access: zcl.AccessRead},
		{ID: 0x0019, Name: "ProductTypeId", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x001A, Name: "CECEDSpecificationVersion", Type: zcl.TypeUint8, Access: zcl.AccessRead},
	},
}

var MeterIdentification = zcl.ClusterDef{
	ID:   0x0B01,
	Name: "Meter Identification",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "CompanyName", Type: zcl.TypeCharStr, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "MeterTypeID", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x0004, Name: "DataQualityID", Type: zcl.TypeUint16, Access: zcl.AccessRead},
		{ID: 0x000C, Name: "POD", Type: zcl.TypeCharStr, Access: zcl.AccessRead},
		{ID: 0x000D, Name: "AvailablePower", Type: zcl.TypeInt24, Access: zcl.AccessRead},
		{ID: 0x000E, Name: "PowerThreshold", Type: zcl.TypeInt24, Access: zcl.AccessRead},
	},
}

var ApplianceEventsAndAlerts = zcl.ClusterDef{
	ID:   0x0B02,
	Name: "Appliance Events and Alerts",
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "GetAlerts", Direction: zcl.DirectionToServer},
		{ID: 0x00, Name: "GetAlertsResponse", Direction: zcl.DirectionToClient},
		{ID: 0x01, Name: "AlertsNotification", Direction: zcl.DirectionToClient},
		{ID: 0x02, Name: "EventNotification", Direction: zcl.DirectionToClient},
	},
}

var ApplianceStatistics = zcl.ClusterDef{
	ID:   0x0B03,
	Name: "Appliance Statistics",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "LogMaxSize", Type: zcl.TypeUint32, Access: zcl.AccessRead},
		{ID: 0x0001, Name: "LogQueueMaxSize", Type: zcl.TypeUint8, Access: zcl.AccessRead},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "LogNotification", Direction: zcl.DirectionToClient},
		{ID: 0x01, Name: "LogResponse", Direction: zcl.DirectionToClient},
		{ID: 0x02, Name: "LogQueueResponse", Direction: zcl.DirectionToClient},
		{ID: 0x03, Name: "StatisticsAvailable", Direction: zcl.DirectionToClient},
		{ID: 0x00, Name: "LogRequest", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "LogQueueRequest", Direction: zcl.DirectionToServer},
	},
}
