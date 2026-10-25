package cloudtask

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestMiddlewareDoesNotLogToken: the Authorization header of a task is a valid token, so it is not logged.
func TestMiddlewareDoesNotLogToken(t *testing.T) {
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))

	h := MiddlewareFunc(func(http.ResponseWriter, *http.Request) {})
	for _, task := range []string{`task1`, ``} {
		r := httptest.NewRequest(`POST`, `https://chatbot.example/tq/Message`, nil)
		r.Header.Set(`Authorization`, `Bearer secret-token`)
		r.Header.Set(`X-Cloudtasks-Taskname`, task)
		r.Header.Set(`X-Cloudtasks-Queuename`, `chatbot`)
		h(httptest.NewRecorder(), r)
	}
	if strings.Contains(buf.String(), `secret-token`) {
		t.Errorf("the token is logged: %s", buf.String())
	}
	if !strings.Contains(buf.String(), `queue=chatbot`) {
		t.Errorf("the queue is not logged: %s", buf.String())
	}
}
