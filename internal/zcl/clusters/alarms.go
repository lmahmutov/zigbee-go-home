package clusters

import "zigbee-go-home/internal/zcl"

var Alarms = zcl.ClusterDef{
	ID:   0x0009,
	Name: "Alarms",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "AlarmCount", Type: zcl.TypeUint16, Access: zcl.AccessRead},
	},
	Commands: []zcl.CommandDef{
		{ID: 0x00, Name: "ResetAlarm", Direction: zcl.DirectionToServer},
		{ID: 0x01, Name: "ResetAllAlarms", Direction: zcl.DirectionToServer},
		{ID: 0x02, Name: "GetAlarm", Direction: zcl.DirectionToServer},
		{ID: 0x03, Name: "ResetAlarmLog", Direction: zcl.DirectionToServer},
		{ID: 0x00, Name: "Alarm", Direction: zcl.DirectionToClient},
		{ID: 0x01, Name: "GetAlarmResponse", Direction: zcl.DirectionToClient},
	},
}
