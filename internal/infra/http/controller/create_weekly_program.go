package controller

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bruli/waterSystemAdmin/internal/domain/programs"
	"github.com/bruli/waterSystemAdmin/internal/domain/zones"
	pongo3 "github.com/flosch/pongo2/v6"
)

func CreateWeeklyProgram(set *pongo3.TemplateSet, zonesSvc *zones.FindZones, createSvc *programs.CreateWeekly, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl, err := set.FromFile("create_weekly_program.html")
		if err != nil {
			log.ErrorContext(r.Context(), "error parsing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		context := map[string]any{
			"page": "programs",
		}
		foundZones, err := zonesSvc.Find(r.Context())
		if err != nil {
			log.ErrorContext(r.Context(), "error finding zones", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		switch {
		case len(foundZones) == 0:
			context["error_msg"] = "No zones found"
		default:
			context["zones"] = foundZones
		}
		if r.Method == http.MethodPost {
			processCreateWeeklyForm(r, context, createSvc)
		}

		if err = tpl.ExecuteWriter(context, w); err != nil {
			log.ErrorContext(r.Context(), "error executing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func processCreateWeeklyForm(r *http.Request, context map[string]any, svc *programs.CreateWeekly) {
	if err := r.ParseForm(); err != nil {
		context["error_msg"] = "invalid form"
		return
	}
	day, err := programs.ParseWeekDay(r.FormValue("day"))
	if err != nil {
		context["error_msg"] = err.Error()
		return
	}
	hours := r.Form["hours[]"]
	if len(hours) == 0 {
		context["error_msg"] = "hours is required"
		return
	}
	prgms := make([]programs.Program, len(hours))
	for i, hour := range hours {
		execSeconds := r.Form["executions_"+strconv.Itoa(i)+"_seconds[]"]

		if len(execSeconds) == 0 {
			context["error_msg"] = "hours is required"
			return
		}
		h, err := programs.ParseHour(hour)
		if err != nil {
			context["error_msg"] = err.Error()
			return
		}
		exec := make([]programs.Execution, len(execSeconds))
		for n, execSecond := range execSeconds {
			sec, err := strconv.ParseInt(execSecond, 10, 64)
			if err != nil {
				context["error_msg"] = "invalid seconds"
				return
			}
			foundZones := r.Form["executions_"+strconv.Itoa(i)+"_zones_"+strconv.Itoa(n)+"[]"]
			if len(foundZones) == 0 {
				context["error_msg"] = "zone is required"
				return
			}
			exec[n] = programs.Execution{
				Seconds: int(sec),
				Zones:   foundZones,
			}
		}
		prgms[i] = programs.Program{
			Executions: exec,
			Hour:       h,
		}
	}
	weekly := programs.Weekly{
		WeekDay:  day,
		Programs: prgms,
	}
	if err = svc.Create(r.Context(), &weekly); err != nil {
		context["error_msg"] = err.Error()
		return
	}
	context["success"] = true
}
