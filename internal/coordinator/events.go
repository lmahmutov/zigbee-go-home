package coordinator

import (
	"log/slog"
	"sync"
)

// Event types
const (
	EventDeviceJoined    = "device_joined"
	EventDeviceLeft      = "device_left"
	EventDeviceAnnounce  = "device_announce"
	EventAttributeReport  = "attribute_report"
	EventClusterCommand   = "cluster_command"
	EventPropertyUpdate   = "property_update"
	EventNetworkState    = "network_state"
	EventPermitJoin      = "permit_join"
)

// Event represents a coordinator event.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// EventHandler is a callback for events.
type EventHandler func(Event)

// EventBus provides pub/sub for coordinator events.
type EventBus struct {
	mu          sync.RWMutex
	handlers    map[string]map[uint64]EventHandler
	allHandlers map[uint64]EventHandler
	nextID      uint64
	logger      *slog.Logger
}

// NewEventBus creates a new event bus.
func NewEventBus(logger *slog.Logger) *EventBus {
	return &EventBus{
		handlers:    make(map[string]map[uint64]EventHandler),
		allHandlers: make(map[uint64]EventHandler),
		logger:      logger,
	}
}

// On registers a handler for a specific event type.
// Returns an unsubscribe function.
func (eb *EventBus) On(eventType string, handler EventHandler) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	id := eb.nextID
	eb.nextID++
	if eb.handlers[eventType] == nil {
		eb.handlers[eventType] = make(map[uint64]EventHandler)
	}
	eb.handlers[eventType][id] = handler
	return func() {
		eb.mu.Lock()
		defer eb.mu.Unlock()
		delete(eb.handlers[eventType], id)
	}
}

// OnAll registers a handler that receives all events.
// Returns an unsubscribe function.
func (eb *EventBus) OnAll(handler EventHandler) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	id := eb.nextID
	eb.nextID++
	eb.allHandlers[id] = handler
	return func() {
		eb.mu.Lock()
		defer eb.mu.Unlock()
		delete(eb.allHandlers, id)
	}
}

// Emit sends an event to all matching handlers.
// Handlers are called synchronously; a panicking handler is recovered.
func (eb *EventBus) Emit(event Event) {
	eb.mu.RLock()
	handlers := make([]EventHandler, 0, len(eb.handlers[event.Type])+len(eb.allHandlers))
	for _, h := range eb.handlers[event.Type] {
		handlers = append(handlers, h)
	}
	for _, h := range eb.allHandlers {
		handlers = append(handlers, h)
	}
	eb.mu.RUnlock()

	for _, h := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					eb.logger.Error("event handler panic", "type", event.Type, "panic", r)
				}
			}()
			h(event)
		}()
	}
}
