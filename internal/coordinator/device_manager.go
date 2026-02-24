package coordinator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	"zigbee-go-home/internal/ncp"
	"zigbee-go-home/internal/store"
	"zigbee-go-home/internal/zcl"
)

type interviewEntry struct {
	cancel context.CancelFunc
	gen    uint64
}

// DeviceManager handles device lifecycle (join, leave, interview).
type DeviceManager struct {
	coord  *Coordinator
	logger *slog.Logger

	// Interview cancellation: tracks active interview cancel funcs by IEEE.
	interviewMu      sync.Mutex
	interviewCancels map[string]interviewEntry
	interviewGen     atomic.Uint64
	interviewWg      sync.WaitGroup

	// Debounce duplicate join events (e.g., unsecure→secure join transitions).
	lastJoinMu sync.Mutex
	lastJoin   map[string]time.Time

	// In-memory short address -> IEEE index for fast lookup.
	addrMu    sync.RWMutex
	addrIndex map[uint16]string
}

// NewDeviceManager creates a new device manager.
func NewDeviceManager(coord *Coordinator) *DeviceManager {
	return &DeviceManager{
		coord:            coord,
		logger:           coord.logger.With("component", "device_manager"),
		interviewCancels: make(map[string]interviewEntry),
		lastJoin:         make(map[string]time.Time),
		addrIndex:        make(map[uint16]string),
	}
}

// CancelAllInterviews cancels all running interview goroutines and waits for them.
func (dm *DeviceManager) CancelAllInterviews() {
	dm.interviewMu.Lock()
	for ieee, entry := range dm.interviewCancels {
		entry.cancel()
		delete(dm.interviewCancels, ieee)
	}
	dm.interviewMu.Unlock()
	dm.interviewWg.Wait()
}

// updateAddrIndex updates the short address -> IEEE mapping.
func (dm *DeviceManager) updateAddrIndex(ieee string, shortAddr uint16) {
	dm.addrMu.Lock()
	dm.addrIndex[shortAddr] = ieee
	dm.addrMu.Unlock()
}

// removeFromAddrIndex removes a short address from the index.
func (dm *DeviceManager) removeFromAddrIndex(shortAddr uint16) {
	dm.addrMu.Lock()
	delete(dm.addrIndex, shortAddr)
	dm.addrMu.Unlock()
}

// lookupIEEE finds IEEE address by short address from in-memory index.
func (dm *DeviceManager) lookupIEEE(shortAddr uint16) string {
	dm.addrMu.RLock()
	defer dm.addrMu.RUnlock()
	return dm.addrIndex[shortAddr]
}

// deviceName returns a human-readable display name for a device.
// Returns "Manufacturer Model" if available, or empty string for unknown devices.
func deviceName(dev *store.Device) string {
	if dev == nil {
		return ""
	}
	if dev.FriendlyName != "" {
		return dev.FriendlyName
	}
	if dev.Manufacturer != "" || dev.Model != "" {
		name := dev.Manufacturer
		if dev.Model != "" {
			if name != "" {
				name += " "
			}
			name += dev.Model
		}
		return name
	}
	return ""
}

// RebuildAddrIndex loads all devices from store and populates the index.
func (dm *DeviceManager) RebuildAddrIndex() {
	devices, err := dm.coord.Store().ListDevices()
	if err != nil {
		dm.logger.Error("rebuild addr index", "err", err)
		return
	}
	dm.addrMu.Lock()
	clear(dm.addrIndex)
	for _, d := range devices {
		dm.addrIndex[d.ShortAddress] = d.IEEEAddress
	}
	dm.addrMu.Unlock()
}

