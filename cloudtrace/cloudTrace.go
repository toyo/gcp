package cloudtrace

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/toyo/gcp/gce"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
)

func init() {
	ctx := context.Background()
	if gce.GetProjectID(ctx) != `` {
		// Create and register a OpenCensus Stackdriver Trace exporter.
		exporter, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID: gce.GetProjectID(ctx),
		})
		if err != nil {
			fmt.Println(err.Error())
		}
		trace.RegisterExporter(exporter)
	}
}

// Context return context and Span. Need Span.Close()
func Context(r *http.Request) (context.Context, *trace.Span) {
	c := r.Context()
	if gce.GetProjectID(c) != `` {

		HTTPFormat := &tracecontext.HTTPFormat{}

		if spanContext, ok := HTTPFormat.SpanContextFromRequest(r); ok {
			return trace.StartSpanWithRemoteParent(c, r.RequestURI, spanContext)
		} else if r.Header.Get(`X-Cloud-Trace-Context`) != `` {
			s := r.Header.Get(`X-Cloud-Trace-Context`)
			s1 := strings.Split(s, `;`)
			s2 := strings.Split(s1[0], `/`)

			traceID := s2[0]

			spanid, err := strconv.ParseInt(s2[1], 10, 64)
			var spanID string
			if err == nil {
				spanID = fmt.Sprintf(`%016x`, spanid)
			}

			var traceFlags string
			if len(s1) >= 2 && len(s1[1]) >= 3 {
				traceFlags = `0` + string(s1[1][2])
			} else {
				traceFlags = `01`
			}

			version := `00`
			versionFormat := traceID + "-" + spanID + "-" + traceFlags

			traceparent := version + `-` + versionFormat

			if spanContext, ok := HTTPFormat.SpanContextFromHeaders(traceparent, ``); ok {
				r.Header[`traceparent`] = []string{traceparent}
				return trace.StartSpanWithRemoteParent(c, r.RequestURI, spanContext)
			}
		}
		return trace.StartSpan(c, r.RequestURI)
	} else {
		return c, nil
	}
}
