package cloudtask

import (
	"log/slog"
	"net/http"

	"github.com/toyo/gcp/googlecloudlogging"
)

// MiddlewareFunc passes only requests from Cloud Tasks (with X-Cloudtasks-Taskname). The header is not
// authentication; Cloud Run IAM or googleidtoken checks the token. Only the task headers are logged: the
// Authorization header is a valid token.
func MiddlewareFunc(next http.HandlerFunc) http.HandlerFunc {

	const cloudtaskHeader = "X-Cloudtasks-Taskname" // "X-Appengine-Taskname"

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = googlecloudlogging.ContextInit(ctx, r)

		task := r.Header.Get(cloudtaskHeader)
		if task == `` {
			slog.ErrorContext(ctx, `Not a request from Cloud Tasks`, `missing`, cloudtaskHeader, `url`, r.URL.String())
			http.Error(w, "Cloudtask authorization failed", http.StatusBadRequest)
			return
		}
		slog.DebugContext(ctx, `Cloudtask authorization success`, `task`, task,
			`queue`, r.Header.Get(`X-Cloudtasks-Queuename`), `retry`, r.Header.Get(`X-Cloudtasks-Taskretrycount`))
		next(w, r)
	}
}

//

func Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(MiddlewareFunc(handler.ServeHTTP))
}
