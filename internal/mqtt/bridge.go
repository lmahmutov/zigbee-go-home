//go:build !no_mqtt

package mqtt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

	// Async event processing channel.
	eventCh chan coordinator.Event
	eventWg sync.WaitGroup

	// Per-device state accumulator.
	mu     sync.Mutex
	states map[string]map[string]any // IEEE -> property map

	// Cached topic names to avoid DB reads on every publish.
	topicNames map[string]string // IEEE -> topic name

	// Track pending delayed discovery goroutines per IEEE to avoid duplicates.
	pendingDiscovery map[string]context.CancelFunc
	discoveryGen     map[string]uint64
	nextDiscGen      uint64

	// WaitGroup for delayedDiscovery goroutines so Stop() can wait.
	discWg sync.WaitGroup
}

// NewBridge creates and connects an MQTT bridge.
func NewBridge(coord *coordinator.Coordinator, cfg Config, logger *slog.Logger) (*Bridge, error) {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Bridge{
		coord:            coord,
		prefix:           cfg.TopicPrefix,
		logger:           logger.With("component", "mqtt"),
		eventCh:          make(chan coordinator.Event, 256),
		states:           make(map[string]map[string]any),
		topicNames:       make(map[string]string),
		pendingDiscovery: make(map[string]context.CancelFunc),
		discoveryGen:     make(map[string]uint64),
		ctx:              ctx,
		cancel:           cancel,
	}

	opts := pahomqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID("zigbee-go-home-" + shortRandomID()).
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
	b.eventWg.Add(1)
	go b.eventLoop()

	b.unsub = b.coord.Events().OnAll(func(event coordinator.Event) {
		select {
		case b.eventCh <- event:
		default:
			b.logger.Warn("MQTT event channel full, dropping event", "type", event.Type)
		}
	})
	b.logger.Info("MQTT bridge started", "prefix", b.prefix)
}

// Stop publishes offline state, unsubscribes, waits for goroutines, and disconnects.
func (b *Bridge) Stop() {
	if b.unsub != nil {
		b.unsub()
	}
	b.cancel()
	// Don't close eventCh — the OnAll callback may still be in-flight after unsub().
	// The eventLoop exits via ctx cancellation instead.
	b.eventWg.Wait()
	b.discWg.Wait()
	b.publishBridgeStateSynchronous("offline")
	b.client.Disconnect(1000)
	b.logger.Info("MQTT bridge stopped")
}

func (b *Bridge) eventLoop() {
	defer b.eventWg.Done()
	for {
		select {
		case event := <-b.eventCh:
			b.handleEvent(event)
		case <-b.ctx.Done():
			return
		}
	}
}

func (b *Bridge) handleEvent(event coordinator.Event) {
	switch event.Type {
	case coordinator.EventAttributeReport:
		b.handleAttributeReport(event)
	case coordinator.EventPropertyUpdate:
		b.handlePropertyUpdate(event)
	case coordinator.EventDeviceAnnounce:
		// Publish discovery after a delay to let interview complete.
		b.discWg.Add(1)
		go func() {
			defer b.discWg.Done()
			b.delayedDiscovery(event)
		}()
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

	// Normalize values for Home Assistant compatibility.
	switch propName {
	case "state":
		// Convert bool OnOff to "ON"/"OFF" strings.
		if bv, ok := value.(bool); ok {
			if bv {
				value = "ON"
			} else {
				value = "OFF"
			}
		}
	case "occupancy":
		// Normalize occupancy to bool for HA binary_sensor.
		value = toBool(value)
	case "zone_status":
		// Normalize zone_status (bitmap16) to bool for HA binary_sensor.
		value = toBool(value)
	case "brightness":
		// HA JSON schema light needs color_mode alongside brightness.
		b.updateAndPublishState(ieee, "color_mode", "brightness")
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
	// Read device info outside the lock to avoid holding it during DB reads.
	dev, _ := b.coord.Devices().GetDevice(ieee)

	b.mu.Lock()
	state, ok := b.states[ieee]
	if !ok {
		state = make(map[string]any)
		b.states[ieee] = state
	}
	state[prop] = value

	if dev != nil {
		state["linkquality"] = dev.LQI
		state["last_seen"] = dev.LastSeen.Format(time.RFC3339)
	}

	payload := mustJSON(state)
	b.mu.Unlock()

	topic := b.prefix + "/" + b.cachedTopicName(ieee)
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

	// Clear retained state topic (publish empty payload).
	stateTopic := b.prefix + "/" + b.cachedTopicName(ieee)
	b.publish(stateTopic, nil, true)

	// Remove discovery entries.
	dev := &store.Device{IEEEAddress: ieee}
	for _, msg := range buildRemoveDiscovery(dev) {
		b.publish(msg.Topic, msg.Payload, true)
	}

	// Clear accumulated state and cached topic name.
	b.mu.Lock()
	delete(b.states, ieee)
	delete(b.topicNames, ieee)
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
			// Cache the topic name after interview completes.
			b.mu.Lock()
			b.topicNames[ieee] = deviceTopicName(dev)
			b.mu.Unlock()
			b.publishDeviceDiscovery(dev)
			b.subscribeDeviceCommands(dev)
			return
		}
	}
}

