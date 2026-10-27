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
		queue, retry := r.Header.Get(`X-Cloudtasks-Queuename`), r.Header.Get(`X-Cloudtasks-Taskretrycount`)
		slog.DebugContext(ctx, `Cloudtask authorization success`, `task`, task, `queue`, queue, `retry`, retry)
		if retry == alertRetry {
			// キューは再試行の回数と期限の両方に達すると、知らせずにタスクを消す（例：jmx・chatbot は 5 回と 1 時間）。
			// 消える前に気づけるよう、失敗が続いているタスクを 1 回だけ知らせる（DECISIONS.md「Cloud Tasks の再試行の期限」）。
			slog.Log(ctx, googlecloudlogging.LevelCritical, `The task keeps failing, it is dropped at the retry limit of the queue`,
				`task`, task, `queue`, queue, `retry`, retry)
		}
		next(w, r)
	}
}

// alertRetry は、CRITICAL で知らせる再試行の回数（X-Cloudtasks-Taskretrycount。初回は 0）。間隔が 10 秒から倍々なら、
// 最初の失敗から約 5 分後。
const alertRetry = `5`

//

func Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(MiddlewareFunc(handler.ServeHTTP))
}
