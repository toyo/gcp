package cloudtask

import (
	"log/slog"
	"net/http"

	"github.com/toyo/gcp/googlecloudlogging"
)

func MiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {

	const cloudtaskHeader = "X-Cloudtasks-Taskname" // "X-Appengine-Taskname"

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = googlecloudlogging.ContextInit(ctx, r)

		t, ok := r.Header[cloudtaskHeader]
		if !ok || len(t[0]) == 0 {
			slog.ErrorContext(ctx, `Cloudtask authorization failed. No "+cloudtaskHeader+" request header found.`,
				`Header`, r.Header)
			http.Error(w, "Cloudtask authorization failed", http.StatusBadRequest)
			return
		} else {
			slog.DebugContext(ctx, `Cloudtask authorization success`, `Header`, r.Header)
			next(w, r)
		}
	}
}

//

func Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(MiddlewareFunc(handler.ServeHTTP))
}
