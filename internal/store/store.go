package store

import "errors"

// ErrNotFound is returned when a requested entity does not exist in the store.
var ErrNotFound = errors.New("not found")

// Store defines the persistence interface.
type Store interface {
	// Device operations
	SaveDevice(dev *Device) error
	GetDevice(ieee string) (*Device, error)
	DeleteDevice(ieee string) error
	ListDevices() ([]*Device, error)

	// UpdateDevice atomically reads, modifies, and saves a device in a single
	// transaction. Returns ErrNotFound if the device does not exist.
	UpdateDevice(ieee string, fn func(dev *Device) error) error

	// Network state
	SaveNetworkState(state *NetworkState) error
	GetNetworkState() (*NetworkState, error)

	// Close the store
	Close() error
}
