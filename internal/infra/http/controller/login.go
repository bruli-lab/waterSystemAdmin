package controller

import (
	"log/slog"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/password"
	"github.com/bruli/waterSystemAdmin/internal/domain/status"

	"github.com/flosch/pongo2/v6"
	"github.com/gorilla/sessions"
)

func Login(tplSet *pongo2.TemplateSet, store *sessions.CookieStore, check *password.Check, stSvc *status.FindStatus, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl, err := tplSet.FromFile("login.html")
		if err != nil {
			log.ErrorContext(r.Context(), "error parsing template",
				slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tplCtx, failed := buildStatusInTemplateController(r.Context(), stSvc)
		if failed {
			log.ErrorContext(r.Context(), "error building status in template controller")
		}
		if r.Method == http.MethodPost {
			valid, err := check.Check(r.Context(), r.FormValue("password"))
			if err != nil {
				log.ErrorContext(r.Context(), "error checking password", slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			switch {
			case valid:
				session, _ := store.Get(r, "session")
				session.Values["authenticated"] = true
				if err = session.Save(r, w); err != nil {
					log.ErrorContext(r.Context(), "error saving session",
						slog.String("error", err.Error()))
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			default:
				tplCtx.AddError("Invalid password")
			}
		}

		if err = tpl.ExecuteWriter(tplCtx.toPongoContext(), w); err != nil {
			log.ErrorContext(r.Context(), "error executing template",
				slog.String("error", err.Error()))
			http.Error(w, "Error executant login template ", http.StatusInternalServerError)
		}
	}
}
