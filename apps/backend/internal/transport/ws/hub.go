// Package ws implements the WebSocket fan-out hub used by the dashboard.
package ws

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/contrib/websocket"
)

// Hub broadcasts events to all connected clients.
//
// Backpressure rule: each connection has a bounded buffered channel. If a
// client cannot keep up, the hub drops events for that client (and tags it
// as slow). The publisher never blocks.
type Hub struct {
	mu        sync.RWMutex
	conns     map[*conn]struct{}
	connSeq   atomic.Uint64
	bufSize   int
	heartbeat time.Duration
}

type conn struct {
	id     uint64
	c      *websocket.Conn
	ch     chan []byte
	closed atomic.Bool
}

func NewHub(bufSize int, heartbeat time.Duration) *Hub {
	if bufSize <= 0 {
		bufSize = 64
	}
	if heartbeat <= 0 {
		heartbeat = 20 * time.Second
	}
	return &Hub{
		conns:     map[*conn]struct{}{},
		bufSize:   bufSize,
		heartbeat: heartbeat,
	}
}

// Publish implements service.Publisher. JSON-encodes the payload once and
// fans out to all live connections; per-conn send is non-blocking.
func (h *Hub) Publish(event any) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.conns {
		if c.closed.Load() {
			continue
		}
		select {
		case c.ch <- data:
		default:
			// Slow consumer; drop and tag.
			c.closed.Store(true)
		}
	}
}

// ConnCount returns the number of live connections.
func (h *Hub) ConnCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns)
}

// Add registers a connection with the hub and returns a close function.
func (h *Hub) Add(c *websocket.Conn) func() {
	id := h.connSeq.Add(1)
	k := &conn{id: id, c: c, ch: make(chan []byte, h.bufSize)}
	h.mu.Lock()
	h.conns[k] = struct{}{}
	h.mu.Unlock()

	go h.writer(k)
	go h.heartbeatLoop(k)

	return func() {
		if !k.closed.CompareAndSwap(false, true) {
			return
		}
		h.mu.Lock()
		delete(h.conns, k)
		h.mu.Unlock()
		_ = c.Close()
	}
}

func (h *Hub) writer(c *conn) {
	for data := range c.ch {
		if err := c.c.WriteMessage(websocket.TextMessage, data); err != nil {
			c.closed.Store(true)
			return
		}
	}
}

func (h *Hub) heartbeatLoop(c *conn) {
	t := time.NewTicker(h.heartbeat)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if c.closed.Load() {
				return
			}
			payload := []byte(`{"type":"agent.heartbeat"}`)
			select {
			case c.ch <- payload:
			default:
				c.closed.Store(true)
				return
			}
		}
	}
}
