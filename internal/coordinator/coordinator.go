package coordinator

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"

	"zigbee-go-home/internal/ncp"
	"zigbee-go-home/internal/store"
	"zigbee-go-home/internal/zcl"
)

// Config holds coordinator configuration.
type Config struct {
	Channel  uint8
	PanID    uint16
	ExtPanID [8]byte
}

// NCPConfig holds NCP hardware/port configuration for display purposes.
type NCPConfig struct {
	Type string
	Port string
	Baud int
}

// ParseIEEE parses "DD:DD:DD:DD:DD:DD:DD:DD" or "DDDDDDDDDDDDDDDD" into [8]byte.
func ParseIEEE(s string) ([8]byte, error) {
	var result [8]byte
	s = strings.ReplaceAll(s, ":", "")
	b, err := hex.DecodeString(s)
	if err != nil {
		return result, fmt.Errorf("parse ieee address: %w", err)
	}
	if len(b) != 8 {
		return result, fmt.Errorf("ieee address must be 8 bytes, got %d", len(b))
	}
	copy(result[:], b)
	return result, nil
}

// ParseExtPanID parses "DD:DD:DD:DD:DD:DD:DD:DD" into [8]byte.
func ParseExtPanID(s string) ([8]byte, error) {
	return ParseIEEE(s)
}

// Coordinator manages the Zigbee network via an NCP backend.
type Coordinator struct {
	ncp        ncp.NCP
	store      store.Store
	registry   *zcl.Registry
	deviceDB   *DeviceDB
	events     *EventBus
	devices    *DeviceManager
	logger     *slog.Logger
	config     Config
	ncpConfig  NCPConfig
	localIEEE [8]byte // coordinator's own IEEE address, cached at Start
	ctx       context.Context
	cancel       context.CancelFunc
}

// New creates a new Coordinator using the nRF52840 NCP backend.
func New(backend ncp.NCP, st store.Store, registry *zcl.Registry, deviceDB *DeviceDB, events *EventBus, cfg Config, ncpCfg NCPConfig, logger *slog.Logger) *Coordinator {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Coordinator{
		ncp:       backend,
		store:     st,
		registry:  registry,
		deviceDB:  deviceDB,
		events:    events,
		logger:    logger,
		config:    cfg,
		ncpConfig: ncpCfg,
		ctx:       ctx,
		cancel:    cancel,
	}
	c.devices = NewDeviceManager(c)
	c.devices.RebuildAddrIndex()
	c.registerIndicationHandlers()
	return c
}

// Context returns the coordinator's context, which is cancelled on Stop().
func (c *Coordinator) Context() context.Context {
	return c.ctx
}

// Start initializes the NCP and forms or resumes the network.
// If a network was previously formed with the same parameters, it resumes
// from NCP NVRAM instead of re-forming (which would generate a new network
// key and orphan all paired devices).
func (c *Coordinator) Start(ctx context.Context) error {
	c.logger.Info("initializing NCP...")

	// Try to resume an existing network if our DB says it was formed with matching params.
	if c.canResumeNetwork() {
		c.logger.Info("resuming existing network...")
		// Soft reset (no NVRAM erase) to get NCP into a clean LL protocol
		// state. Without this, the NCP ignores our packets because its
		// packet sequence numbers are stale from the previous session.
		if err := c.ncp.Reset(ctx); err != nil {
			return fmt.Errorf("ncp reset (resume): %w", err)
		}
		if err := c.ncp.Init(ctx); err != nil {
			return fmt.Errorf("ncp init: %w", err)
		}
		if err := c.ncp.StartNetwork(ctx); err == nil {
			c.cacheLocalIEEE(ctx)
			c.logger.Info("network resumed", "channel", c.config.Channel, "panID", fmt.Sprintf("0x%04X", c.config.PanID))
			c.events.Emit(Event{Type: EventNetworkState, Data: "started"})
			return nil
		}
		c.logger.Warn("network resume failed, re-forming")
	}

	// Form a new network.
	// First try with a simple reset (no USB re-enumeration) — works when
	// the NCP was already rebooted manually or is in a clean state.
	ncpCfg := ncp.NetworkConfig{
		Channel:  c.config.Channel,
		PanID:    c.config.PanID,
		ExtPanID: c.config.ExtPanID,
	}

	c.logger.Info("forming new network (simple reset)...")
	if err := c.ncp.Reset(ctx); err != nil {
		return fmt.Errorf("ncp reset: %w", err)
	}
	if err := c.ncp.Init(ctx); err != nil {
		return fmt.Errorf("ncp init: %w", err)
	}
	if err := c.ncp.FormNetwork(ctx, ncpCfg); err != nil {
		// Formation failed — NCP may have stale NVRAM state. Factory reset and retry.
		c.logger.Warn("formation failed, trying factory reset", "err", err)
		if err := c.ncp.FactoryReset(ctx); err != nil {
			return fmt.Errorf("ncp factory reset: %w", err)
		}
		if err := c.ncp.Init(ctx); err != nil {
			return fmt.Errorf("ncp init after factory reset: %w", err)
		}
		if err := c.ncp.FormNetwork(ctx, ncpCfg); err != nil {
			return fmt.Errorf("form network: %w", err)
		}
	}
	if err := c.ncp.StartNetwork(ctx); err != nil {
		return fmt.Errorf("start network: %w", err)
	}

	c.saveNetworkState()
	c.cacheLocalIEEE(ctx)
	c.logger.Info("network formed", "channel", c.config.Channel, "panID", fmt.Sprintf("0x%04X", c.config.PanID))
	c.events.Emit(Event{Type: EventNetworkState, Data: "started"})
	return nil
}