// HandleJoin processes a device join event.
func (dm *DeviceManager) HandleJoin(evt ncp.DeviceJoinedEvent) {
	ieee := fmt.Sprintf("%016X", evt.IEEEAddr)

	dm.updateAddrIndex(ieee, evt.ShortAddr)

	// Check if device already exists (rejoin case).
	dev, err := dm.coord.Store().GetDevice(ieee)
	if err == nil {
		// Device exists: update address and LastSeen only, preserving interview data.
		dev.ShortAddress = evt.ShortAddr
		dev.LastSeen = time.Now()
	} else {
		// New device.
		dev = &store.Device{
			IEEEAddress:  ieee,
			ShortAddress: evt.ShortAddr,
			JoinedAt:     time.Now(),
			LastSeen:     time.Now(),
		}
	}

	dm.logger.Info("device joined", "ieee", ieee, "short", fmt.Sprintf("0x%04X", evt.ShortAddr), "name", deviceName(dev))

	if err := dm.coord.Store().SaveDevice(dev); err != nil {
		dm.logger.Error("save device", "err", err, "ieee", ieee)
		return
	}

	dm.coord.Events().Emit(Event{
		Type: EventDeviceJoined,
		Data: map[string]interface{}{
			"ieee":       ieee,
			"short_addr": evt.ShortAddr,
		},
	})

	// Do NOT start interview here. DevUpdateInd fires before the TC key
	// exchange completes — the device doesn't have the network key yet and
	// can't respond to ZDO requests. Wait for DevAnnceInd (HandleAnnounce)
	// which arrives after key exchange succeeds.
}

// HandleLeave processes a device leave event: cancels interview, removes from
// address index, deletes from store, and emits EventDeviceLeft.
func (dm *DeviceManager) HandleLeave(evt ncp.DeviceLeftEvent) {
	ieee := fmt.Sprintf("%016X", evt.IEEEAddr)
	dev, _ := dm.coord.Store().GetDevice(ieee)
	name := deviceName(dev)
	dm.logger.Info("device left", "ieee", ieee, "name", name)

	// Cancel any in-progress interview for this device.
	dm.interviewMu.Lock()
	if entry, ok := dm.interviewCancels[ieee]; ok {
		entry.cancel()
		delete(dm.interviewCancels, ieee)
	}
	dm.interviewMu.Unlock()

	dm.lastJoinMu.Lock()
	delete(dm.lastJoin, ieee)
	dm.lastJoinMu.Unlock()

	// Remove from addr index by IEEE (ShortAddr may be 0 for NwkLeaveInd).
	dm.addrMu.Lock()
	for addr, storedIEEE := range dm.addrIndex {
		if storedIEEE == ieee {
			delete(dm.addrIndex, addr)
			break
		}
	}
	dm.addrMu.Unlock()

	if err := dm.coord.Store().DeleteDevice(ieee); err != nil {
		dm.logger.Error("delete device on leave", "err", err, "ieee", ieee)
	} else {
		dm.logger.Info("device removed from store", "ieee", ieee, "name", name)
	}

	dm.coord.Events().Emit(Event{
		Type: EventDeviceLeft,
		Data: map[string]interface{}{"ieee": ieee},
	})
}

// HandleAnnounce processes a device announce event.
func (dm *DeviceManager) HandleAnnounce(evt ncp.DeviceAnnounceEvent) {
	ieee := fmt.Sprintf("%016X", evt.IEEEAddr)
	dev, _ := dm.coord.Store().GetDevice(ieee)
	dm.logger.Info("device announce", "ieee", ieee, "short", fmt.Sprintf("0x%04X", evt.ShortAddr), "name", deviceName(dev))

	dm.updateAddrIndex(ieee, evt.ShortAddr)

	dev, err := dm.coord.Store().GetDevice(ieee)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			// Real DB error — don't create a new device entry.
			dm.logger.Error("get device on announce", "err", err, "ieee", ieee)
			return
		}
		dev = &store.Device{
			IEEEAddress: ieee,
			JoinedAt:    time.Now(),
		}
	}
	dev.ShortAddress = evt.ShortAddr
	dev.LastSeen = time.Now()

	if err := dm.coord.Store().SaveDevice(dev); err != nil {
		dm.logger.Error("save device on announce", "err", err)
	}

	dm.coord.Events().Emit(Event{
		Type: EventDeviceAnnounce,
		Data: map[string]interface{}{
			"ieee":       ieee,
			"short_addr": evt.ShortAddr,
		},
	})

	// Start interview on announce — this means the TC key exchange succeeded
	// and the device has the network key, so ZDO requests will work.
	dm.interviewMu.Lock()
	_, interviewing := dm.interviewCancels[ieee]
	dm.interviewMu.Unlock()

	if interviewing {
		dm.logger.Info("announce during interview, address updated", "ieee", ieee,
			"short", fmt.Sprintf("0x%04X", evt.ShortAddr), "name", deviceName(dev))
		return
	}

	// Debounce: avoid duplicate interviews from rapid announce events.
	dm.lastJoinMu.Lock()
	if last, ok := dm.lastJoin[ieee]; ok && time.Since(last) < 3*time.Second {
		dm.lastJoinMu.Unlock()
		dm.logger.Debug("duplicate announce, interview already started", "ieee", ieee)
		return
	}
	dm.lastJoin[ieee] = time.Now()
	// Evict stale entries to prevent unbounded growth.
	if len(dm.lastJoin) > 50 {
		for k, t := range dm.lastJoin {
			if time.Since(t) > time.Minute {
				delete(dm.lastJoin, k)
			}
		}
	}
	dm.lastJoinMu.Unlock()

	dm.interviewWg.Add(1)
	go dm.Interview(ieee)
}

