package googleidtoken

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/toyo/gcp/googlecloudlogging"
	"google.golang.org/api/idtoken"
)

// The Authorization header is never logged: it is a valid token.

// audience is the URL the request was sent to. Cloud Tasks makes the OIDC token for the URL of the task when the
// task has no audience, so a token for another service or path (e.g. one left in a log) is rejected.
func audience(r *http.Request) string {
	return `https://` + r.Host + r.URL.Path
}

// ValidateByEmail passes requests with an ID token of the service account email for this URL.
func ValidateByEmail(handler func(http.ResponseWriter, *http.Request), email string) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := googlecloudlogging.ContextInit(r.Context(), r)

		authorization := r.Header.Get(`Authorization`)

		authslice := strings.SplitN(authorization, ` `, 2)
		if authslice[0] == `Bearer` && len(authslice) >= 2 {
			idtokenstr := authslice[1]

			if p, err := idtoken.Validate(ctx, idtokenstr, audience(r)); p != nil && err == nil {
				if em, ok := p.Claims[`email`].(string); ok && em == email {
					slog.DebugContext(ctx, `ID Token Validated by email`, `email`, em)
					handler(w, r)
				} else {
					slog.ErrorContext(ctx, `ID Token Validated but wrong email`, `email`, em)
					w.Header().Add(`WWW-Authenticate`, `Bearer error="invalid_token"`)
					w.WriteHeader(http.StatusForbidden)
				}
			} else {
				slog.ErrorContext(ctx, `ID Token Not Validated`, `audience`, audience(r), `err`, err)
				w.Header().Add(`WWW-Authenticate`, `Bearer error="invalid_token"`)
				w.WriteHeader(http.StatusUnauthorized)
			}
		} else {
			slog.ErrorContext(ctx, `No Bearer token in the Authorization header`)
			w.Header().Add(`WWW-Authenticate`, `Bearer realm="realm"`)
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}
