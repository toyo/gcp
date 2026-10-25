package googleidtoken

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAudience: the audience is the URL of the task, as Cloud Tasks makes it, e.g.
// https://line-cwauu7hyba-an.a.run.app/tq/Push.
func TestAudience(t *testing.T) {
	r := httptest.NewRequest(`POST`, `https://line-cwauu7hyba-an.a.run.app/tq/Push`, nil)
	if a := audience(r); a != `https://line-cwauu7hyba-an.a.run.app/tq/Push` {
		t.Errorf("audience %s", a)
	}
}

func TestValidateByEmailRejectsNoToken(t *testing.T) {
	called := false
	h := ValidateByEmail(func(http.ResponseWriter, *http.Request) { called = true }, `jmxservice@tyo-sv.iam.gserviceaccount.com`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(`POST`, `https://line-cwauu7hyba-an.a.run.app/tq/Push`, nil)
	r.Header.Set(`Authorization`, `Bearer not-a-token`)
	h(w, r)
	if called || w.Code != http.StatusUnauthorized {
		t.Errorf("called %v, status %d", called, w.Code)
	}
}
