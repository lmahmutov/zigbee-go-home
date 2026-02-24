//go:build !no_mqtt

package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"zigbee-go-home/internal/coordinator"
	"zigbee-go-home/internal/store"
)

// Config holds MQTT bridge configuration.
type Config struct {
	Broker      string
	Username    string
	Password    string
	TopicPrefix string
}

// Bridge connects the Zigbee coordinator to MQTT with HA autodiscovery.
type Bridge struct {
	client pahomqtt.Client
	coord  *coordinator.Coordinator
	prefix string
	logger *slog.Logger
	unsub  func()
	ctx    context.Context
	cancel context.CancelFunc

	// Per-device state accumulator.
	mu     sync.Mutex
	states map[string]map[string]any // IEEE -> property map

	// Track pending delayed discovery goroutines per IEEE to avoid duplicates.
	pendingDiscovery map[string]context.CancelFunc
	discoveryGen     map[string]uint64
	nextDiscGen      uint64
}

// NewBridge creates and connects an MQTT bridge.
func NewBridge(coord *coordinator.Coordinator, cfg Config, logger *slog.Logger) (*Bridge, error) {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Bridge{
		coord:            coord,
		prefix:           cfg.TopicPrefix,
		logger:           logger.With("component", "mqtt"),
		states:           make(map[string]map[string]any),
		pendingDiscovery: make(map[string]context.CancelFunc),
		discoveryGen:     make(map[string]uint64),
		ctx:              ctx,
		cancel:           cancel,
	}

	opts := pahomqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID("zigbee-go-home").
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetWill(cfg.TopicPrefix+"/bridge/state", "offline", 1, true).
		SetOnConnectHandler(func(_ pahomqtt.Client) {
			b.logger.Info("MQTT connected")
			b.publishBridgeState("online")
			b.publishAllDiscovery()
			b.subscribeCommands()
		}).
		SetConnectionLostHandler(func(_ pahomqtt.Client, err error) {
			b.logger.Warn("MQTT connection lost", "err", err)
		})

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
		opts.SetPassword(cfg.Password)
	}

	client := pahomqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return nil, fmt.Errorf("mqtt connect timeout")
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("mqtt connect: %w", err)
	}

	b.client = client
	return b, nil
}

// Start subscribes to coordinator events and begins MQTT publishing.
func (b *Bridge) Start() {
	b.unsub = b.coord.Events().OnAll(b.handleEvent)
	b.logger.Info("MQTT bridge started", "prefix", b.prefix)
}

// Stop publishes offline state, unsubscribes, and disconnects.
func (b *Bridge) Stop() {
	b.cancel()
	if b.unsub != nil {
		b.unsub()
	}
	b.publishBridgeState("offline")
	b.client.Disconnect(1000)
	b.logger.Info("MQTT bridge stopped")
}

func (b *Bridge) handleEvent(event coordinator.Event) {
	switch event.Type {
	case coordinator.EventAttributeReport:
		b.handleAttributeReport(event)
	case coordinator.EventPropertyUpdate:
		b.handlePropertyUpdate(event)
	case coordinator.EventDeviceAnnounce:
		// Publish discovery after a delay to let interview complete.
		go b.delayedDiscovery(event)
	case coordinator.EventDeviceLeft:
		b.handleDeviceLeft(event)
	}
}

func (b *Bridge) handleAttributeReport(event coordinator.Event) {
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		return
	}
	ieee, _ := data["ieee"].(string)
	if ieee == "" {
		return
	}

	clusterID, _ := data["cluster_id"].(uint16)
	attrName, _ := data["attr_name"].(string)
	value := data["value"]

	// Map known cluster/attribute combos to state property names.
	propName := mapAttributeToProperty(clusterID, attrName)
	if propName == "" {
		return
	}

	// Convert bool OnOff to "ON"/"OFF" strings for Home Assistant.
	if propName == "state" {
		if b, ok := value.(bool); ok {
			if b {
				value = "ON"
			} else {
				value = "OFF"
			}
		}
	}

	b.updateAndPublishState(ieee, propName, value)
}

func (b *Bridge) handlePropertyUpdate(event coordinator.Event) {
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		return
	}
	ieee, _ := data["ieee"].(string)
	prop, _ := data["property"].(string)
	value := data["value"]
	if ieee == "" || prop == "" {
		return
	}

	b.updateAndPublishState(ieee, prop, value)
}

