package controller

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bruli/waterSystemAdmin/internal/domain/status"
	"github.com/bruli/waterSystemAdmin/internal/domain/zones"
	pongo3 "github.com/flosch/pongo2/v6"
)

func CreateZone(set *pongo3.TemplateSet, svc *zones.Create, stSvc *status.FindStatus, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl, err := set.FromFile("create_zone.html")
		if err != nil {
			log.ErrorContext(r.Context(), "error parsing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tplCtx, failed := buildStatusInTemplateController(r.Context(), stSvc)
		if failed {
			log.ErrorContext(r.Context(), "error building status in template controller")
		}
		tplCtx.Add("page", "zones")
		tplCtx.Add("relays", []int{1, 2, 3})
		if r.Method == http.MethodPost {
			processCreateZoneForm(r, tplCtx, svc)
		}

		if err = tpl.ExecuteWriter(tplCtx.toPongoContext(), w); err != nil {
			log.ErrorContext(r.Context(), "error executing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func processCreateZoneForm(r *http.Request, context TemplateController, svc *zones.Create) {
	id := r.FormValue("id")
	name := r.FormValue("name")
	relayValues := r.Form["relays"]

	if id == "" || name == "" || len(relayValues) == 0 {
		context.AddError("All fields are required")
		return
	}
	rel := make([]int, len(relayValues))
	for i, val := range relayValues {
		num, err := strconv.Atoi(val)
		if err != nil {
			context.AddError("Invalid relay value")
			return
		}
		rel[i] = num
	}
	if err := svc.Create(r.Context(), &zones.Zone{
		ID:     id,
		Name:   name,
		Relays: rel,
	}); err != nil {
		context.AddError("Create zone failed.")
		return
	}

	context.Add("success", true)
}
