package v1

import (
	"encoding/json"
	"time"
)

type Event struct {
	Id         string          `json:"id,omitempty" db:"id"`
	TraceId    string          `json:"trace_id" db:"trace_id"`
	StartTime  time.Time       `json:"start_time" db:"start_time"`
	EndTime    time.Time       `json:"end_time" db:"end_time"`
	DurationMs int64           `json:"duration_ms" db:"duration_ms"`
	StatusCode int             `json:"status_code" db:"status"`
	TokenCount int             `json:"token_count" db:"token_count"`
	Model      string          `json:"model" db:"model"`
	Request    json.RawMessage `json:"request,omitempty" db:"request"`
	Response   string          `json:"response,omitempty" db:"response"`
}
