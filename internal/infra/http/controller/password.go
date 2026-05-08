package controller

import (
	"log/slog"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/password"
	pongo3 "github.com/flosch/pongo2/v6"
)

func Password(set *pongo3.TemplateSet, create *password.Create, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl, err := set.FromFile("password.html")
		if err != nil {
			log.ErrorContext(r.Context(), "error parsing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		context := map[string]any{}
		if r.Method == http.MethodPost {
			passw := r.FormValue("password")
			rePassword := r.FormValue("re-password")
			switch {
			case passw != rePassword:
				context["error_msg"] = "Passwords do not match"
			default:
				if err = create.Create(r.Context(), passw); err != nil {
					log.ErrorContext(r.Context(), "error creating password", slog.String("error", err.Error()))
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

		}

		if err = tpl.ExecuteWriter(context, w); err != nil {
			log.ErrorContext(r.Context(), "error executing template", slog.String("error", err.Error()))
		}
	}
}
