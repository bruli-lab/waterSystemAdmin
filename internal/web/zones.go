package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	"github.com/bruli/waterSystemAdmin/internal/model/zones"
)

var availableRelays = []int{1, 2, 3}

type zonesPage struct {
	Zones  []zones.Zone
	Loaded bool
}

func (a *App) Zones(w http.ResponseWriter, r *http.Request) {
	v := view{Template: "zones.html", Page: pageZones}
	list, err := a.api.Zones.FindAll(r.Context())
	if err != nil {
		a.log.ErrorContext(r.Context(), "error finding zones", slog.String("error", err.Error()))
		v.Error = "Could not read the zones: " + err.Error()
		v.Data = zonesPage{}
	} else {
		v.Data = zonesPage{Zones: list, Loaded: true}
	}
	a.render(w, r, &v)
}

func (a *App) DeleteZone(w http.ResponseWriter, r *http.Request) {
	if err := a.api.Zones.Delete(r.Context(), r.PathValue("id")); err != nil {
		a.failure(w, r, "/zones", "Could not delete the zone", err)
		return
	}
	a.success(w, r, "/zones", "Zone deleted.")
}

func (a *App) Clean(w http.ResponseWriter, r *http.Request) {
	if err := a.api.Execute.SendClean(r.Context(), r.PathValue("id")); err != nil {
		a.failure(w, r, "/zones", "Could not start the clean process", err)
		return
	}
	a.success(w, r, "/zones", "Clean process started.")
}

type zoneForm struct {
	Zone   zones.Zone
	Relays []int
	Edit   bool
}

func (a *App) CreateZone(w http.ResponseWriter, r *http.Request) {
	v := view{Template: "zone_form.html", Page: pageZones, Data: zoneForm{Relays: availableRelays}}
	if r.Method != http.MethodPost {
		a.render(w, r, &v)
		return
	}
	z, err := parseZoneForm(r, r.FormValue("id"))
	if err != nil {
		v.Error, v.Data = err.Error(), zoneForm{Zone: z, Relays: availableRelays}
		a.render(w, r, &v)
		return
	}
	if err = a.api.Zones.Create(r.Context(), &z); err != nil {
		a.log.ErrorContext(r.Context(), "error creating zone", slog.String("error", err.Error()))
		v.Error, v.Data = "Could not create the zone: "+err.Error(), zoneForm{Zone: z, Relays: availableRelays}
		a.render(w, r, &v)
		return
	}
	a.success(w, r, "/zones", fmt.Sprintf("Zone %q created.", z.Name))
}

func (a *App) UpdateZone(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	v := view{Template: "zone_form.html", Page: pageZones}
	if r.Method == http.MethodPost {
		z, err := parseZoneForm(r, id)
		if err == nil {
			err = a.api.Zones.Update(r.Context(), &z)
			if err == nil {
				a.success(w, r, "/zones", fmt.Sprintf("Zone %q updated.", z.Name))
				return
			}
			a.log.ErrorContext(r.Context(), "error updating zone", slog.String("error", err.Error()))
		}
		v.Error, v.Data = err.Error(), zoneForm{Zone: z, Relays: availableRelays, Edit: true}
		a.render(w, r, &v)
		return
	}

	list, err := a.api.Zones.FindAll(r.Context())
	if err != nil {
		a.failure(w, r, "/zones", "Could not read the zones", err)
		return
	}
	i := slices.IndexFunc(list, func(z zones.Zone) bool { return z.ID == id })
	if i < 0 {
		a.redirectWithFlash(w, r, "/zones", flashError, fmt.Sprintf("Zone %q does not exist.", id))
		return
	}
	v.Data = zoneForm{Zone: list[i], Relays: availableRelays, Edit: true}
	a.render(w, r, &v)
}

func parseZoneForm(r *http.Request, id string) (zones.Zone, error) {
	z := zones.Zone{ID: id, Name: r.FormValue("name")}
	rel, err := parseRelays(r.Form["relays"])
	z.Relays = rel
	switch {
	case err != nil:
		return z, err
	case id == "" || z.Name == "" || len(rel) == 0:
		return z, errors.New("all fields are required")
	}
	return z, nil
}

func parseRelays(values []string) ([]int, error) {
	rel := make([]int, 0, len(values))
	for _, val := range values {
		num, err := strconv.Atoi(val)
		if err != nil || !slices.Contains(availableRelays, num) || slices.Contains(rel, num) {
			return nil, errors.New("invalid relay")
		}
		rel = append(rel, num)
	}
	return rel, nil
}

type executeForm struct {
	ZoneID     string
	Fertilizer bool
}

func (a *App) Execute(w http.ResponseWriter, r *http.Request) { a.executeZone(w, r, false) }

func (a *App) Fertilizer(w http.ResponseWriter, r *http.Request) { a.executeZone(w, r, true) }

func (a *App) executeZone(w http.ResponseWriter, r *http.Request, fertilizer bool) {
	id := r.PathValue("id")
	v := view{Template: "execute.html", Page: pageZones, Data: executeForm{ZoneID: id, Fertilizer: fertilizer}}
	if r.Method != http.MethodPost {
		a.render(w, r, &v)
		return
	}
	seconds, err := parseSeconds(r.FormValue("seconds"))
	if err != nil {
		v.Error = err.Error()
		a.render(w, r, &v)
		return
	}
	exe := &zones.Execution{ID: id, Seconds: seconds, Fertilizer: fertilizer}
	if fertilizer {
		err = a.api.Execute.SendFertilizer(r.Context(), exe)
	} else {
		err = a.api.Execute.SendExecution(r.Context(), exe)
	}
	if err != nil {
		a.log.ErrorContext(r.Context(), "error executing zone", slog.String("id", id), slog.String("error", err.Error()))
		v.Error = "Could not execute the zone: " + err.Error()
		a.render(w, r, &v)
		return
	}
	what := "with water only"
	if fertilizer {
		what = "with water and fertilizer"
	}
	a.success(w, r, "/zones", fmt.Sprintf("Zone %q watering %s for %d s.", id, what, seconds))
}
