// Package googlecloudlogging is an implementation for google cloudlog.
package googlecloudlogging

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/toyo/gcp/gce"
)

func init() {
	setLog(slog.LevelDebug)
}

func setLog(loglevel slog.Leveler) {
	slog.SetDefault(
		slog.New(handlerWithSpanContext(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				ReplaceAttr: replacer,
				Level:       loglevel,
				AddSource:   true}))))
}

type contextKey string

const tokenContextSaver contextKey = "AppEngine2ndGenerationLogger-Saver"

type contextSaver struct {
	trace        string
	spanID       string
	traceSampled *bool
	requestID    string
}

func ContextInit(ctx context.Context, r *http.Request) context.Context {
	if gce.GetProjectID() != "" || true {
		var trace, span, requestID string
		var sampled *bool

		traceHeader := r.Header.Get("X-Cloud-Trace-Context")
		traceParts := strings.Split(traceHeader, "/")
		if len(traceParts) > 0 && len(traceParts[0]) > 0 {
			trace = "projects/" + gce.GetProjectID() + "/traces/" + traceParts[0]
		}

		spanParts := strings.Split(traceParts[1], `;`)
		if len(spanParts) > 0 && len(spanParts[0]) > 0 {
			span = spanParts[0]
			if len(spanParts) > 1 && len(spanParts[1]) == 3 {
				sampled = new(bool)
				*sampled = spanParts[1][2] == 1
			}
		}

		requestID = r.Header.Get("X-Request-ID")
		if requestID == "" && len(traceParts) > 0 {
			requestID = traceParts[0]
		}

		ctx = context.WithValue(ctx, tokenContextSaver, contextSaver{
			trace: trace, spanID: span, traceSampled: sampled, requestID: requestID,
		})
	}
	return ctx
}

// This is from https://docs.cloud.google.com/stackdriver/docs/instrumentation/setup/go?hl=ja#config-structured-logging

func handlerWithSpanContext(handler slog.Handler) *spanContextLogHandler {
	return &spanContextLogHandler{Handler: handler}
}

// spanContextLogHandler is a slog.Handler which adds attributes from the
// span context.
type spanContextLogHandler struct {
	slog.Handler
}

// Handle overrides slog.Handler's Handle method. This adds attributes from the
// span context to the slog.Record.
func (t *spanContextLogHandler) Handle(ctx context.Context, record slog.Record) error {

	if cs, ok := ctx.Value(tokenContextSaver).(contextSaver); ok {
		// Add trace context attributes following Cloud Logging structured log format described
		// in https://cloud.google.com/logging/docs/structured-logging#special-payload-fields
		if cs.trace != "" {
			record.AddAttrs(
				slog.String("logging.googleapis.com/trace", cs.trace),
			)
		}

		if cs.requestID != "" {
			record.AddAttrs(
				slog.String("requestId", cs.requestID),
			)
		}

		if cs.spanID != "" {
			record.AddAttrs(
				slog.Any("logging.googleapis.com/spanId", cs.spanID),
			)
		}

		if cs.traceSampled != nil {
			record.AddAttrs(
				slog.Bool("logging.googleapis.com/trace_sampled", *cs.traceSampled),
			)
		}

		if cs.requestID != "" {
			record.AddAttrs(slog.String("requestId", cs.requestID))
		}
	}

	return t.Handler.Handle(ctx, record)
}

func replacer(groups []string, a slog.Attr) slog.Attr {
	// Rename attribute keys to match Cloud Logging structured log format

	_ = groups // is unused
	switch a.Key {
	case slog.LevelKey:
		a.Key = "severity"
		// Map slog.Level string values to Cloud Logging LogSeverity
		// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
		if level := a.Value.Any().(slog.Level); level == slog.LevelWarn {
			a.Value = slog.StringValue("WARNING")
		}
	case slog.TimeKey:
		a.Key = "timestamp"
	case slog.MessageKey:
		a.Key = "message"
	case slog.SourceKey:
		a.Key = "logging.googleapis.com/sourceLocation"
	}
	return a
}