func (b *Bridge) updateAndPublishState(ieee, prop string, value any) {
	b.mu.Lock()
	state, ok := b.states[ieee]
	if !ok {
		state = make(map[string]any)
		b.states[ieee] = state
	}
	state[prop] = value

	// Always include LQI and last_seen from the device store.
	dev, err := b.coord.Devices().GetDevice(ieee)
	if err == nil {
		state["linkquality"] = dev.LQI
		state["last_seen"] = dev.LastSeen.Format(time.RFC3339)
	}

	payload := mustJSON(state)
	b.mu.Unlock()

	topic := b.prefix + "/" + b.topicName(ieee)
	b.publish(topic, payload, true)
}

func (b *Bridge) handleDeviceLeft(event coordinator.Event) {
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		return
	}
	ieee, _ := data["ieee"].(string)
	if ieee == "" {
		return
	}

	// Remove discovery entries.
	dev := &store.Device{IEEEAddress: ieee}
	for _, msg := range buildRemoveDiscovery(dev) {
		b.publish(msg.Topic, msg.Payload, true)
	}

	// Clear accumulated state.
	b.mu.Lock()
	delete(b.states, ieee)
	b.mu.Unlock()
}

func (b *Bridge) delayedDiscovery(event coordinator.Event) {
	data, ok := event.Data.(map[string]interface{})
	if !ok {
		return
	}
	ieee, _ := data["ieee"].(string)
	if ieee == "" {
		return
	}

	// Cancel any previous delayed discovery for this device.
	discCtx, discCancel := context.WithCancel(b.ctx)
	b.mu.Lock()
	if prev, ok := b.pendingDiscovery[ieee]; ok {
		prev()
	}
	b.nextDiscGen++
	gen := b.nextDiscGen
	b.pendingDiscovery[ieee] = discCancel
	b.discoveryGen[ieee] = gen
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		if b.discoveryGen[ieee] == gen {
			delete(b.pendingDiscovery, ieee)
			delete(b.discoveryGen, ieee)
		}
		b.mu.Unlock()
		discCancel()
	}()

	// Wait for interview to complete (up to 3 minutes, checking periodically).
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for i := 0; i < 36; i++ {
		select {
		case <-ticker.C:
		case <-discCtx.Done():
			return
		}
		dev, err := b.coord.Devices().GetDevice(ieee)
		if err != nil {
			return
		}
		if dev.Interviewed {
			b.publishDeviceDiscovery(dev)
			return
		}
	}
}

func (b *Bridge) publishBridgeState(state string) {
	topic := b.prefix + "/bridge/state"
	b.publish(topic, []byte(state), true)
}

func (b *Bridge) publishAllDiscovery() {
	devices, err := b.coord.Devices().ListDevices()
	if err != nil {
		b.logger.Error("list devices for discovery", "err", err)
		return
	}
	for _, dev := range devices {
		if dev.Interviewed {
			b.publishDeviceDiscovery(dev)
		}
	}
}

func (b *Bridge) publishDeviceDiscovery(dev *store.Device) {
	for _, msg := range buildDiscovery(dev, b.prefix) {
		b.publish(msg.Topic, msg.Payload, true)
	}
	b.logger.Info("published HA discovery", "ieee", dev.IEEEAddress, "name", deviceDisplayName(dev))
}

func (b *Bridge) subscribeCommands() {
	devices, err := b.coord.Devices().ListDevices()
	if err != nil {
		b.logger.Error("list devices for command subscription", "err", err)
		return
	}
	for _, dev := range devices {
		if !dev.Interviewed {
			continue
		}
		b.subscribeDeviceCommands(dev)
	}
}

func (b *Bridge) subscribeDeviceCommands(dev *store.Device) {
	topic := b.prefix + "/" + deviceTopicName(dev) + "/set"
	ieee := dev.IEEEAddress
	b.client.Subscribe(topic, 1, func(_ pahomqtt.Client, msg pahomqtt.Message) {
		b.handleCommand(ieee, msg.Payload())
	})
}

