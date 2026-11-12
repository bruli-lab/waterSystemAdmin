package controller

import (
	"log/slog"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/password"

	"github.com/gorilla/sessions"
)

type MiddlewareFunc func(next http.HandlerFunc) http.HandlerFunc

func AuthMiddleware(store *sessions.CookieStore, exists *password.Exists, log *slog.Logger) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			passwdExists, err := exists.Exists(r.Context())
			if err != nil {
				log.ErrorContext(r.Context(), "error checking if password exists", slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !passwdExists {
				http.Redirect(w, r, "/password", http.StatusSeeOther)
				return
			}
			session, err := store.Get(r, "session")
			if err != nil {
				// A cookie signed with another key (e.g. after a restart with a new SESSION_KEY)
				// can't be decoded: treat it as logged out so the login overwrites it.
				log.WarnContext(r.Context(), "invalid session cookie, redirecting to login", slog.String("error", err.Error()))
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if session == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			authenticated, ok := session.Values["authenticated"]
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			auth, _ := authenticated.(bool)
			if auth {
				next(w, r)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	}
}
