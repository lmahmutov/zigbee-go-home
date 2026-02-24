//go:build !no_mqtt

package mqtt

import (
	"fmt"
	"strings"

	"zigbee-go-home/internal/store"
)

// discoveryMsg is a Home Assistant MQTT discovery payload.
type discoveryMsg struct {
	Topic   string // e.g. "homeassistant/sensor/zigbee_00158D.../temperature/config"
	Payload []byte // JSON, empty means delete
}

// haDevice is the "device" block in HA discovery.
type haDevice struct {
	Identifiers  []string `json:"identifiers"`
	Manufacturer string   `json:"manufacturer,omitempty"`
	Model        string   `json:"model,omitempty"`
	Name         string   `json:"name"`
}

// haDiscovery is a generic HA discovery payload.
type haDiscovery struct {
	Name                string   `json:"name"`
	UniqueID            string   `json:"unique_id"`
	StateTopic          string   `json:"state_topic"`
	CommandTopic        string   `json:"command_topic,omitempty"`
	AvailabilityTopic   string   `json:"availability_topic"`
	ValueTemplate       string   `json:"value_template,omitempty"`
	UnitOfMeasurement   string   `json:"unit_of_measurement,omitempty"`
	DeviceClass         string   `json:"device_class,omitempty"`
	StateClass          string   `json:"state_class,omitempty"`
	PayloadOn           string   `json:"payload_on,omitempty"`
	PayloadOff          string   `json:"payload_off,omitempty"`
	BrightnessScale     int      `json:"brightness_scale,omitempty"`
	BrightnessStateTopic   string `json:"brightness_state_topic,omitempty"`
	BrightnessCommandTopic string `json:"brightness_command_topic,omitempty"`
	SupportedColorModes []string `json:"supported_color_modes,omitempty"`
	Schema              string   `json:"schema,omitempty"`
	Device              haDevice `json:"device"`
}

// deviceDisplayName returns a display name for the device.
func deviceDisplayName(dev *store.Device) string {
	if dev.FriendlyName != "" {
		return dev.FriendlyName
	}
	if dev.Manufacturer != "" && dev.Model != "" {
		return dev.Manufacturer + " " + dev.Model
	}
	if dev.Model != "" {
		return dev.Model
	}
	return dev.IEEEAddress
}

// deviceIdentifier returns the unique identifier for HA device registry.
func deviceIdentifier(dev *store.Device) string {
	return "zigbee_" + dev.IEEEAddress
}

// deviceTopicName returns the topic name for a device (friendly name or IEEE).
func deviceTopicName(dev *store.Device) string {
	if dev.FriendlyName != "" {
		// Sanitize: lowercase and keep only safe chars for MQTT topics.
		name := strings.ToLower(dev.FriendlyName)
		name = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
				return r
			}
			return '_'
		}, name)
		return name
	}
	return dev.IEEEAddress
}

// buildDiscovery generates HA discovery messages for a device based on its clusters.
func buildDiscovery(dev *store.Device, prefix string) []discoveryMsg {
	if !dev.Interviewed || len(dev.Endpoints) == 0 {
		return nil
	}

	avail := prefix + "/bridge/state"
	stateTopic := prefix + "/" + deviceTopicName(dev)
	nodeID := deviceIdentifier(dev)
	displayName := deviceDisplayName(dev)

	haDev := haDevice{
		Identifiers:  []string{nodeID},
		Manufacturer: dev.Manufacturer,
		Model:        dev.Model,
		Name:         displayName,
	}

	// Collect cluster IDs across all endpoints.
	hasCluster := make(map[uint16]bool)
	for _, ep := range dev.Endpoints {
		for _, cid := range ep.InClusters {
			hasCluster[cid] = true
		}
	}

	var msgs []discoveryMsg

	// Light vs Switch: if device has Level Control (0x0008), it's a light.
	// If only On/Off (0x0006), it's a switch.
	hasOnOff := hasCluster[0x0006]
	hasLevel := hasCluster[0x0008]

	if hasOnOff && hasLevel {
		msgs = append(msgs, buildLight(nodeID, displayName, stateTopic, avail, haDev, prefix, dev))
	} else if hasOnOff {
		msgs = append(msgs, buildSwitch(nodeID, displayName, stateTopic, avail, haDev, prefix, dev))
	}

	// Temperature (0x0402)
	if hasCluster[0x0402] {
		msgs = append(msgs, buildSensor(nodeID, displayName, stateTopic, avail, haDev,
			"temperature", "Temperature", "temperature", "°C", "measurement",
			"{{ value_json.temperature }}"))
	}

	// Humidity (0x0405)
	if hasCluster[0x0405] {
		msgs = append(msgs, buildSensor(nodeID, displayName, stateTopic, avail, haDev,
			"humidity", "Humidity", "humidity", "%", "measurement",
			"{{ value_json.humidity }}"))
	}

	// Pressure (0x0403)
	if hasCluster[0x0403] {
		msgs = append(msgs, buildSensor(nodeID, displayName, stateTopic, avail, haDev,
			"pressure", "Pressure", "pressure", "hPa", "measurement",
			"{{ value_json.pressure }}"))
	}

	// Illuminance (0x0400)
	if hasCluster[0x0400] {
		msgs = append(msgs, buildSensor(nodeID, displayName, stateTopic, avail, haDev,
			"illuminance", "Illuminance", "illuminance", "lx", "measurement",
			"{{ value_json.illuminance }}"))
	}

	// Occupancy (0x0406)
	if hasCluster[0x0406] {
		msgs = append(msgs, buildBinarySensor(nodeID, displayName, stateTopic, avail, haDev,
			"occupancy", "Occupancy", "occupancy",
			"{{ 'ON' if value_json.occupancy else 'OFF' }}"))
	}

	// IAS Zone (0x0500)
	if hasCluster[0x0500] {
		msgs = append(msgs, buildBinarySensor(nodeID, displayName, stateTopic, avail, haDev,
			"zone", "Zone", "safety",
			"{{ 'ON' if value_json.zone_status else 'OFF' }}"))
	}

	// Power Configuration (0x0001) → battery sensor
	if hasCluster[0x0001] {
		msgs = append(msgs, buildSensor(nodeID, displayName, stateTopic, avail, haDev,
			"battery", "Battery", "battery", "%", "measurement",
			"{{ value_json.battery }}"))
	}

	// Also check properties for battery (some devices use property-based battery via TLV)
	if dev.Properties != nil {
		if _, ok := dev.Properties["battery"]; ok && !hasCluster[0x0001] {
			msgs = append(msgs, buildSensor(nodeID, displayName, stateTopic, avail, haDev,
				"battery", "Battery", "battery", "%", "measurement",
				"{{ value_json.battery }}"))
		}
	}

	// Analog Input (0x000C)
	if hasCluster[0x000C] {
		msgs = append(msgs, buildSensor(nodeID, displayName, stateTopic, avail, haDev,
			"analog", "Analog Input", "", "", "measurement",
			"{{ value_json.analog }}"))
	}

	// Link quality sensor for all devices.
	// No device_class — "signal_strength" requires dB/dBm units, but LQI is unitless.
	msgs = append(msgs, buildSensor(nodeID, displayName, stateTopic, avail, haDev,
		"linkquality", "Link Quality", "", "lqi", "measurement",
		"{{ value_json.linkquality }}"))

	return msgs
}

