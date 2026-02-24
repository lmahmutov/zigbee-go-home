package zcl

import (
	"log/slog"
	"os"
	"testing"
)

func TestRegistryRegisterAndGet(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	r := NewRegistry(logger)

	c := ClusterDef{
		ID:   0x0006,
		Name: "On/Off",
		Attributes: []AttributeDef{
			{ID: 0, Name: "OnOff", Type: TypeBool, Access: AccessRead},
		},
	}
	r.Register(c)

	got := r.Get(0x0006)
	if got == nil {
		t.Fatal("cluster not found")
	}
	if got.Name != "On/Off" {
		t.Errorf("name = %q, want %q", got.Name, "On/Off")
	}
	if len(got.Attributes) != 1 {
		t.Errorf("attrs = %d, want 1", len(got.Attributes))
	}
}

func TestRegistryMerge(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	r := NewRegistry(logger)

	// Register base cluster
	r.Register(ClusterDef{
		ID:   0x0006,
		Name: "On/Off",
		Attributes: []AttributeDef{
			{ID: 0, Name: "OnOff", Type: TypeBool, Access: AccessRead},
		},
	})

	// Merge additional attribute
	r.Register(ClusterDef{
		ID: 0x0006,
		Attributes: []AttributeDef{
			{ID: 0x4003, Name: "StartUpOnOff", Type: TypeEnum8, Access: AccessRead | AccessWrite},
		},
	})

	got := r.Get(0x0006)
	if len(got.Attributes) != 2 {
		t.Errorf("after merge: attrs = %d, want 2", len(got.Attributes))
	}

	attr := got.FindAttribute(0x4003)
	if attr == nil {
		t.Fatal("merged attribute not found")
	}
	if attr.Name != "StartUpOnOff" {
		t.Errorf("name = %q, want StartUpOnOff", attr.Name)
	}
}

func TestRegistryAll(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	r := NewRegistry(logger)

	r.Register(ClusterDef{ID: 1, Name: "A"})
	r.Register(ClusterDef{ID: 2, Name: "B"})
	r.Register(ClusterDef{ID: 3, Name: "C"})

	all := r.All()
	if len(all) != 3 {
		t.Errorf("got %d clusters, want 3", len(all))
	}
}

