package store

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketDevices = []byte("devices")
	bucketNetwork = []byte("network")
	keyNetState   = []byte("state")
)

// BoltStore implements Store using BoltDB.
type BoltStore struct {
	db *bolt.DB
}

// NewBoltStore opens or creates a BoltDB database.
func NewBoltStore(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}

	// Create buckets
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketDevices, bucketNetwork} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create buckets: %w", err)
	}

	return &BoltStore{db: db}, nil
}

func (s *BoltStore) SaveDevice(dev *Device) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDevices)
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucketDevices)
		}
		data, err := json.Marshal(dev)
		if err != nil {
			return err
		}
		return b.Put([]byte(dev.IEEEAddress), data)
	})
}

func (s *BoltStore) GetDevice(ieee string) (*Device, error) {
	var dev Device
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDevices)
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucketDevices)
		}
		data := b.Get([]byte(ieee))
		if data == nil {
			return fmt.Errorf("device %s: %w", ieee, ErrNotFound)
		}
		return json.Unmarshal(data, &dev)
	})
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

func (s *BoltStore) DeleteDevice(ieee string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDevices)
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucketDevices)
		}
		return b.Delete([]byte(ieee))
	})
}

func (s *BoltStore) ListDevices() ([]*Device, error) {
	var devices []*Device
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDevices)
		if b == nil {
			return nil // no bucket = no devices
		}
		devices = make([]*Device, 0, b.Stats().KeyN)
		return b.ForEach(func(k, v []byte) error {
			var dev Device
			if err := json.Unmarshal(v, &dev); err != nil {
				return err
			}
			devices = append(devices, &dev)
			return nil
		})
	})
	return devices, err
}

func (s *BoltStore) SaveNetworkState(state *NetworkState) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNetwork)
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucketNetwork)
		}
		// Use internal storage struct to persist the network key.
		st := networkStateStorage{
			Channel:    state.Channel,
			PanID:      state.PanID,
			ExtPanID:   state.ExtPanID,
			NetworkKey: state.NetworkKey,
			Formed:     state.Formed,
		}
		data, err := json.Marshal(st)
		if err != nil {
			return err
		}
		return b.Put(keyNetState, data)
	})
}

func (s *BoltStore) GetNetworkState() (*NetworkState, error) {
	var state NetworkState
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNetwork)
		if b == nil {
			return fmt.Errorf("bucket %q not found", bucketNetwork)
		}
		data := b.Get(keyNetState)
		if data == nil {
			return fmt.Errorf("network state: %w", ErrNotFound)
		}
		// Deserialize via internal storage struct to recover the network key.
		var st networkStateStorage
		if err := json.Unmarshal(data, &st); err != nil {
			return err
		}
		state = NetworkState{
			Channel:    st.Channel,
			PanID:      st.PanID,
			ExtPanID:   st.ExtPanID,
			NetworkKey: st.NetworkKey,
			Formed:     st.Formed,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (s *BoltStore) Close() error {
	return s.db.Close()
}
