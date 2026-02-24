package store

import (
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *BoltStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := NewBoltStore(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestSaveAndGetDevice(t *testing.T) {
	s := newTestStore(t)

	dev := &Device{
		IEEEAddress:  "00158D00012A3B4C",
		ShortAddress: 0x1234,
		Manufacturer: "LUMI",
		Model:        "lumi.sensor_magnet.aq2",
		Interviewed:  true,
		JoinedAt:     time.Now().Truncate(time.Millisecond),
		LastSeen:     time.Now().Truncate(time.Millisecond),
		Endpoints: []Endpoint{
			{ID: 1, ProfileID: 0x0104, DeviceID: 0x0015, InClusters: []uint16{0, 6}, OutClusters: []uint16{6}},
		},
	}

	if err := s.SaveDevice(dev); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetDevice(dev.IEEEAddress)
	if err != nil {
		t.Fatal(err)
	}

	if got.IEEEAddress != dev.IEEEAddress {
		t.Errorf("ieee = %q, want %q", got.IEEEAddress, dev.IEEEAddress)
	}
	if got.ShortAddress != dev.ShortAddress {
		t.Errorf("short = 0x%04X, want 0x%04X", got.ShortAddress, dev.ShortAddress)
	}
	if got.Manufacturer != dev.Manufacturer {
		t.Errorf("manufacturer = %q, want %q", got.Manufacturer, dev.Manufacturer)
	}
	if got.Model != dev.Model {
		t.Errorf("model = %q, want %q", got.Model, dev.Model)
	}
	if !got.Interviewed {
		t.Error("interviewed = false, want true")
	}
	if len(got.Endpoints) != 1 {
		t.Fatalf("endpoints = %d, want 1", len(got.Endpoints))
	}
	if got.Endpoints[0].ID != 1 {
		t.Errorf("ep id = %d, want 1", got.Endpoints[0].ID)
	}
}

func TestDeleteDevice(t *testing.T) {
	s := newTestStore(t)

	dev := &Device{IEEEAddress: "00158D00012A3B4C", ShortAddress: 0x1234}
	if err := s.SaveDevice(dev); err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteDevice(dev.IEEEAddress); err != nil {
		t.Fatal(err)
	}

	_, err := s.GetDevice(dev.IEEEAddress)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestListDevices(t *testing.T) {
	s := newTestStore(t)

	devs := []*Device{
		{IEEEAddress: "0000000000000001", ShortAddress: 0x0001},
		{IEEEAddress: "0000000000000002", ShortAddress: 0x0002},
		{IEEEAddress: "0000000000000003", ShortAddress: 0x0003},
	}
	for _, d := range devs {
		if err := s.SaveDevice(d); err != nil {
			t.Fatal(err)
		}
	}

	list, err := s.ListDevices()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("list count = %d, want 3", len(list))
	}

	// Verify all devices present.
	found := make(map[string]bool)
	for _, d := range list {
		found[d.IEEEAddress] = true
	}
	for _, d := range devs {
		if !found[d.IEEEAddress] {
			t.Errorf("device %s not in list", d.IEEEAddress)
		}
	}
}

func TestGetDeviceNotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.GetDevice("FFFFFFFFFFFFFFFF")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSaveAndGetNetworkState(t *testing.T) {
	s := newTestStore(t)

	state := &NetworkState{
		Channel:    15,
		PanID:      0x1A62,
		ExtPanID:   "DDDDDDDDDDDDDDDD",
		NetworkKey: "aabbccddeeff0011",
		Formed:     true,
	}

	if err := s.SaveNetworkState(state); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetNetworkState()
	if err != nil {
		t.Fatal(err)
	}

	if got.Channel != state.Channel {
		t.Errorf("channel = %d, want %d", got.Channel, state.Channel)
	}
	if got.PanID != state.PanID {
		t.Errorf("pan_id = 0x%04X, want 0x%04X", got.PanID, state.PanID)
	}
	if got.ExtPanID != state.ExtPanID {
		t.Errorf("ext_pan_id = %q, want %q", got.ExtPanID, state.ExtPanID)
	}
	if got.NetworkKey != state.NetworkKey {
		t.Errorf("network_key = %q, want %q", got.NetworkKey, state.NetworkKey)
	}
	if !got.Formed {
		t.Error("formed = false, want true")
	}
}