// lookupOrRebuild looks up an IEEE address by short address from the in-memory
// index. If not found, rebuilds the index from the store under a write lock
// with a double-check to avoid redundant rebuilds.
func (dm *DeviceManager) lookupOrRebuild(shortAddr uint16) string {
	// Fast path: read lock.
	dm.addrMu.RLock()
	ieee := dm.addrIndex[shortAddr]
	dm.addrMu.RUnlock()
	if ieee != "" {
		return ieee
	}

	// Slow path: rebuild under write lock.
	dm.addrMu.Lock()
	defer dm.addrMu.Unlock()

	// Double-check after acquiring write lock.
	if ieee = dm.addrIndex[shortAddr]; ieee != "" {
		return ieee
	}

	devices, err := dm.coord.Store().ListDevices()
	if err != nil {
		dm.logger.Error("rebuild addr index for lookup", "err", err)
		return ""
	}
	clear(dm.addrIndex)
	for _, d := range devices {
		dm.addrIndex[d.ShortAddress] = d.IEEEAddress
		if d.ShortAddress == shortAddr {
			ieee = d.IEEEAddress
		}
	}
	return ieee
}

// HandleAttributeReport processes an attribute report event.
func (dm *DeviceManager) HandleAttributeReport(evt ncp.AttributeReportEvent) {
	ieee := dm.lookupOrRebuild(evt.SrcAddr)

	var decoded interface{}
	if len(evt.Value) > 0 {
		val, _, decErr := zcl.DecodeValue(evt.DataType, evt.Value)
		if decErr == nil {
			decoded = val
		} else {
			decoded = fmt.Sprintf("%X", evt.Value)
		}
	}

	clusterName := fmt.Sprintf("0x%04X", evt.ClusterID)
	attrName := fmt.Sprintf("0x%04X", evt.AttrID)
	if cluster := dm.coord.Registry().Get(evt.ClusterID); cluster != nil {
		clusterName = cluster.Name
		if attr := cluster.FindAttribute(evt.AttrID); attr != nil {
			attrName = attr.Name
		}
	}

	var dev *store.Device
	if ieee != "" {
		if d, err := dm.coord.Store().GetDevice(ieee); err == nil {
			dev = d
			dev.LastSeen = time.Now()
			if evt.LQI > 0 {
				dev.LQI = evt.LQI
				dev.RSSI = evt.RSSI
			}
			if saveErr := dm.coord.Store().SaveDevice(dev); saveErr != nil {
				dm.logger.Error("save device last_seen", "err", saveErr, "ieee", ieee)
			}
		}
	}

	dm.logger.Info("attribute report",
		"ieee", ieee,
		"name", deviceName(dev),
		"cluster", clusterName,
		"attr", attrName,
		"value", decoded,
	)

	dm.coord.Events().Emit(Event{
		Type: EventAttributeReport,
		Data: map[string]interface{}{
			"ieee":         ieee,
			"short_addr":   evt.SrcAddr,
			"endpoint":     evt.SrcEP,
			"cluster_id":   evt.ClusterID,
			"cluster_name": clusterName,
			"attr_id":      evt.AttrID,
			"attr_name":    attrName,
			"value":        decoded,
		},
	})

	// Emit property_update for well-known ZCL attributes so Blockly
	// automations can trigger on standard properties like on_off, temperature, etc.
	dm.emitStandardProperty(ieee, dev, evt, decoded)

	dm.processProperties(ieee, dev, evt, decoded)
}

