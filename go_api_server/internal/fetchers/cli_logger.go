package fetchers

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

// queryCounter is a global atomic counter for sequential query IDs.
var queryCounter atomic.Int64

// queryCounterMu protects reset operations.
var queryCounterMu sync.Mutex

// nextQueryID returns the next sequential query ID like Q-001, Q-002, etc.
func nextQueryID() string {
	n := queryCounter.Add(1)
	return fmt.Sprintf("Q-%03d", n)
}

// ResetQueryCounter resets the counter (useful between refresh cycles).
func ResetQueryCounter() {
	queryCounterMu.Lock()
	defer queryCounterMu.Unlock()
	queryCounter.Store(0)
}

// LogAPICall logs a structured API call entry matching the Python cli_logger format.
// It assigns a sequential query ID and logs the CLI command with status.
func LogAPICall(cliCommand string, status string) string {
	queryID := nextQueryID()

	if status == "error" {
		slog.Error("aws_api_call",
			"log_type", "api_call",
			"query_id", queryID,
			"cli_command", cliCommand,
			"status", status,
			"component", "fetcher",
		)
	} else {
		slog.Info("aws_api_call",
			"log_type", "api_call",
			"query_id", queryID,
			"cli_command", cliCommand,
			"status", status,
			"component", "fetcher",
		)
	}

	return queryID
}

// LogAPICallError logs a structured API call entry with an error message.
func LogAPICallError(cliCommand string, errMsg string) string {
	queryID := nextQueryID()

	slog.Error("aws_api_call",
		"log_type", "api_call",
		"query_id", queryID,
		"cli_command", cliCommand,
		"status", "error",
		"error", errMsg,
		"component", "fetcher",
	)

	return queryID
}