func (b *Bridge) publishBridgeState(state string) {
	topic := b.prefix + "/bridge/state"
	b.publish(topic, []byte(state), true)
}

// publishBridgeStateSynchronous publishes bridge state and waits for completion.
// Used during shutdown to ensure the message is sent before Disconnect().
func (b *Bridge) publishBridgeStateSynchronous(state string) {
	topic := b.prefix + "/bridge/state"
	token := b.client.Publish(topic, 1, true, []byte(state))
	if !token.WaitTimeout(5 * time.Second) {
		b.logger.Warn("MQTT publish timeout on shutdown", "topic", topic)
	} else if err := token.Error(); err != nil {
		b.logger.Warn("MQTT publish error on shutdown", "topic", topic, "err", err)
	}
}

func (b *Bridge) publishAllDiscovery() {
	devices, err := b.coord.Devices().ListDevices()
	if err != nil {
		b.logger.Error("list devices for discovery", "err", err)
		return
	}
	for _, dev := range devices {
		if dev.Interviewed {
			// Populate topic name cache on startup.
			b.mu.Lock()
			b.topicNames[dev.IEEEAddress] = deviceTopicName(dev)
			b.mu.Unlock()
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

	ctx, cancel := context.WithTimeout(b.coord.Context(), 10*time.Second)
	defer cancel()

	// Handle state command (ON/OFF).
	if state, ok := cmd["state"].(string); ok {
		ep := findEndpointWithCluster(dev, 0x0006)
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
		ep := findEndpointWithCluster(dev, 0x0008)
		if brightness < 0 {
			brightness = 0
		}
		if brightness > 254 {
			brightness = 254
		}
		level := uint8(brightness)
		// Move to Level with On/Off, transition time 5 (0.5s).
		cmdPayload := []byte{level, 0x05, 0x00}
		if err := b.coord.SendClusterCommand(ctx, dev.ShortAddress, ep, 0x0008, 0x04, cmdPayload); err != nil {
			b.logger.Warn("brightness command failed", "ieee", ieee, "err", err)
		} else {
			b.updateAndPublishState(ieee, "brightness", level)
			b.updateAndPublishState(ieee, "color_mode", "brightness")
		}
	}
}

// findEndpointWithCluster returns the endpoint ID that has the given cluster
// as an input cluster. Falls back to the first endpoint if not found.
func findEndpointWithCluster(dev *store.Device, clusterID uint16) uint8 {
	for _, ep := range dev.Endpoints {
		for _, cid := range ep.InClusters {
			if cid == clusterID {
				return ep.ID
			}
		}
	}
	return dev.Endpoints[0].ID
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

// cachedTopicName returns the MQTT topic name for a device, using a cache to avoid DB reads.
func (b *Bridge) cachedTopicName(ieee string) string {
	b.mu.Lock()
	name, ok := b.topicNames[ieee]
	b.mu.Unlock()
	if ok {
		return name
	}
	// Cache miss — read from DB.
	dev, err := b.coord.Devices().GetDevice(ieee)
	if err != nil {
		return ieee
	}
	name = deviceTopicName(dev)
	b.mu.Lock()
	b.topicNames[ieee] = name
	b.mu.Unlock()
	return name
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
	case 0x0500: // IAS Zone
		if attrName == "ZoneStatus" || attrName == "zone_status" {
			return "zone_status"
		}
	case 0x0001: // Power Configuration
		if attrName == "BatteryPercentageRemaining" || attrName == "battery_percentage_remaining" {
			return "battery"
		}
	}
	return ""
}

// toBool converts various types to a boolean (non-zero = true).
func toBool(v interface{}) bool {
	switch n := v.(type) {
	case bool:
		return n
	case uint8:
		return n != 0
	case uint16:
		return n != 0
	case uint32:
		return n != 0
	case int:
		return n != 0
	case int64:
		return n != 0
	case float64:
		return n != 0
	default:
		return false
	}
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

func shortRandomID() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