func buildSensor(nodeID, displayName, stateTopic, avail string, haDev haDevice,
	objectID, suffix, deviceClass, unit, stateClass, valueTmpl string) discoveryMsg {

	topic := fmt.Sprintf("homeassistant/sensor/%s/%s/config", nodeID, objectID)
	payload := haDiscovery{
		Name:              displayName + " " + suffix,
		UniqueID:          nodeID + "_" + objectID,
		StateTopic:        stateTopic,
		AvailabilityTopic: avail,
		ValueTemplate:     valueTmpl,
		UnitOfMeasurement: unit,
		DeviceClass:       deviceClass,
		StateClass:        stateClass,
		Device:            haDev,
	}
	return discoveryMsg{Topic: topic, Payload: mustJSON(payload)}
}

func buildBinarySensor(nodeID, displayName, stateTopic, avail string, haDev haDevice,
	objectID, suffix, deviceClass, valueTmpl string) discoveryMsg {

	topic := fmt.Sprintf("homeassistant/binary_sensor/%s/%s/config", nodeID, objectID)
	payload := haDiscovery{
		Name:              displayName + " " + suffix,
		UniqueID:          nodeID + "_" + objectID,
		StateTopic:        stateTopic,
		AvailabilityTopic: avail,
		ValueTemplate:     valueTmpl,
		DeviceClass:       deviceClass,
		PayloadOn:         "ON",
		PayloadOff:        "OFF",
		Device:            haDev,
	}
	return discoveryMsg{Topic: topic, Payload: mustJSON(payload)}
}

func buildLight(nodeID, displayName, stateTopic, avail string, haDev haDevice, prefix string, dev *store.Device) discoveryMsg {
	topic := fmt.Sprintf("homeassistant/light/%s/light/config", nodeID)
	cmdTopic := prefix + "/" + deviceTopicName(dev) + "/set"
	payload := haDiscovery{
		Name:                displayName,
		UniqueID:            nodeID + "_light",
		StateTopic:          stateTopic,
		CommandTopic:        cmdTopic,
		AvailabilityTopic:   avail,
		SupportedColorModes: []string{"brightness"},
		BrightnessScale:     254,
		Schema:              "json",
		Device:              haDev,
	}
	return discoveryMsg{Topic: topic, Payload: mustJSON(payload)}
}

func buildSwitch(nodeID, displayName, stateTopic, avail string, haDev haDevice, prefix string, dev *store.Device) discoveryMsg {
	topic := fmt.Sprintf("homeassistant/switch/%s/switch/config", nodeID)
	cmdTopic := prefix + "/" + deviceTopicName(dev) + "/set"
	payload := haDiscovery{
		Name:              displayName,
		UniqueID:          nodeID + "_switch",
		StateTopic:        stateTopic,
		CommandTopic:      cmdTopic,
		AvailabilityTopic: avail,
		ValueTemplate:     "{{ value_json.state }}",
		PayloadOn:         "ON",
		PayloadOff:        "OFF",
		Device:            haDev,
	}
	return discoveryMsg{Topic: topic, Payload: mustJSON(payload)}
}

// buildRemoveDiscovery generates empty retained messages to remove a device from HA.
func buildRemoveDiscovery(dev *store.Device) []discoveryMsg {
	nodeID := deviceIdentifier(dev)

	// Remove all possible component types.
	components := []struct{ comp, obj string }{
		{"light", "light"},
		{"switch", "switch"},
		{"sensor", "temperature"},
		{"sensor", "humidity"},
		{"sensor", "pressure"},
		{"sensor", "illuminance"},
		{"sensor", "battery"},
		{"sensor", "analog"},
		{"sensor", "linkquality"},
		{"binary_sensor", "occupancy"},
		{"binary_sensor", "zone"},
	}

	var msgs []discoveryMsg
	for _, c := range components {
		msgs = append(msgs, discoveryMsg{
			Topic:   fmt.Sprintf("homeassistant/%s/%s/%s/config", c.comp, nodeID, c.obj),
			Payload: nil, // empty retained = delete
		})
	}
	return msgs
}
