package coordinator

import (
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"
)

func TestParseIEEE(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    [8]byte
		wantErr bool
	}{
		{
			"hex string no colons",
			"00124B001234ABCD",
			[8]byte{0x00, 0x12, 0x4B, 0x00, 0x12, 0x34, 0xAB, 0xCD},
			false,
		},
		{
			"hex string with colons",
			"00:12:4B:00:12:34:AB:CD",
			[8]byte{0x00, 0x12, 0x4B, 0x00, 0x12, 0x34, 0xAB, 0xCD},
			false,
		},
		{
			"all zeros",
			"0000000000000000",
			[8]byte{},
			false,
		},
		{
			"all FF",
			"FFFFFFFFFFFFFFFF",
			[8]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			false,
		},
		{
			"too short",
			"00124B",
			[8]byte{},
			true,
		},
		{
			"too long",
			"00124B001234ABCD00",
			[8]byte{},
			true,
		},
		{
			"invalid hex",
			"ZZZZZZZZZZZZZZZZ",
			[8]byte{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIEEE(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIEEE(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseIEEE(%q) = %X, want %X", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseExtPanID(t *testing.T) {
	// ParseExtPanID is an alias for ParseIEEE
	got, err := ParseExtPanID("DD:CC:BB:AA:00:11:22:33")
	if err != nil {
		t.Fatal(err)
	}
	expected := [8]byte{0xDD, 0xCC, 0xBB, 0xAA, 0x00, 0x11, 0x22, 0x33}
	if got != expected {
		t.Errorf("got %X, want %X", got, expected)
	}
}

// --- EventBus tests ---

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestEventBusEmitOn(t *testing.T) {
	eb := NewEventBus(newTestLogger())
	var received Event

	eb.On(EventDeviceJoined, func(e Event) {
		received = e
	})

	eb.Emit(Event{Type: EventDeviceJoined, Data: "test"})

	if received.Type != EventDeviceJoined {
		t.Errorf("type = %q, want %q", received.Type, EventDeviceJoined)
	}
	if received.Data != "test" {
		t.Errorf("data = %v, want %q", received.Data, "test")
	}
}

func TestEventBusOnDoesNotReceiveOtherTypes(t *testing.T) {
	eb := NewEventBus(newTestLogger())
	called := false

	eb.On(EventDeviceJoined, func(e Event) {
		called = true
	})

	eb.Emit(Event{Type: EventDeviceLeft, Data: "test"})

	if called {
		t.Error("handler called for wrong event type")
	}
}

func TestEventBusOnAll(t *testing.T) {
	eb := NewEventBus(newTestLogger())
	var count atomic.Int32

	eb.OnAll(func(e Event) {
		count.Add(1)
	})

	eb.Emit(Event{Type: EventDeviceJoined})
	eb.Emit(Event{Type: EventDeviceLeft})
	eb.Emit(Event{Type: EventAttributeReport})

	if count.Load() != 3 {
		t.Errorf("onAll called %d times, want 3", count.Load())
	}
}

func TestEventBusUnsubscribe(t *testing.T) {
	eb := NewEventBus(newTestLogger())
	var count atomic.Int32

	unsub := eb.On(EventDeviceJoined, func(e Event) {
		count.Add(1)
	})

	eb.Emit(Event{Type: EventDeviceJoined})
	if count.Load() != 1 {
		t.Fatalf("expected 1 call before unsub, got %d", count.Load())
	}

	unsub()
	eb.Emit(Event{Type: EventDeviceJoined})
	if count.Load() != 1 {
		t.Errorf("expected 1 call after unsub, got %d", count.Load())
	}
}

func TestEventBusOnAllUnsubscribe(t *testing.T) {
	eb := NewEventBus(newTestLogger())
	var count atomic.Int32

	unsub := eb.OnAll(func(e Event) {
		count.Add(1)
	})

	eb.Emit(Event{Type: EventDeviceJoined})
	unsub()
	eb.Emit(Event{Type: EventDeviceJoined})

	if count.Load() != 1 {
		t.Errorf("expected 1 call, got %d", count.Load())
	}
}

func TestEventBusPanicRecovery(t *testing.T) {
	eb := NewEventBus(newTestLogger())
	var called atomic.Int32

	// Register two handlers — one panics, one increments counter.
	// Both should be attempted despite the panic.
	eb.On(EventDeviceJoined, func(e Event) {
		called.Add(1)
		panic("test panic")
	})
	eb.On(EventDeviceJoined, func(e Event) {
		called.Add(1)
	})

	// Should not panic
	eb.Emit(Event{Type: EventDeviceJoined})

	// Both handlers should have been called despite one panicking.
	if c := called.Load(); c != 2 {
		t.Errorf("expected 2 handlers called, got %d", c)
	}
}

func TestEventBusConcurrentEmit(t *testing.T) {
	eb := NewEventBus(newTestLogger())
	var count atomic.Int32

	eb.OnAll(func(e Event) {
		count.Add(1)
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			eb.Emit(Event{Type: EventAttributeReport})
		}()
	}
	wg.Wait()

	if count.Load() != 100 {
		t.Errorf("got %d, want 100", count.Load())
	}
}

func TestEventBusMultipleHandlersSameType(t *testing.T) {
	eb := NewEventBus(newTestLogger())
	var count atomic.Int32

	eb.On(EventDeviceJoined, func(e Event) { count.Add(1) })
	eb.On(EventDeviceJoined, func(e Event) { count.Add(1) })
	eb.On(EventDeviceJoined, func(e Event) { count.Add(1) })

	eb.Emit(Event{Type: EventDeviceJoined})

	if count.Load() != 3 {
		t.Errorf("got %d, want 3", count.Load())
	}
}
