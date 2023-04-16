package log

import (
	"encoding/json"
	"log"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
)

// Entry defines a log Entry.
// https://cloud.google.com/logging/docs/agent/configuration?hl=ja#special-fields
type Entry struct {
	Severity     string     `json:"severity,omitempty"`
	Message      string     `json:"message"`
	Trace        string     `json:"logging.googleapis.com/trace,omitempty"`
	SpanID       string     `json:"logging.googleapis.com/spanId,omitempty"`
	TraceSampled *bool      `json:"logging.googleapis.com/trace_sampled"`
	TimeStamp    *time.Time `json:"time,omitempty"`

	// Optional. Source code location information associated with the log entry,
	// if any.
	SourceLocation *loggingpb.LogEntrySourceLocation `json:"logging.googleapis.com/sourceLocation"`
}

// String renders an entry structure to the JSON format expected by Cloud Logging.
func (e Entry) String() string {
	if e.Severity == "" {
		e.Severity = "INFO"
	}
	out, err := json.Marshal(e)
	if err != nil {
		log.Printf("json.Marshal: %v", err)
	}
	return string(out)
}