// standardPropertyMap maps well-known ZCL cluster+attribute pairs to property names.
// These are used to emit property_update events for common attributes so that
// Blockly automations can trigger on simple property names like "on_off".
var standardPropertyMap = map[uint16]map[uint16]string{
	0x0006: {0x0000: "on_off"},                      // On/Off → OnOff
	0x0008: {0x0000: "brightness"},                   // Level Control → CurrentLevel
	0x0300: {0x0000: "hue", 0x0001: "saturation"},   // Color Control
	0x0402: {0x0000: "temperature"},                  // Temperature Measurement
	0x0403: {0x0000: "pressure"},                     // Pressure Measurement
	0x0405: {0x0000: "humidity"},                     // Relative Humidity
	0x0406: {0x0000: "occupancy"},                    // Occupancy Sensing
	0x0400: {0x0000: "illuminance"},                  // Illuminance Measurement
	0x0001: {0x0021: "battery"},                      // Power Configuration → BatteryPercentage
	0x0500: {0x0002: "zone_status"},                  // IAS Zone
	0x0B04: {0x050B: "power"},                        // Electrical Measurement → ActivePower
	0x0702: {0x0000: "energy"},                       // Metering → CurrentSummation
	0x000C: {0x0055: "analog_value"},                 // Analog Input → PresentValue
	0x0012: {0x0055: "multistate_value"},             // Multistate Input → PresentValue
}

// emitStandardProperty emits a property_update event for well-known ZCL attributes.
func (dm *DeviceManager) emitStandardProperty(ieee string, dev *store.Device, evt ncp.AttributeReportEvent, decoded interface{}) {
	if ieee == "" || decoded == nil {
		return
	}

	attrs, ok := standardPropertyMap[evt.ClusterID]
	if !ok {
		return
	}
	propName, ok := attrs[evt.AttrID]
	if !ok {
		return
	}

	// Store the property value on the device.
	if dev != nil {
		if dev.Properties == nil {
			dev.Properties = make(map[string]any)
		}
		dev.Properties[propName] = decoded
		if saveErr := dm.coord.Store().SaveDevice(dev); saveErr != nil {
			dm.logger.Error("save standard property", "err", saveErr, "ieee", ieee)
		}
	}

	dm.coord.Events().Emit(Event{
		Type: EventPropertyUpdate,
		Data: map[string]interface{}{
			"ieee":     ieee,
			"property": propName,
			"value":    decoded,
		},
	})
}

