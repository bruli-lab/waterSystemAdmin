package controller

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bruli/waterSystemAdmin/internal/domain/programs"
	"github.com/bruli/waterSystemAdmin/internal/domain/status"
	"github.com/bruli/waterSystemAdmin/internal/domain/zones"
	"github.com/flosch/pongo2/v6"
)

func CreateProgram(tplSet *pongo2.TemplateSet, zonesSvc *zones.FindZones, createSvc *programs.Create, stSvc *status.FindStatus, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl, err := tplSet.FromFile("create_program.html")
		if err != nil {
			log.ErrorContext(r.Context(), "error parsing template",
				slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		programType := r.URL.Query().Get("type")
		tplCtx, failed := buildStatusInTemplateController(r.Context(), stSvc)
		if failed {
			log.ErrorContext(r.Context(), "error building status in template controller")
		}
		tplCtx.Add("page", "programs")
		tplCtx.Add("type", programType)
		foundZones, err := zonesSvc.Find(r.Context())
		if err != nil {
			log.ErrorContext(r.Context(), "error finding zones",
				slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		switch {
		case len(foundZones) == 0:
			tplCtx.AddError("no zones found")
		default:
			tplCtx.Add("zones", foundZones)
		}

		if r.Method == http.MethodPost {
			processCreateProgramForm(r, tplCtx.toPongoContext(), createSvc, programType)
		}
		if err = tpl.ExecuteWriter(tplCtx.toPongoContext(), w); err != nil {
			log.ErrorContext(r.Context(), "error executing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func processCreateProgramForm(r *http.Request, context map[string]any, svc *programs.Create, programType string) {
	if err := r.ParseForm(); err != nil {
		context["error_msg"] = "Error parsing form"
		return
	}

	hour, err := programs.ParseHour(r.FormValue("hour"))
	if err != nil {
		context["error_msg"] = "Invalid hour"
		return
	}
	seconds := r.Form["seconds[]"]

	var fertilizer bool
	fertilizer, _ = strconv.ParseBool(r.FormValue("fertilizer"))
	executions := make([]programs.Execution, 0)

	for i, sec := range seconds {
		zonesKey := fmt.Sprintf("zones_%d[]", i)
		foundZones := r.Form[zonesKey]

		if len(foundZones) == 0 {
			context["error_msg"] = "Invalid zones"
			return
		}

		s, err := strconv.Atoi(sec)
		if err != nil {
			context["error_msg"] = "Invalid seconds"
			return
		}

		executions = append(executions, programs.Execution{
			Seconds: s,
			Zones:   foundZones,
		})
	}
	pr := programs.FertilizerProgram{
		Executions: executions,
		Hour:       hour,
		Fertilizer: fertilizer,
	}
	progType, err := programs.ParseProgramType(programType)
	if err != nil {
		context["error_msg"] = "Invalid program type"
		return
	}
	if err = svc.Create(r.Context(), &pr, progType); err != nil {
		context["error_msg"] = "Error creating program"
		return
	}
	context["success"] = true
}