func (c *Coordinator) cacheLocalIEEE(ctx context.Context) {
	ieee, err := c.ncp.GetLocalIEEE(ctx)
	if err != nil {
		c.logger.Warn("get coordinator IEEE", "err", err)
		return
	}
	c.localIEEE = ieee
	c.logger.Info("coordinator IEEE", "ieee", fmt.Sprintf("%016X", ieee))
}

// LocalIEEE returns the coordinator's own IEEE address.
func (c *Coordinator) LocalIEEE() [8]byte {
	return c.localIEEE
}

func (c *Coordinator) saveNetworkState() {
	extPanStr := fmt.Sprintf("%X", c.config.ExtPanID)
	if err := c.store.SaveNetworkState(&store.NetworkState{
		Channel:  c.config.Channel,
		PanID:    c.config.PanID,
		ExtPanID: extPanStr,
		Formed:   true,
	}); err != nil {
		c.logger.Error("save network state", "err", err)
	}
}

// canResumeNetwork checks if the previously formed network matches current config.
func (c *Coordinator) canResumeNetwork() bool {
	ns, err := c.store.GetNetworkState()
	if err != nil || !ns.Formed {
		return false
	}
	extPanStr := fmt.Sprintf("%X", c.config.ExtPanID)
	return ns.Channel == c.config.Channel &&
		ns.PanID == c.config.PanID &&
		ns.ExtPanID == extPanStr
}

// Stop cancels the coordinator context and waits for in-progress interviews.
func (c *Coordinator) Stop() {
	c.cancel()
	c.devices.CancelAllInterviews()
}

// PermitJoin opens or closes the network for device joining.
func (c *Coordinator) PermitJoin(ctx context.Context, duration uint8) error {
	if err := c.ncp.PermitJoin(ctx, duration); err != nil {
		return fmt.Errorf("permit join: %w", err)
	}
	c.logger.Info("permit join", "duration", duration)
	c.events.Emit(Event{Type: EventPermitJoin, Data: map[string]interface{}{"duration": duration}})
	return nil
}

// NetworkInfo returns current network information from cached config.
func (c *Coordinator) NetworkInfo() map[string]interface{} {
	info := map[string]interface{}{
		"channel":          c.config.Channel,
		"pan_id":           fmt.Sprintf("0x%04X", c.config.PanID),
		"ext_pan_id":       fmt.Sprintf("%X", c.config.ExtPanID),
		"ncp_type":         c.ncpConfig.Type,
		"port":             c.ncpConfig.Port,
		"baud":             c.ncpConfig.Baud,
		"coordinator_ieee": fmt.Sprintf("%016X", c.localIEEE),
	}
	if ncpInfo := c.ncp.GetNCPInfo(); ncpInfo != nil {
		info["fw_version"] = ncpInfo.FWVersion
		info["stack_version"] = ncpInfo.StackVersion
		info["protocol_version"] = ncpInfo.ProtocolVersion
	}
	return info
}

// NCP returns the underlying NCP backend.
func (c *Coordinator) NCP() ncp.NCP {
	return c.ncp
}

// Store returns the store.
func (c *Coordinator) Store() store.Store {
	return c.store
}

// Registry returns the ZCL registry.
func (c *Coordinator) Registry() *zcl.Registry {
	return c.registry
}

// DeviceDB returns the device definitions database.
func (c *Coordinator) DeviceDB() *DeviceDB {
	return c.deviceDB
}

// Events returns the event bus.
func (c *Coordinator) Events() *EventBus {
	return c.events
}

// Devices returns the device manager.
func (c *Coordinator) Devices() *DeviceManager {
	return c.devices
}

func (c *Coordinator) registerIndicationHandlers() {
	c.ncp.OnDeviceJoined(func(evt ncp.DeviceJoinedEvent) {
		c.devices.HandleJoin(evt)
	})
	c.ncp.OnDeviceLeft(func(evt ncp.DeviceLeftEvent) {
		c.devices.HandleLeave(evt)
	})
	c.ncp.OnDeviceAnnounce(func(evt ncp.DeviceAnnounceEvent) {
		c.devices.HandleAnnounce(evt)
	})
	c.ncp.OnAttributeReport(func(evt ncp.AttributeReportEvent) {
		c.devices.HandleAttributeReport(evt)
	})
}
