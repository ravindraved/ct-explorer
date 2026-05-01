package handlers

import (
	"log/slog"
	"net/http"

	"nhooyr.io/websocket"

	"go_api_server/internal/logging"
)

// WSLogs godoc
// @Summary      WebSocket log stream
// @Description  Accepts WebSocket upgrade and streams log entries in real-time
// @Tags         logs
// @Router       /ws/logs [get]
func WSLogs(broadcaster *logging.Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // Allow any origin for dev
		})
		if err != nil {
			slog.Error("ws_accept_failed", "error", err.Error())
			return
		}
		defer conn.CloseNow()

		broadcaster.Add(conn)
		defer broadcaster.Remove(conn)

		slog.Info("ws_client_connected", "component", "websocket")

		// Replay buffered log entries
		broadcaster.ReplayBuffer(conn)

		// Keep connection alive — wait for client to disconnect
		for {
			_, _, err := conn.Read(r.Context())
			if err != nil {
				break
			}
		}

		slog.Info("ws_client_disconnected", "component", "websocket")
	}
}
