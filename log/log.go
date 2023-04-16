package log

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/toyo/gcp/gce"
)

// logged is for logging.
func logged(ctx context.Context, severity logging.Severity, payload interface{}) {

	const stdout = false

	if cs, ok := ctx.Value(tokenContextSaver).(contextSaver); ok {

		if !stdout && gce.GetProjectID() != `` {

			e := logging.Entry{
				Timestamp: time.Now(),
				Payload:   payload,
				Severity:  severity,
				Trace:     cs.trace,
				SpanID:    cs.spanID,
			}

			if cs.traceSampled != nil {
				e.TraceSampled = *cs.traceSampled
			}

			if pc, file, line, ok := runtime.Caller(2); ok {
				e.SourceLocation = &loggingpb.LogEntrySourceLocation{
					File:     file,
					Line:     int64(line),
					Function: runtime.FuncForPC(pc).Name(),
				}
			}

			client.Logger(`github.com/toyo/gcp/log`).Log(e)
		} else {
			var message string

			switch v := payload.(type) {
			case string:
				message = v
			case fmt.Stringer:
				message = v.String()
			case fmt.GoStringer:
				message = v.GoString()
			default:
				var builder strings.Builder
				json.NewEncoder(&builder).Encode(payload)
				message = builder.String()
			}
			if len(message) < 65536 {

				t := time.Now()

				e := Entry{
					Severity:     severity.String(),
					Message:      message,
					Trace:        cs.trace,
					SpanID:       cs.spanID,
					TraceSampled: cs.traceSampled,
					TimeStamp:    &t,
				}

				if pc, file, line, ok := runtime.Caller(2); ok {
					e.SourceLocation = &loggingpb.LogEntrySourceLocation{
						File:     file,
						Line:     int64(line),
						Function: runtime.FuncForPC(pc).Name(),
					}
				}
				err := json.NewEncoder(os.Stdout).Encode(&e)
				if err != nil {
					fmt.Printf("%s: %v\n", err.Error(), e)
				}
			} else {
				log.Println(severity.String() + ": " + message)
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
