package log

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/toyo/gcp/gce"
	"go.opencensus.io/trace"
)

type contextKey string

const tokenContextSaver contextKey = "AppEngine2ndGenerationLogger-Saver"

// Entry defines a log entry.
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

type contextSaver struct {
	trace        string
	spanID       string
	traceSampled *bool
}

// NewContextFromReq makes context.
func NewContextFromReq(req *http.Request) (ctx context.Context) {
	return AddContextFromReq(req.Context(), req)
}

// AddContextFromReq send HTTP Request log.
func AddContextFromReq(ctx context.Context, req *http.Request) context.Context {

	if gce.GetProjectID() != "" {
		traceHeader := req.Header.Get("X-Cloud-Trace-Context")
		traceParts := strings.Split(traceHeader, "/")
		if len(traceParts) > 0 && len(traceParts[0]) > 0 {
			trace := "projects/" + gce.GetProjectID() + "/traces/" + traceParts[0]
			ctx = context.WithValue(ctx, tokenContextSaver, contextSaver{trace: trace})
		}
	}

	return ctx
}

func ContextFromSpan(ctx context.Context, span *trace.Span) context.Context {
	ctx = ContextFromTraceID(ctx, span.SpanContext().TraceID)
	ctx = ContextFromSpanID(ctx, span.SpanContext().SpanID)
	ctx = ContextFromTraceSampled(ctx, span.SpanContext().IsSampled())

	return ctx
}

// ContextFromTraceID make context.
func ContextFromTraceID(ctx context.Context, traceid trace.TraceID) context.Context {

	if gce.GetProjectID() != "" {
		cs, ok := ctx.Value(tokenContextSaver).(contextSaver)
		if !ok {
			cs = contextSaver{}
		}
		cs.trace = "projects/" + gce.GetProjectID() + "/traces/" + hex.EncodeToString(traceid[:])
		ctx = context.WithValue(ctx, tokenContextSaver, cs)
	}

	return ctx
}

// ContextFromSpanID make context.
func ContextFromSpanID(ctx context.Context, spanid trace.SpanID) context.Context {

	cs, ok := ctx.Value(tokenContextSaver).(contextSaver)
	if !ok {
		cs = contextSaver{}
	}
	cs.spanID = hex.EncodeToString(spanid[:])
	ctx = context.WithValue(ctx, tokenContextSaver, cs)

	return ctx
}

// ContextFromTraceSampled make context.
func ContextFromTraceSampled(ctx context.Context, traceSampled bool) context.Context {

	cs, ok := ctx.Value(tokenContextSaver).(contextSaver)
	if !ok {
		cs = contextSaver{}
	}
	cs.traceSampled = &traceSampled
	ctx = context.WithValue(ctx, tokenContextSaver, cs)

	return ctx
}

//var logName = `StackDriver`

// logged is for logging.
func logged(ctx context.Context, severity logging.Severity, payloadi interface{}) {

	const stdout = false

	var payload string

	switch v := payloadi.(type) {
	case string:
		payload = v
	case fmt.Stringer:
		payload = v.String()
	case fmt.GoStringer:
		payload = v.GoString()
	default:
		var builder strings.Builder
		json.NewEncoder(&builder).Encode(payloadi)
		payload = builder.String()
	}

	//

	if cs, ok := ctx.Value(tokenContextSaver).(contextSaver); ok {

		if !stdout {
			// Create a Client
			client, err := logging.NewClient(ctx, `projects/`+gce.GetProjectID())
			if err != nil {
				panic(err.Error())
			}
			lg := client.Logger(`logger`)

			e := logging.Entry{
				Timestamp:    time.Now(),
				Payload:      payload,
				Severity:     severity,
				Trace:        cs.trace,
				SpanID:       cs.spanID,
				TraceSampled: *cs.traceSampled,
			}

			if pc, file, line, ok := runtime.Caller(2); ok {
				e.SourceLocation = &loggingpb.LogEntrySourceLocation{
					File:     file,
					Line:     int64(line),
					Function: runtime.FuncForPC(pc).Name(),
				}
			}

			lg.Log(e)
		} else {
			if len(payload) < 65536 {

				t := time.Now()

				e := Entry{}
				e.Message = payload
				e.Severity = severity.String()
				e.TimeStamp = &t
				e.Trace = cs.trace
				e.SpanID = cs.spanID
				e.TraceSampled = cs.traceSampled

				if pc, file, line, ok := runtime.Caller(2); ok {
					e.SourceLocation = &loggingpb.LogEntrySourceLocation{
						File:     file,
						Line:     int64(line),
						Function: runtime.FuncForPC(pc).Name(),
					}
				}
				//cs.client.Logger(logName).Log(e)
				err := json.NewEncoder(os.Stdout).Encode(&e)
				if err != nil {
					fmt.Printf("%s %v\n", err.Error(), e)
				}
			} else {
				log.Println(severity.String() + ": " + payload)
			}
		}
	}
}

// Default send Application log.
func Default(ctx context.Context, a interface{}) {
	logged(ctx, logging.Debug, a)
}

// Debug send Application log.
func Debug(ctx context.Context, a interface{}) {
	logged(ctx, logging.Debug, a)
}

// Info send Application log.
func Info(ctx context.Context, a interface{}) {
	logged(ctx, logging.Info, a)
}

// Notice send Application log.
func Notice(ctx context.Context, a interface{}) {
	logged(ctx, logging.Notice, a)
}

// Warning send Application log.
func Warning(ctx context.Context, a interface{}) {
	logged(ctx, logging.Warning, a)
}

// Error send Application log.
func Error(ctx context.Context, a interface{}) {
	logged(ctx, logging.Error, a)
}

// Critical send Application log.
func Critical(ctx context.Context, a interface{}) {
	logged(ctx, logging.Critical, a)
}

// Alert send Application log.
func Alert(ctx context.Context, a interface{}) {
	logged(ctx, logging.Alert, a)
}

// Emergency send Application log.
func Emergency(ctx context.Context, a interface{}) {
	logged(ctx, logging.Emergency, a)
}

//

// Defaultf send Application log.
func Defaultf(ctx context.Context, format string, a ...interface{}) {
	logged(ctx, logging.Default, fmt.Sprintf(format, a...))
}

// Debugf send Application log.
func Debugf(ctx context.Context, format string, a ...interface{}) {
	logged(ctx, logging.Debug, fmt.Sprintf(format, a...))
}

// Infof send Application log.
func Infof(ctx context.Context, format string, a ...interface{}) {
	logged(ctx, logging.Info, fmt.Sprintf(format, a...))
}

// Noticef send Application log.
func Noticef(ctx context.Context, format string, a ...interface{}) {
	logged(ctx, logging.Notice, fmt.Sprintf(format, a...))
}

// Warningf send Application log.
func Warningf(ctx context.Context, format string, a ...interface{}) {
	logged(ctx, logging.Warning, fmt.Sprintf(format, a...))
}

// Errorf send Application log.
func Errorf(ctx context.Context, format string, a ...interface{}) {
	logged(ctx, logging.Error, fmt.Sprintf(format, a...))
}

// Criticalf send Application log.
func Criticalf(ctx context.Context, format string, a ...interface{}) {
	logged(ctx, logging.Critical, fmt.Sprintf(format, a...))
}

// Alertf send Application log.
func Alertf(ctx context.Context, format string, a ...interface{}) {
	logged(ctx, logging.Alert, fmt.Sprintf(format, a...))
}
