package controller

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bruli/waterSystemAdmin/internal/domain/status"
	"github.com/bruli/waterSystemAdmin/internal/domain/zones"
	pongo3 "github.com/flosch/pongo2/v6"
)

func UpdateZone(set *pongo3.TemplateSet, findSvc *zones.FindZones, update *zones.Update, stSvc *status.FindStatus, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl, err := set.FromFile("update_zone.html")
		if err != nil {
			log.ErrorContext(r.Context(), "error parsing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id := r.PathValue("id")
		tplCtx, failed := buildStatusInTemplateController(r.Context(), stSvc)
		if failed {
			log.ErrorContext(r.Context(), "error building status in template controller")
		}
		tplCtx.Add("page", "zones")
		if r.Method == http.MethodPost {
			processUpdateZoneForm(r, id, tplCtx, update)
		}
		list, err := findSvc.Find(r.Context())
		if err != nil {
			log.ErrorContext(r.Context(), "error finding zones", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var zo *zones.Zone
		for _, zon := range list {
			if zon.ID == id {
				zo = &zon
			}
		}

		tplCtx.Add("zone", zo)
		tplCtx.Add("relays", []int{1, 2, 3, 4})
		if err = tpl.ExecuteWriter(tplCtx.toPongoContext(), w); err != nil {
			log.ErrorContext(r.Context(), "error executing template", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func processUpdateZoneForm(r *http.Request, id string, context TemplateController, update *zones.Update) {
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
	zo := &zones.Zone{
		ID:     id,
		Name:   name,
		Relays: rel,
	}
	if err := update.Update(r.Context(), zo); err != nil {
		context["error_msg"] = "Update zone failed."
		context.AddError("Update zone failed.")
		return
	}
	context.Add("success", true)
}
