package clusters

import "zigbee-go-home/internal/zcl"

var ThermostatUserInterfaceConfiguration = zcl.ClusterDef{
	ID:   0x0204,
	Name: "Thermostat User Interface Configuration",
	Attributes: []zcl.AttributeDef{
		{ID: 0x0000, Name: "TemperatureDisplayMode", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0001, Name: "KeypadLockout", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
		{ID: 0x0002, Name: "ScheduleProgrammingVisibility", Type: zcl.TypeEnum8, Access: zcl.AccessRead | zcl.AccessWrite},
	},
}
