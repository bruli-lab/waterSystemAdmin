package controller

import (
	"log/slog"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/zones"
)

func DeleteZone(svc *zones.Delete, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			id := r.PathValue("id")
			if err := svc.Delete(r.Context(), id); err != nil {
				log.ErrorContext(r.Context(), "error deleting zone", slog.String("id", id), slog.String("error", err.Error()))
			}
		}

		http.Redirect(w, r, "/zones", http.StatusSeeOther)
	}
}
