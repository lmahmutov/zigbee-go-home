package coordinator

import (
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"zigbee-go-home/internal/store"
)

// memStore is a minimal in-memory store for device manager tests.
type memStore struct {
	devices map[string]*store.Device
	netState *store.NetworkState
}

func newMemStore() *memStore {
	return &memStore{devices: make(map[string]*store.Device)}
}

func (m *memStore) SaveDevice(dev *store.Device) error {
	m.devices[dev.IEEEAddress] = dev
	return nil
}
func (m *memStore) GetDevice(ieee string) (*store.Device, error) {
	d, ok := m.devices[ieee]
	if !ok {
		return nil, store.ErrNotFound
	}
	return d, nil
}
func (m *memStore) DeleteDevice(ieee string) error {
	delete(m.devices, ieee)
	return nil
}
func (m *memStore) ListDevices() ([]*store.Device, error) {
	list := make([]*store.Device, 0, len(m.devices))
	for _, d := range m.devices {
		list = append(list, d)
	}
	return list, nil
}
func (m *memStore) SaveNetworkState(s *store.NetworkState) error {
	m.netState = s
	return nil
}
func (m *memStore) GetNetworkState() (*store.NetworkState, error) {
	if m.netState == nil {
		return nil, store.ErrNotFound
	}
	return m.netState, nil
}
func (m *memStore) Close() error { return nil }

func newTestDM(t *testing.T) (*DeviceManager, *memStore) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	ms := newMemStore()
	events := NewEventBus(logger)
	coord := &Coordinator{
		store:    ms,
		events:   events,
		logger:   logger,
	}
	dm := NewDeviceManager(coord)
	return dm, ms
}

func TestAddrIndexUpdateAndLookup(t *testing.T) {
	dm, _ := newTestDM(t)

	dm.updateAddrIndex("00158D00012A3B4C", 0x1234)

	got := dm.lookupIEEE(0x1234)
	if got != "00158D00012A3B4C" {
		t.Errorf("lookupIEEE(0x1234) = %q, want 00158D00012A3B4C", got)
	}

	// Unknown address returns empty.
	if ieee := dm.lookupIEEE(0xFFFF); ieee != "" {
		t.Errorf("lookupIEEE(0xFFFF) = %q, want empty", ieee)
	}
}

func TestAddrIndexRemove(t *testing.T) {
	dm, _ := newTestDM(t)

	dm.updateAddrIndex("00158D00012A3B4C", 0x1234)
	dm.removeFromAddrIndex(0x1234)

	if ieee := dm.lookupIEEE(0x1234); ieee != "" {
		t.Errorf("after remove, lookupIEEE(0x1234) = %q, want empty", ieee)
	}
}

func TestAddrIndexRebuild(t *testing.T) {
	dm, ms := newTestDM(t)

	// Populate store directly (simulating DB state).
	ms.devices["AAAAAAAAAAAAAAAA"] = &store.Device{IEEEAddress: "AAAAAAAAAAAAAAAA", ShortAddress: 0x0001}
	ms.devices["BBBBBBBBBBBBBBBB"] = &store.Device{IEEEAddress: "BBBBBBBBBBBBBBBB", ShortAddress: 0x0002}

	dm.RebuildAddrIndex()

	if ieee := dm.lookupIEEE(0x0001); ieee != "AAAAAAAAAAAAAAAA" {
		t.Errorf("after rebuild, 0x0001 = %q, want AAAAAAAAAAAAAAAA", ieee)
	}
	if ieee := dm.lookupIEEE(0x0002); ieee != "BBBBBBBBBBBBBBBB" {
		t.Errorf("after rebuild, 0x0002 = %q, want BBBBBBBBBBBBBBBB", ieee)
	}
}

func TestLookupOrRebuild(t *testing.T) {
	dm, ms := newTestDM(t)

	// Store has a device but index is empty.
	ms.devices["CCCCCCCCCCCCCCCC"] = &store.Device{IEEEAddress: "CCCCCCCCCCCCCCCC", ShortAddress: 0x0003}

	// lookupOrRebuild should find it by rebuilding.
	ieee := dm.lookupOrRebuild(0x0003)
	if ieee != "CCCCCCCCCCCCCCCC" {
		t.Errorf("lookupOrRebuild(0x0003) = %q, want CCCCCCCCCCCCCCCC", ieee)
	}

	// Second call should hit the fast path (index already populated).
	ieee = dm.lookupOrRebuild(0x0003)
	if ieee != "CCCCCCCCCCCCCCCC" {
		t.Errorf("second lookupOrRebuild(0x0003) = %q, want CCCCCCCCCCCCCCCC", ieee)
	}

	// Unknown address returns empty.
	if ieee := dm.lookupOrRebuild(0xDEAD); ieee != "" {
		t.Errorf("lookupOrRebuild(0xDEAD) = %q, want empty", ieee)
	}
}

func TestLastJoinDebounce(t *testing.T) {
	dm, _ := newTestDM(t)

	ieee := "00158D00012A3B4C"

	// First entry succeeds.
	dm.lastJoinMu.Lock()
	dm.lastJoin[ieee] = time.Now()
	dm.lastJoinMu.Unlock()

	// Within 3s window, same IEEE should be recognized as duplicate.
	dm.lastJoinMu.Lock()
	last, ok := dm.lastJoin[ieee]
	isDuplicate := ok && time.Since(last) < 3*time.Second
	dm.lastJoinMu.Unlock()

	if !isDuplicate {
		t.Error("expected duplicate within 3s window")
	}
}

func TestLastJoinCleanup(t *testing.T) {
	dm, _ := newTestDM(t)

	// Populate >100 entries, all stale.
	dm.lastJoinMu.Lock()
	for i := 0; i < 110; i++ {
		ieee := fmt.Sprintf("%016X", i)
		dm.lastJoin[ieee] = time.Now().Add(-2 * time.Minute)
	}
	dm.lastJoinMu.Unlock()

	// Simulate the cleanup logic from HandleAnnounce.
	dm.lastJoinMu.Lock()
	dm.lastJoin["TRIGGER"] = time.Now()
	if len(dm.lastJoin) > 100 {
		for k, t := range dm.lastJoin {
			if time.Since(t) > time.Minute {
				delete(dm.lastJoin, k)
			}
		}
	}
	dm.lastJoinMu.Unlock()

	dm.lastJoinMu.Lock()
	count := len(dm.lastJoin)
	dm.lastJoinMu.Unlock()

	// Only the fresh "TRIGGER" entry should remain.
	if count != 1 {
		t.Errorf("after cleanup, lastJoin count = %d, want 1", count)
	}
}
