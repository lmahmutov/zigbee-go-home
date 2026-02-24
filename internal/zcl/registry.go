package zcl

import (
	"fmt"
	"log/slog"
	"sync"
)

// Registry holds all known ZCL cluster definitions.
type Registry struct {
	mu       sync.RWMutex
	clusters map[uint16]*ClusterDef
	logger   *slog.Logger
}

// NewRegistry creates an empty registry.
func NewRegistry(logger *slog.Logger) *Registry {
	return &Registry{
		clusters: make(map[uint16]*ClusterDef),
		logger:   logger,
	}
}

// Register adds a cluster definition to the registry.
func (r *Registry) Register(c ClusterDef) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.clusters[c.ID]; ok {
		existing.Merge(&c)
		r.logger.Debug("cluster merged", "id", fmt.Sprintf("0x%04X", c.ID), "name", existing.Name)
	} else {
		clone := c
		r.clusters[c.ID] = &clone
		r.logger.Debug("cluster registered", "id", fmt.Sprintf("0x%04X", c.ID), "name", c.Name)
	}
}

// Get returns a cluster definition by ID, or nil if not found.
// The returned value is a deep copy; callers may modify it safely.
func (r *Registry) Get(id uint16) *ClusterDef {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c := r.clusters[id]
	if c == nil {
		return nil
	}
	return c.DeepCopy()
}

// All returns all registered cluster definitions.
// Each entry is a deep copy; callers may modify them safely.
func (r *Registry) All() []ClusterDef {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ClusterDef, 0, len(r.clusters))
	for _, c := range r.clusters {
		result = append(result, *c.DeepCopy())
	}
	return result
}

