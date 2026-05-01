// Package logging provides structured logging with WebSocket broadcast support.
package logging

import (
	"context"
	"regexp"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

const ringBufferCapacity = 100

// credentialPattern matches AWS credential-like strings.
var credentialPattern = regexp.MustCompile(
	`(?i)(AKIA[0-9A-Z]{16}|(?:aws_)?secret_access_key|session_token|x-amz-security-token)`,
)

// Broadcaster manages WebSocket clients and broadcasts log entries.
type Broadcaster struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
	buffer  [][]byte
	bufIdx  int
	bufFull bool
}

// NewBroadcaster creates a new Broadcaster with an empty ring buffer.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[*websocket.Conn]struct{}),
		buffer:  make([][]byte, ringBufferCapacity),
	}
}

// Add registers a WebSocket connection for broadcast.
func (b *Broadcaster) Add(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[conn] = struct{}{}
}

// Remove unregisters a WebSocket connection.
func (b *Broadcaster) Remove(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, conn)
}

// Broadcast scrubs credentials from entry and sends it to all connected clients.
// Failed clients are removed automatically.
func (b *Broadcaster) Broadcast(entry []byte) {
	scrubbed := scrubCredentials(entry)

	b.mu.Lock()
	// Store in ring buffer
	b.buffer[b.bufIdx] = make([]byte, len(scrubbed))
	copy(b.buffer[b.bufIdx], scrubbed)
	b.bufIdx++
	if b.bufIdx >= ringBufferCapacity {
		b.bufIdx = 0
		b.bufFull = true
	}

	// Snapshot clients
	clients := make([]*websocket.Conn, 0, len(b.clients))
	for c := range b.clients {
		clients = append(clients, c)
	}
	b.mu.Unlock()

	// Send to all clients outside the lock
	var failed []*websocket.Conn
	for _, conn := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := conn.Write(ctx, websocket.MessageText, scrubbed)
		cancel()
		if err != nil {
			failed = append(failed, conn)
		}
	}

	// Remove failed clients
	if len(failed) > 0 {
		b.mu.Lock()
		for _, conn := range failed {
			delete(b.clients, conn)
		}
		b.mu.Unlock()
	}
}

// ReplayBuffer sends all buffered entries to a single connection.
func (b *Broadcaster) ReplayBuffer(conn *websocket.Conn) {
	b.mu.Lock()
	entries := b.bufferedEntries()
	b.mu.Unlock()

	for _, entry := range entries {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := conn.Write(ctx, websocket.MessageText, entry)
		cancel()
		if err != nil {
			return
		}
	}
}

// bufferedEntries returns the ring buffer contents in order. Must be called with mu held.
func (b *Broadcaster) bufferedEntries() [][]byte {
	entries := make([][]byte, 0)
	if b.bufFull {
		// Read from bufIdx to end, then 0 to bufIdx
		for i := b.bufIdx; i < ringBufferCapacity; i++ {
			if b.buffer[i] != nil {
				cp := make([]byte, len(b.buffer[i]))
				copy(cp, b.buffer[i])
				entries = append(entries, cp)
			}
		}
		for i := 0; i < b.bufIdx; i++ {
			if b.buffer[i] != nil {
				cp := make([]byte, len(b.buffer[i]))
				copy(cp, b.buffer[i])
				entries = append(entries, cp)
			}
		}
	} else {
		for i := 0; i < b.bufIdx; i++ {
			if b.buffer[i] != nil {
				cp := make([]byte, len(b.buffer[i]))
				copy(cp, b.buffer[i])
				entries = append(entries, cp)
			}
		}
	}
	return entries
}

// scrubCredentials replaces credential-like patterns with [REDACTED].
func scrubCredentials(data []byte) []byte {
	return credentialPattern.ReplaceAll(data, []byte("[REDACTED]"))
}
