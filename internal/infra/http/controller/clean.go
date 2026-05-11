package controller

import (
	"log/slog"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/execution"
	"github.com/bruli/waterSystemAdmin/internal/domain/status"
	pongo3 "github.com/flosch/pongo2/v6"
)

func Clean(set *pongo3.TemplateSet, svc *execution.Clean, stSvc *status.FindStatus, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if r.Method != http.MethodPost {
			log.ErrorContext(r.Context(), "Method not allowed", slog.String("method", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := svc.Execute(r.Context(), id); err != nil {
			log.ErrorContext(r.Context(), "error activating", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/zones", http.StatusSeeOther)

	}
}
