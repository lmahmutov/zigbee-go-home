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

	// Network state
	SaveNetworkState(state *NetworkState) error
	GetNetworkState() (*NetworkState, error)

	// Close the store
	Close() error
}
