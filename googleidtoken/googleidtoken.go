package googleidtoken

import (
	"net/http"
	"strings"

	"github.com/toyo/gcp/cloudtrace"
	"github.com/toyo/gcp/log"
	"google.golang.org/api/idtoken"
)

func HandleFunc(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := log.NewContextFromReq(r)

		authorization := r.Header.Get(`Authorization`)

		authslice := strings.SplitN(authorization, ` `, 2)
		if authslice[0] == `Bearer` && len(authslice) >= 2 {
			idtokenstr := authslice[1]

			if p, err := idtoken.Validate(ctx, idtokenstr, ``); p != nil && err == nil {
				log.Debugf(ctx, `ID Token Validated: %s / %#v`, p.Claims[`email`].(string), p)
				handler(w, r)
			} else {
				log.Errorf(ctx, `ID Token Not Validated: %s`, authorization)
				w.Header().Add(`WWW-Authenticate`, `Bearer error="invalid_token"`)
				w.WriteHeader(http.StatusUnauthorized)
			}
		} else {
			log.Errorf(ctx, `ID Bearer Header: %s`, authorization)
			w.Header().Add(`WWW-Authenticate`, `Bearer realm="realm"`)
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}

func ValidateByEmail(handler func(http.ResponseWriter, *http.Request), email string) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := cloudtrace.Context(r)
		ctx = log.ContextFromSpan(ctx, span)
		defer span.End()

		authorization := r.Header.Get(`Authorization`)

		authslice := strings.SplitN(authorization, ` `, 2)
		if authslice[0] == `Bearer` && len(authslice) >= 2 {
			idtokenstr := authslice[1]

			if p, err := idtoken.Validate(ctx, idtokenstr, ``); p != nil && err == nil {
				if em, ok := p.Claims[`email`].(string); ok && em == email {
					log.Debugf(ctx, `ID Token Validated by email: %s`, em)
					handler(w, r)
				} else {
					log.Errorf(ctx, `ID Token Validated but wrong email: %s`, em)
					w.Header().Add(`WWW-Authenticate`, `Bearer error="invalid_token"`)
					w.WriteHeader(http.StatusForbidden)
				}
			} else {
				log.Errorf(ctx, `ID Token Not Validated: %s`, authorization)
				w.Header().Add(`WWW-Authenticate`, `Bearer error="invalid_token"`)
				w.WriteHeader(http.StatusUnauthorized)
			}
		} else {
			log.Errorf(ctx, `ID Bearer Header: %s`, authorization)
			w.Header().Add(`WWW-Authenticate`, `Bearer realm="realm"`)
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}
