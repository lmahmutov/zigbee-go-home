package web

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func newTestHub() *WSHub {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return NewWSHub(logger)
}

func TestWSHubRegisterUnregister(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Stop()

	client := &wsClient{send: make(chan []byte, 16)}
	hub.register <- client

	// Give hub time to process
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	count := len(hub.clients)
	hub.mu.RUnlock()
	if count != 1 {
		t.Errorf("after register: count = %d, want 1", count)
	}

	hub.unregister <- client

	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	count = len(hub.clients)
	hub.mu.RUnlock()
	if count != 0 {
		t.Errorf("after unregister: count = %d, want 0", count)
	}
}

func TestWSHubBroadcast(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Stop()

	c1 := &wsClient{send: make(chan []byte, 16)}
	c2 := &wsClient{send: make(chan []byte, 16)}

	hub.register <- c1
	hub.register <- c2
	time.Sleep(10 * time.Millisecond)

	hub.Broadcast(map[string]string{"type": "test"})
	time.Sleep(10 * time.Millisecond)

	select {
	case msg := <-c1.send:
		if len(msg) == 0 {
			t.Error("c1 received empty message")
		}
	default:
		t.Error("c1 did not receive broadcast")
	}

	select {
	case msg := <-c2.send:
		if len(msg) == 0 {
			t.Error("c2 received empty message")
		}
	default:
		t.Error("c2 did not receive broadcast")
	}
}

func TestWSHubSlowClientEviction(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Stop()

	// Create a client with a tiny buffer that will fill up
	slow := &wsClient{send: make(chan []byte, 1)}
	fast := &wsClient{send: make(chan []byte, 64)}

	hub.register <- slow
	hub.register <- fast
	time.Sleep(10 * time.Millisecond)

	// Fill slow client's buffer
	hub.Broadcast("msg1")
	time.Sleep(10 * time.Millisecond)

	// Second message should evict the slow client (buffer full, can't receive)
	hub.Broadcast("msg2")
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	_, slowPresent := hub.clients[slow]
	_, fastPresent := hub.clients[fast]
	hub.mu.RUnlock()

	if slowPresent {
		t.Error("slow client should have been evicted")
	}
	if !fastPresent {
		t.Error("fast client should still be present")
	}
}

func TestWSHubBroadcastDropsWhenFull(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Stop()

	// Fill the broadcast channel
	for i := 0; i < 256; i++ {
		hub.Broadcast(i)
	}

	// This should not block; it should drop
	done := make(chan struct{})
	go func() {
		hub.Broadcast("overflow")
		close(done)
	}()

	select {
	case <-done:
		// Good, didn't block
	case <-time.After(1 * time.Second):
		t.Error("Broadcast blocked when channel is full")
	}
}

func TestWSHubStopIdempotent(t *testing.T) {
	hub := newTestHub()
	go hub.Run()

	// First stop
	hub.Stop()

	// Second stop should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("second Stop() panicked: %v", r)
		}
	}()
	hub.Stop()
}

func TestWSHubStopClosesClients(t *testing.T) {
	hub := newTestHub()
	go hub.Run()

	client := &wsClient{send: make(chan []byte, 16)}
	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	hub.Stop()
	time.Sleep(10 * time.Millisecond)

	// send channel should be closed
	_, ok := <-client.send
	if ok {
		t.Error("client.send should be closed after hub stop")
	}
}

func TestWSHubUnregisterNonExistentClient(t *testing.T) {
	hub := newTestHub()
	go hub.Run()
	defer hub.Stop()

	// Unregistering a client that was never registered should not panic
	unknown := &wsClient{send: make(chan []byte, 16)}
	hub.unregister <- unknown
	time.Sleep(10 * time.Millisecond)

	// Channel should NOT be closed since client was never registered
	select {
	case unknown.send <- []byte("test"):
		// Good, channel still open
	default:
		t.Error("channel should still be open for non-registered client")
	}
}