func (b *Bridge) handleCommand(ieee string, payload []byte) {
	dev, err := b.coord.Devices().GetDevice(ieee)
	if err != nil {
		b.logger.Warn("command for unknown device", "ieee", ieee)
		return
	}
	if len(dev.Endpoints) == 0 {
		return
	}

	var cmd map[string]interface{}
	if err := json.Unmarshal(payload, &cmd); err != nil {
		b.logger.Warn("invalid command JSON", "ieee", ieee, "err", err)
		return
	}

	ep := dev.Endpoints[0].ID
	ctx, cancel := context.WithTimeout(b.coord.Context(), 10*time.Second)
	defer cancel()

	// Handle state command (ON/OFF).
	if state, ok := cmd["state"].(string); ok {
		switch strings.ToUpper(state) {
		case "ON":
			if err := b.coord.SendClusterCommand(ctx, dev.ShortAddress, ep, 0x0006, 0x01, nil); err != nil {
				b.logger.Warn("on command failed", "ieee", ieee, "err", err)
			} else {
				b.updateAndPublishState(ieee, "state", "ON")
			}
		case "OFF":
			if err := b.coord.SendClusterCommand(ctx, dev.ShortAddress, ep, 0x0006, 0x00, nil); err != nil {
				b.logger.Warn("off command failed", "ieee", ieee, "err", err)
			} else {
				b.updateAndPublishState(ieee, "state", "OFF")
			}
		case "TOGGLE":
			if err := b.coord.SendClusterCommand(ctx, dev.ShortAddress, ep, 0x0006, 0x02, nil); err != nil {
				b.logger.Warn("toggle command failed", "ieee", ieee, "err", err)
			}
		}
	}

	// Handle brightness command.
	if brightness, ok := toFloat64(cmd["brightness"]); ok {
		level := uint8(brightness)
		if level > 254 {
			level = 254
		}
		// Move to Level with On/Off, transition time 5 (0.5s).
		payload := []byte{level, 0x05, 0x00}
		if err := b.coord.SendClusterCommand(ctx, dev.ShortAddress, ep, 0x0008, 0x04, payload); err != nil {
			b.logger.Warn("brightness command failed", "ieee", ieee, "err", err)
		} else {
			b.updateAndPublishState(ieee, "brightness", level)
		}
	}
}

func (b *Bridge) publish(topic string, payload []byte, retained bool) {
	token := b.client.Publish(topic, 1, retained, payload)
	go func() {
		if !token.WaitTimeout(5 * time.Second) {
			b.logger.Warn("MQTT publish timeout", "topic", topic)
		} else if err := token.Error(); err != nil {
			b.logger.Warn("MQTT publish error", "topic", topic, "err", err)
		}
	}()
}

// topicName returns the MQTT topic name for a device by IEEE.
func (b *Bridge) topicName(ieee string) string {
	dev, err := b.coord.Devices().GetDevice(ieee)
	if err != nil {
		return ieee
	}
	return deviceTopicName(dev)
}

// mapAttributeToProperty maps well-known cluster/attribute combos to property names.
func mapAttributeToProperty(clusterID uint16, attrName string) string {
	switch clusterID {
	case 0x0006: // On/Off
		if attrName == "OnOff" || attrName == "on_off" {
			return "state"
		}
	case 0x0008: // Level Control
		if attrName == "CurrentLevel" || attrName == "current_level" {
			return "brightness"
		}
	case 0x0402: // Temperature
		if attrName == "MeasuredValue" || attrName == "measured_value" {
			return "temperature"
		}
	case 0x0405: // Humidity
		if attrName == "MeasuredValue" || attrName == "measured_value" {
			return "humidity"
		}
	case 0x0403: // Pressure
		if attrName == "MeasuredValue" || attrName == "measured_value" {
			return "pressure"
		}
	case 0x0400: // Illuminance
		if attrName == "MeasuredValue" || attrName == "measured_value" {
			return "illuminance"
		}
	case 0x0406: // Occupancy
		if attrName == "Occupancy" || attrName == "occupancy" {
			return "occupancy"
		}
	case 0x0001: // Power Configuration
		if attrName == "BatteryPercentageRemaining" || attrName == "battery_percentage_remaining" {
			return "battery"
		}
	}
	return ""
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	default:
		return 0, false
	}
}

func mustJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return data
}
