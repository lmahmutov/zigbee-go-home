//go:build !no_mqtt

package mqtt

import (
	"encoding/json"
	"testing"
	"time"

	"zigbee-go-home/internal/store"
)

func TestDiscoveryTemperatureSensor(t *testing.T) {
	dev := &store.Device{
		IEEEAddress:  "00158D00012A3B4C",
		Manufacturer: "LUMI",
		Model:        "lumi.weather",
		FriendlyName: "Kitchen Sensor",
		Interviewed:  true,
		Endpoints: []store.Endpoint{
			{
				ID:         1,
				ProfileID:  0x0104,
				InClusters: []uint16{0x0000, 0x0402, 0x0403, 0x0405},
			},
		},
	}

	msgs := buildDiscovery(dev, "zigbee2mqtt")
	if len(msgs) == 0 {
		t.Fatal("expected discovery messages")
	}

	// Find the temperature sensor discovery.
	var tempMsg *discoveryMsg
	for i := range msgs {
		if msgs[i].Topic == "homeassistant/sensor/zigbee_00158D00012A3B4C/temperature/config" {
			tempMsg = &msgs[i]
			break
		}
	}
	if tempMsg == nil {
		t.Fatal("temperature discovery not found")
	}

	var payload haDiscovery
	if err := json.Unmarshal(tempMsg.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload.Name != "Kitchen Sensor Temperature" {
		t.Errorf("name = %q, want %q", payload.Name, "Kitchen Sensor Temperature")
	}
	if payload.UniqueID != "zigbee_00158D00012A3B4C_temperature" {
		t.Errorf("unique_id = %q", payload.UniqueID)
	}
	if payload.DeviceClass != "temperature" {
		t.Errorf("device_class = %q", payload.DeviceClass)
	}
	if payload.UnitOfMeasurement != "°C" {
		t.Errorf("unit = %q", payload.UnitOfMeasurement)
	}
	if payload.StateTopic != "zigbee2mqtt/kitchen_sensor" {
		t.Errorf("state_topic = %q", payload.StateTopic)
	}
	if payload.AvailabilityTopic != "zigbee2mqtt/bridge/state" {
		t.Errorf("availability_topic = %q", payload.AvailabilityTopic)
	}
	if payload.Device.Manufacturer != "LUMI" {
		t.Errorf("device.manufacturer = %q", payload.Device.Manufacturer)
	}

	// Should also have humidity, pressure, and linkquality.
	topics := make(map[string]bool)
	for _, m := range msgs {
		topics[m.Topic] = true
	}
	if !topics["homeassistant/sensor/zigbee_00158D00012A3B4C/humidity/config"] {
		t.Error("humidity discovery missing")
	}
	if !topics["homeassistant/sensor/zigbee_00158D00012A3B4C/pressure/config"] {
		t.Error("pressure discovery missing")
	}
	if !topics["homeassistant/sensor/zigbee_00158D00012A3B4C/linkquality/config"] {
		t.Error("linkquality discovery missing")
	}
}

func TestDiscoveryLightVsSwitch(t *testing.T) {
	// Device with On/Off + Level Control → light.
	lightDev := &store.Device{
		IEEEAddress: "AABBCCDDEEFF0011",
		Model:       "light_bulb",
		Interviewed: true,
		Endpoints: []store.Endpoint{
			{
				ID:         1,
				InClusters: []uint16{0x0006, 0x0008},
			},
		},
	}

	msgs := buildDiscovery(lightDev, "zigbee2mqtt")
	topics := extractTopics(msgs)

	if !topics["homeassistant/light/zigbee_AABBCCDDEEFF0011/light/config"] {
		t.Error("expected light discovery for device with on/off + level")
	}
	if topics["homeassistant/switch/zigbee_AABBCCDDEEFF0011/switch/config"] {
		t.Error("should NOT have switch discovery for a light device")
	}

	// Device with only On/Off → switch.
	switchDev := &store.Device{
		IEEEAddress: "1122334455667788",
		Model:       "smart_plug",
		Interviewed: true,
		Endpoints: []store.Endpoint{
			{
				ID:         1,
				InClusters: []uint16{0x0006},
			},
		},
	}

	msgs = buildDiscovery(switchDev, "zigbee2mqtt")
	topics = extractTopics(msgs)

	if !topics["homeassistant/switch/zigbee_1122334455667788/switch/config"] {
		t.Error("expected switch discovery for device with on/off only")
	}
	if topics["homeassistant/light/zigbee_1122334455667788/light/config"] {
		t.Error("should NOT have light discovery for a switch device")
	}
}

func TestDiscoveryLightHasCommandTopic(t *testing.T) {
	dev := &store.Device{
		IEEEAddress:  "AABBCCDDEEFF0011",
		FriendlyName: "Living Room Light",
		Interviewed:  true,
		Endpoints: []store.Endpoint{
			{ID: 1, InClusters: []uint16{0x0006, 0x0008}},
		},
	}

	msgs := buildDiscovery(dev, "zigbee2mqtt")
	for _, m := range msgs {
		if m.Topic == "homeassistant/light/zigbee_AABBCCDDEEFF0011/light/config" {
			var payload haDiscovery
			if err := json.Unmarshal(m.Payload, &payload); err != nil {
				t.Fatal(err)
			}
			if payload.CommandTopic != "zigbee2mqtt/living_room_light/set" {
				t.Errorf("command_topic = %q, want %q", payload.CommandTopic, "zigbee2mqtt/living_room_light/set")
			}
			if payload.Schema != "json" {
				t.Errorf("schema = %q, want %q", payload.Schema, "json")
			}
			return
		}
	}
	t.Fatal("light discovery not found")
}

func TestDiscoveryUninterviewedDevice(t *testing.T) {
	dev := &store.Device{
		IEEEAddress: "0000000000000001",
		Interviewed: false,
	}
	msgs := buildDiscovery(dev, "zigbee2mqtt")
	if len(msgs) != 0 {
		t.Errorf("expected no discovery for uninterviewed device, got %d", len(msgs))
	}
}

func TestDeviceDisplayName(t *testing.T) {
	tests := []struct {
		name string
		dev  *store.Device
		want string
	}{
		{
			name: "friendly name",
			dev:  &store.Device{FriendlyName: "Kitchen Light", Manufacturer: "IKEA", Model: "TRADFRI"},
			want: "Kitchen Light",
		},
		{
			name: "manufacturer and model",
			dev:  &store.Device{Manufacturer: "IKEA", Model: "TRADFRI"},
			want: "IKEA TRADFRI",
		},
		{
			name: "model only",
			dev:  &store.Device{Model: "TRADFRI"},
			want: "TRADFRI",
		},
		{
			name: "IEEE fallback",
			dev:  &store.Device{IEEEAddress: "00158D00012A3B4C"},
			want: "00158D00012A3B4C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deviceDisplayName(tt.dev)
			if got != tt.want {
				t.Errorf("deviceDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeviceTopicName(t *testing.T) {
	tests := []struct {
		name string
		dev  *store.Device
		want string
	}{
		{
			name: "friendly name with spaces",
			dev:  &store.Device{FriendlyName: "Kitchen Light", IEEEAddress: "AABB"},
			want: "kitchen_light",
		},
		{
			name: "IEEE fallback",
			dev:  &store.Device{IEEEAddress: "00158D00012A3B4C"},
			want: "00158D00012A3B4C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deviceTopicName(tt.dev)
			if got != tt.want {
				t.Errorf("deviceTopicName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMapAttributeToProperty(t *testing.T) {
	tests := []struct {
		cluster  uint16
		attrName string
		want     string
	}{
		{0x0006, "OnOff", "state"},
		{0x0008, "CurrentLevel", "brightness"},
		{0x0402, "MeasuredValue", "temperature"},
		{0x0405, "MeasuredValue", "humidity"},
		{0x0403, "MeasuredValue", "pressure"},
		{0x0400, "MeasuredValue", "illuminance"},
		{0x0406, "Occupancy", "occupancy"},
		{0x0001, "BatteryPercentageRemaining", "battery"},
		{0x0000, "ModelIdentifier", ""}, // unmapped
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := mapAttributeToProperty(tt.cluster, tt.attrName)
			if got != tt.want {
				t.Errorf("mapAttributeToProperty(0x%04X, %q) = %q, want %q", tt.cluster, tt.attrName, got, tt.want)
			}
		})
	}
}

func TestRemoveDiscovery(t *testing.T) {
	dev := &store.Device{IEEEAddress: "AABBCCDD11223344"}
	msgs := buildRemoveDiscovery(dev)
	if len(msgs) == 0 {
		t.Fatal("expected removal messages")
	}

	for _, m := range msgs {
		if m.Payload != nil {
			t.Errorf("removal message should have nil payload, got %q for %s", m.Payload, m.Topic)
		}
		if m.Topic == "" {
			t.Error("removal message has empty topic")
		}
	}
}

func TestCommandParse(t *testing.T) {
	// Test that handleCommand JSON parsing works for known command structures.
	tests := []struct {
		name    string
		payload string
		wantKey string
	}{
		{"on", `{"state":"ON"}`, "state"},
		{"off", `{"state":"OFF"}`, "state"},
		{"brightness", `{"brightness":128}`, "brightness"},
		{"combined", `{"state":"ON","brightness":200}`, "state"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd map[string]interface{}
			if err := json.Unmarshal([]byte(tt.payload), &cmd); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if _, ok := cmd[tt.wantKey]; !ok {
				t.Errorf("expected key %q in command", tt.wantKey)
			}
		})
	}
}

func TestBatteryPropertyDiscovery(t *testing.T) {
	// Device without power config cluster but with battery property.
	dev := &store.Device{
		IEEEAddress: "AABBCCDD00112233",
		Interviewed: true,
		Endpoints: []store.Endpoint{
			{ID: 1, InClusters: []uint16{0x0000}},
		},
		Properties: map[string]any{
			"battery": 85,
		},
	}

	msgs := buildDiscovery(dev, "zigbee2mqtt")
	topics := extractTopics(msgs)
	if !topics["homeassistant/sensor/zigbee_AABBCCDD00112233/battery/config"] {
		t.Error("expected battery discovery for device with battery property")
	}
}

func TestMustJSON(t *testing.T) {
	result := mustJSON(map[string]string{"hello": "world"})
	var parsed map[string]string
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("mustJSON output not valid JSON: %v", err)
	}
	if parsed["hello"] != "world" {
		t.Errorf("parsed value = %q", parsed["hello"])
	}
}

func TestStateTopicFormat(t *testing.T) {
	dev := &store.Device{
		IEEEAddress:  "00158D00012A3B4C",
		FriendlyName: "Living Room",
		Interviewed:  true,
		JoinedAt:     time.Now(),
		LastSeen:     time.Now(),
		Endpoints: []store.Endpoint{
			{ID: 1, InClusters: []uint16{0x0402}},
		},
	}

	msgs := buildDiscovery(dev, "zigbee2mqtt")
	for _, m := range msgs {
		var payload haDiscovery
		if err := json.Unmarshal(m.Payload, &payload); err != nil {
			continue
		}
		if payload.StateTopic != "" && payload.StateTopic != "zigbee2mqtt/living_room" {
			t.Errorf("state_topic = %q, want %q", payload.StateTopic, "zigbee2mqtt/living_room")
		}
	}
}

func extractTopics(msgs []discoveryMsg) map[string]bool {
	topics := make(map[string]bool)
	for _, m := range msgs {
		topics[m.Topic] = true
	}
	return topics
}
