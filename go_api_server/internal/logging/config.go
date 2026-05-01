package logging

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"
)

// Setup creates a slog.Logger that writes JSON to stdout and also
// broadcasts entries to the WebSocket broadcaster.
func Setup(level slog.Level, broadcaster *Broadcaster) *slog.Logger {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	teeHandler := NewTeeHandler(jsonHandler, broadcaster)
	logger := slog.New(teeHandler)
	slog.SetDefault(logger)
	return logger
}

// TeeHandler wraps a slog.Handler and also sends log entries to the broadcaster.
type TeeHandler struct {
	inner       slog.Handler
	broadcaster *Broadcaster
	attrs       []slog.Attr
	groups      []string
}

// NewTeeHandler creates a TeeHandler wrapping the given handler and broadcaster.
func NewTeeHandler(inner slog.Handler, broadcaster *Broadcaster) *TeeHandler {
	return &TeeHandler{
		inner:       inner,
		broadcaster: broadcaster,
		attrs:       make([]slog.Attr, 0),
		groups:      make([]string, 0),
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle writes the record to the inner handler and broadcasts it.
func (h *TeeHandler) Handle(ctx context.Context, r slog.Record) error {
	// Write to inner handler (stdout)
	err := h.inner.Handle(ctx, r)

	// Also broadcast to WebSocket clients
	if h.broadcaster != nil {
		entry := h.formatRecord(r)
		h.broadcaster.Broadcast(entry)
	}

	return err
}

// WithAttrs returns a new TeeHandler with the given attributes.
func (h *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TeeHandler{
		inner:       h.inner.WithAttrs(attrs),
		broadcaster: h.broadcaster,
		attrs:       append(h.attrs, attrs...),
		groups:      h.groups,
	}
}

// WithGroup returns a new TeeHandler with the given group name.
func (h *TeeHandler) WithGroup(name string) slog.Handler {
	return &TeeHandler{
		inner:       h.inner.WithGroup(name),
		broadcaster: h.broadcaster,
		attrs:       h.attrs,
		groups:      append(h.groups, name),
	}
}

// formatRecord converts a slog.Record to a JSON byte slice for broadcasting.
func (h *TeeHandler) formatRecord(r slog.Record) []byte {
	entry := make(map[string]any)
	entry["time"] = r.Time.Format(time.RFC3339Nano)
	entry["level"] = r.Level.String()
	entry["msg"] = r.Message

	// Add handler-level attrs (e.g., component)
	for _, a := range h.attrs {
		entry[a.Key] = a.Value.Any()
	}

	// Add record-level attrs
	r.Attrs(func(a slog.Attr) bool {
		entry[a.Key] = a.Value.Any()
		return true
	})

	data, _ := json.Marshal(entry)
	return data
}
