package controller

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bruli/waterSystemAdmin/internal/domain/execution"
	"github.com/bruli/waterSystemAdmin/internal/domain/status"
	pongo3 "github.com/flosch/pongo2/v6"
)

func Execution(set *pongo3.TemplateSet, svc *execution.ExecuteZone, stSvc *status.FindStatus, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		tpl, err := set.FromFile("execution.html")
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
		tplCtx.Add("page", "execution")
		tplCtx.Add("id", id)
		if r.Method == http.MethodPost {
			seconds, err := strconv.Atoi(r.FormValue("seconds"))
			if err != nil {
				log.ErrorContext(r.Context(), "error parsing seconds",
					slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err = svc.Execute(r.Context(), &execution.Execution{
				ID:         id,
				Seconds:    seconds,
				Fertilizer: false,
			}); err != nil {
				tplCtx.AddError(fmt.Sprintf("Error executing zone: %s", err.Error()))
				if err = tpl.ExecuteWriter(tplCtx.toPongoContext(), w); err != nil {
					log.ErrorContext(r.Context(), "error executing template", slog.String("error", err.Error()))
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
			tplCtx.Add("success", true)
			if err = tpl.ExecuteWriter(tplCtx.toPongoContext(), w); err != nil {
				log.ErrorContext(r.Context(), "error executing template", slog.String("error", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if err = tpl.ExecuteWriter(tplCtx.toPongoContext(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
