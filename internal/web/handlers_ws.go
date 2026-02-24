package web

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

// WSHub manages WebSocket connections and broadcasts events.
type WSHub struct {
	clients map[*wsClient]struct{}
	mu      sync.RWMutex
	logger  *slog.Logger

	register   chan *wsClient
	unregister chan *wsClient
	broadcast  chan interface{}

	done     chan struct{}
	stopOnce sync.Once
}

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

// NewWSHub creates a new WebSocket hub.
func NewWSHub(logger *slog.Logger) *WSHub {
	return &WSHub{
		clients:    make(map[*wsClient]struct{}),
		logger:     logger,
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient),
		broadcast:  make(chan interface{}, 256),
		done:       make(chan struct{}),
	}
}

// Run starts the hub event loop.
func (h *WSHub) Run() {
	for {
		select {
		case <-h.done:
			// Close all remaining clients on shutdown
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = struct{}{}
			total := len(h.clients)
			h.mu.Unlock()
			h.logger.Debug("ws client connected", "total", total)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			total := len(h.clients)
			h.mu.Unlock()
			h.logger.Debug("ws client disconnected", "total", total)

		case msg := <-h.broadcast:
			data, err := json.Marshal(msg)
			if err != nil {
				h.logger.Error("ws marshal", "err", err)
				continue
			}
			h.mu.Lock()
			var slow []*wsClient
			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					// Client too slow, mark for eviction
					slow = append(slow, client)
				}
			}
			for _, client := range slow {
				delete(h.clients, client)
				close(client.send)
				h.logger.Warn("ws client evicted (too slow)")
			}
			h.mu.Unlock()
		}
	}
}

// Stop signals the hub to shut down. Safe to call multiple times.
func (h *WSHub) Stop() {
	h.stopOnce.Do(func() {
		close(h.done)
	})
}

// Broadcast sends a message to all connected clients.
func (h *WSHub) Broadcast(msg interface{}) {
	select {
	case h.broadcast <- msg:
	default:
		h.logger.Warn("ws broadcast channel full, dropping message")
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	opts := &websocket.AcceptOptions{}
	if len(s.allowedOrigins) > 0 {
		opts.OriginPatterns = s.allowedOrigins
	}
	// If no allowedOrigins configured, nhooyr defaults to same-origin check.

	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		s.logger.Error("ws accept", "err", err)
		return
	}

	conn.SetReadLimit(4096)

	client := &wsClient{
		conn: conn,
		send: make(chan []byte, 64),
	}

	select {
	case s.wsHub.register <- client:
	case <-s.wsHub.done:
		conn.Close(websocket.StatusGoingAway, "server shutdown")
		return
	}

	go s.wsWritePump(client)
	s.wsReadPump(client)
}

func (s *Server) wsWritePump(client *wsClient) {
	for msg := range client.send {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := client.conn.Write(ctx, websocket.MessageText, msg)
		cancel()
		if err != nil {
			return
		}
	}
	// Channel closed by hub; close connection.
	client.conn.Close(websocket.StatusNormalClosure, "")
}

func (s *Server) wsReadPump(client *wsClient) {
	defer func() {
		select {
		case s.wsHub.unregister <- client:
		case <-s.wsHub.done:
			// Hub already shut down; close connection directly.
			client.conn.Close(websocket.StatusGoingAway, "server shutdown")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cancel read context when hub shuts down.
	go func() {
		select {
		case <-s.wsHub.done:
			cancel()
		case <-ctx.Done():
		}
	}()

	for {
		_, _, err := client.conn.Read(ctx)
		if err != nil {
			return
		}
		// We don't process incoming messages from clients for now
	}
}
