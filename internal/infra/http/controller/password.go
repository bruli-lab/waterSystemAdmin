package controller

import (
	"log/slog"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/password"
	pongo3 "github.com/flosch/pongo2/v6"
)

// Password only allows setting the initial password. Once a password exists,
// the endpoint redirects to login so it can't be used to take over the account.
func Password(set *pongo3.TemplateSet, create *password.Create, exists *password.Exists, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		passwdExists, err := exists.Exists(r.Context())
		if err != nil {
			log.ErrorContext(r.Context(), "error checking if password exists", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if passwdExists {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
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
			case passw == "":
				context["error_msg"] = "Password is required"
			case passw != rePassword:
				context["error_msg"] = "Passwords do not match"
			default:
				if err = create.Create(r.Context(), passw); err != nil {
					log.ErrorContext(r.Context(), "error creating password", slog.String("error", err.Error()))
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
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
