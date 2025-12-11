package v1

import "time"

type Event struct {
	TraceID    string    `json:"trace_id"`
	StartTime  time.Time `json:"start_time"`
	DurationMs int64     `json:"duration_ms"`
	Status     int       `json:"status_code"`
	TokenCount int       `json:"token_count"`
	Model      string    `json:"model"`
	Response   string    `json:"response,omitempty"`
}