// Interview queries a device for its endpoints and descriptors.
// Retries up to 3 times, re-reading the device from store each time
// to pick up any short address changes from re-joins.
func (dm *DeviceManager) Interview(ieee string) {
	gen := dm.interviewGen.Add(1)

	defer func() {
		dm.interviewMu.Lock()
		if entry, ok := dm.interviewCancels[ieee]; ok && entry.gen == gen {
			delete(dm.interviewCancels, ieee)
		}
		dm.interviewMu.Unlock()
		dm.interviewWg.Done()
	}()

	ctx, cancel := context.WithTimeout(dm.coord.Context(), 3*time.Minute)
	defer cancel()

	dm.interviewMu.Lock()
	// Cancel any previous interview for this device.
	if prev, ok := dm.interviewCancels[ieee]; ok {
		prev.cancel()
	}
	dm.interviewCancels[ieee] = interviewEntry{cancel: cancel, gen: gen}
	dm.interviewMu.Unlock()

	const maxRetries = 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Re-read device from store each attempt to get latest short address.
		dev, err := dm.coord.Store().GetDevice(ieee)
		if err != nil {
			dm.logger.Error("interview: device not found", "ieee", ieee)
			return
		}

		name := deviceName(dev)
		dm.logger.Info("starting interview", "ieee", ieee, "name", name,
			"short", fmt.Sprintf("0x%04X", dev.ShortAddress), "attempt", attempt)

		endpoints, err := dm.coord.NCP().ActiveEndpoints(ctx, dev.ShortAddress)
		if err != nil {
			dm.logger.Warn("interview: active EP failed", "err", err, "ieee", ieee, "name", name, "attempt", attempt)
			if ctx.Err() != nil {
				return // Overall context expired, give up.
			}
			if attempt < maxRetries {
				jitter := time.Duration(rand.IntN(3001)) * time.Millisecond
				delay := 5*time.Second + jitter
				dm.logger.Info("interview: will retry", "ieee", ieee, "delay", delay)
				select {
				case <-time.After(delay):
				case <-ctx.Done():
					return
				}
			}
			continue
		}

		// Read model/manufacturer early so we have the real name for logging.
		if len(endpoints) > 0 {
			dm.readBasicAttributes(ctx, dev, endpoints[0])
		}

		// Look up device definition by manufacturer+model.
		var def *DeviceDefinition
		if db := dm.coord.DeviceDB(); db != nil {
			def = db.Lookup(dev.Manufacturer, dev.Model)
		}

		// Set friendly name: prefer definition, fall back to model.
		if def != nil && def.FriendlyName != "" {
			dev.FriendlyName = def.FriendlyName
		} else if dev.FriendlyName == "" && dev.Model != "" {
			dev.FriendlyName = dev.Model
		}
		name = deviceName(dev)

		dev.Endpoints = make([]store.Endpoint, 0, len(endpoints))

		for _, ep := range endpoints {
			sd, err := dm.coord.NCP().SimpleDescriptor(ctx, dev.ShortAddress, ep)
			if err != nil {
				dm.logger.Warn("interview: simple desc", "err", err, "ieee", ieee, "name", name, "ep", ep)
				continue
			}

			endpoint := store.Endpoint{
				ID:          ep,
				ProfileID:   sd.ProfileID,
				DeviceID:    sd.DeviceID,
				InClusters:  sd.InClusters,
				OutClusters: sd.OutClusters,
			}
			dev.Endpoints = append(dev.Endpoints, endpoint)

			dm.logger.Info("endpoint discovered",
				"ieee", ieee, "name", name, "ep", ep,
				"profile", fmt.Sprintf("0x%04X", sd.ProfileID),
				"device", fmt.Sprintf("0x%04X", sd.DeviceID),
				"in_clusters", len(sd.InClusters),
				"out_clusters", len(sd.OutClusters),
			)
		}

		// Configure bindings immediately while the device is still awake.
		if def != nil {
			dm.configureDevice(ctx, dev, def)
		} else {
			dm.logger.Info("no device definition found, skipping configure",
				"ieee", ieee, "name", name,
				"manufacturer", dev.Manufacturer, "model", dev.Model)
		}

		dev.Interviewed = true
		if err := dm.coord.Store().SaveDevice(dev); err != nil {
			dm.logger.Error("interview: save", "err", err, "ieee", ieee, "name", name)
		}
		dm.logger.Info("interview complete", "ieee", ieee, "name", name, "endpoints", len(dev.Endpoints))
		return
	}

	dm.logger.Error("interview failed after retries", "ieee", ieee, "attempts", maxRetries)
}

func (dm *DeviceManager) readBasicAttributes(ctx context.Context, dev *store.Device, ep uint8) {
	results, err := dm.coord.NCP().ReadAttributes(ctx, ncp.ReadAttributesRequest{
		DstAddr:   dev.ShortAddress,
		DstEP:     ep,
		ClusterID: 0x0000,
		AttrIDs:   []uint16{0x0004, 0x0005},
	})
	if err != nil {
		dm.logger.Warn("read basic attributes", "err", err)
		return
	}

	for _, r := range results {
		if r.Status != 0 || len(r.Value) == 0 {
			continue
		}
		val, _, err := zcl.DecodeValue(r.DataType, r.Value)
		if err != nil {
			continue
		}
		switch r.AttrID {
		case 0x0004:
			if s, ok := val.(string); ok {
				dev.Manufacturer = s
			}
		case 0x0005:
			if s, ok := val.(string); ok {
				dev.Model = s
			}
		}
	}
}

