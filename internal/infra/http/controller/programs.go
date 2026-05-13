package controller

import (
	"log/slog"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/programs"
	"github.com/bruli/waterSystemAdmin/internal/domain/status"
	pongo3 "github.com/flosch/pongo2/v6"
)

type Section struct {
	ID    string
	Title string
	Data  []programs.FertilizerProgram
}

type Weekly struct {
	WeekDay  string
	Programs []programs.Program
}

func Programs(set *pongo3.TemplateSet, svc *programs.FindAllPrograms, stSvc *status.FindStatus, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl, err := set.FromFile("programs.html")
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
		progrms, err := svc.Find(r.Context())
		switch {
		case err != nil:
			log.ErrorContext(r.Context(), "error finding programs", slog.String("error", err.Error()))
			tplCtx.AddError(err.Error())
		default:
			sections := []Section{
				{ID: "collapseOne", Title: "Daily", Data: progrms.Daily},
				{ID: "collapseTwo", Title: "Odd", Data: progrms.Odd},
				{ID: "collapseThree", Title: "Even", Data: progrms.Even},
			}
			tplCtx.Add("sections", sections)
			tplCtx.Add("weekly", buildWeekly(progrms.Weekly))
			tplCtx.Add("temperature", progrms.Temperature)
		}
		if err = tpl.ExecuteWriter(tplCtx.toPongoContext(), w); err != nil {
			log.ErrorContext(r.Context(), "error executing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func buildWeekly(progrms []programs.Weekly) []Weekly {
	weekly := make([]Weekly, len(progrms))
	for i, week := range progrms {
		weekly[i] = Weekly{
			WeekDay:  week.WeekDay.String(),
			Programs: week.Programs,
		}
	}
	return weekly
}
