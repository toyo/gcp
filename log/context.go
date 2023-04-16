package log

import (
	"context"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/toyo/gcp/gce"
	"go.opencensus.io/trace"
)

type contextKey string

const tokenContextSaver contextKey = "AppEngine2ndGenerationLogger-Saver"

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