// configureDevice binds clusters and sets up reporting based on a device definition.
// Must be called right after interview while the device is still awake.
func (dm *DeviceManager) configureDevice(ctx context.Context, dev *store.Device, def *DeviceDefinition) {
	if len(dev.Endpoints) == 0 {
		return
	}
	name := deviceName(dev)
	coordIEEE := dm.coord.LocalIEEE()

	devIEEE, err := ParseIEEE(dev.IEEEAddress)
	if err != nil {
		dm.logger.Warn("configure: parse device IEEE", "err", err)
		return
	}

	// Bind and configure reporting across all endpoints.
	for _, ep := range dev.Endpoints {
		// Bind clusters listed in the device definition (only if present as OUT clusters).
		for _, cluster := range def.Bind {
			if !hasOutCluster(ep, cluster) {
				continue
			}
			err := dm.coord.NCP().Bind(ctx, ncp.BindRequest{
				TargetShortAddr: dev.ShortAddress,
				SrcIEEE:         devIEEE,
				SrcEP:           ep.ID,
				ClusterID:       cluster,
				DstIEEE:         coordIEEE,
				DstEP:           1,
			})
			if err != nil {
				dm.logger.Warn("configure: bind", "err", err, "name", name, "ep", ep.ID, "cluster", fmt.Sprintf("0x%04X", cluster))
			} else {
				dm.logger.Info("bound cluster", "name", name, "ep", ep.ID, "cluster", fmt.Sprintf("0x%04X", cluster))
			}
		}

		// Configure reporting entries from the device definition.
		for _, r := range def.Reporting {
			if !hasInCluster(ep, r.Cluster) {
				continue
			}
			change := []byte{byte(r.Change)}
			if r.Change > 255 {
				change = []byte{byte(r.Change), byte(r.Change >> 8)}
			}
			err := dm.coord.NCP().ConfigureReporting(ctx, ncp.ConfigureReportingRequest{
				DstAddr:      dev.ShortAddress,
				DstEP:        ep.ID,
				ClusterID:    r.Cluster,
				AttrID:       r.Attribute,
				DataType:     r.Type,
				MinInterval:  r.Min,
				MaxInterval:  r.Max,
				ReportChange: change,
			})
			if err != nil {
				dm.logger.Warn("configure: reporting", "err", err, "name", name,
					"ep", ep.ID,
					"cluster", fmt.Sprintf("0x%04X", r.Cluster),
					"attr", fmt.Sprintf("0x%04X", r.Attribute))
			} else {
				dm.logger.Info("configured reporting", "name", name,
					"ep", ep.ID,
					"cluster", fmt.Sprintf("0x%04X", r.Cluster),
					"attr", fmt.Sprintf("0x%04X", r.Attribute))
			}
		}
	}
}

func hasOutCluster(ep store.Endpoint, cluster uint16) bool {
	for _, c := range ep.OutClusters {
		if c == cluster {
			return true
		}
	}
	return false
}

func hasInCluster(ep store.Endpoint, cluster uint16) bool {
	for _, c := range ep.InClusters {
		if c == cluster {
			return true
		}
	}
	return false
}

// RemoveDevice sends a ZDO leave request, cancels any in-progress interview,
// removes from addr index, and deletes from store.
func (dm *DeviceManager) RemoveDevice(ieee string) error {
	// Cancel interview if running.
	dm.interviewMu.Lock()
	if entry, ok := dm.interviewCancels[ieee]; ok {
		entry.cancel()
		delete(dm.interviewCancels, ieee)
	}
	dm.interviewMu.Unlock()

	// Send ZDO Mgmt Leave to remove the device from the network.
	dev, err := dm.coord.Store().GetDevice(ieee)
	if err == nil {
		var ieeeBytes [8]byte
		if parsed, parseErr := ParseIEEE(ieee); parseErr == nil {
			ieeeBytes = parsed
			ctx, cancel := context.WithTimeout(dm.coord.Context(), 10*time.Second)
			defer cancel()
			if leaveErr := dm.coord.NCP().MgmtLeave(ctx, dev.ShortAddress, ieeeBytes); leaveErr != nil {
				dm.logger.Warn("mgmt leave request failed", "ieee", ieee, "name", deviceName(dev), "err", leaveErr)
			} else {
				dm.logger.Info("device removed from network", "ieee", ieee, "name", deviceName(dev))
			}
		}
	}

	// Remove from addr index by scanning for the IEEE value.
	dm.addrMu.Lock()
	for addr, storedIEEE := range dm.addrIndex {
		if storedIEEE == ieee {
			delete(dm.addrIndex, addr)
			break
		}
	}
	dm.addrMu.Unlock()

	return dm.coord.Store().DeleteDevice(ieee)
}

// ListDevices returns all known devices.
func (dm *DeviceManager) ListDevices() ([]*store.Device, error) {
	return dm.coord.Store().ListDevices()
}

// GetDevice returns a device by IEEE address.
func (dm *DeviceManager) GetDevice(ieee string) (*store.Device, error) {
	return dm.coord.Store().GetDevice(ieee)
}

// SaveDevice persists a device to the store.
func (dm *DeviceManager) SaveDevice(dev *store.Device) error {
	return dm.coord.Store().SaveDevice(dev)
}
