package controller

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/programs"
	"github.com/bruli/waterSystemAdmin/internal/domain/status"
	pongo3 "github.com/flosch/pongo2/v6"
)

func RemoveWeeklyProgram(set *pongo3.TemplateSet, stSvc *status.FindStatus, removeSvc *programs.RemoveWeekly, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.ErrorContext(r.Context(), "Method not allowed",
				slog.String("method", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tpl, err := set.FromFile("redirect_message.html")
		if err != nil {
			log.ErrorContext(r.Context(), "error parsing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tplCtx, failed := buildStatusInTemplateController(r.Context(), stSvc)
		if failed {
			log.ErrorContext(r.Context(), "error building status in template controller")
		}
		tplCtx.Add("page", "programs")

		err = removingWeeklyProgram(r, removeSvc)
		switch {
		case err != nil:
			log.ErrorContext(r.Context(), "error removing program", slog.String("error", err.Error()))
			tplCtx.AddError(err.Error())
		default:
			tplCtx.Add("success", true)
		}

		if err = tpl.ExecuteWriter(tplCtx.toPongoContext(), w); err != nil {
			log.ErrorContext(r.Context(), "error executing template", slog.String("error", err.Error()))

			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func removingWeeklyProgram(r *http.Request, svc *programs.RemoveWeekly) error {
	day, err := programs.ParseWeekDay(r.PathValue("weekday"))
	if err != nil {
		return fmt.Errorf("invalid day: %w", err)
	}
	if err = svc.Remove(r.Context(), &day); err != nil {
		return fmt.Errorf("faild remove program: %w", err)
	}
	return nil
}
