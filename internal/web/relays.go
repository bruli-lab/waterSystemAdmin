package web

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/bruli-lab/waterSystemAdmin/internal/api"
)

type relay struct {
	ID   int
	Name string
}

var relays = []relay{
	{ID: 1, Name: "Relay 1"},
	{ID: 2, Name: "Relay 2"},
	{ID: 3, Name: "Relay 3"},
	{ID: 4, Name: "Relay 4"},
	{ID: 5, Name: "Fertilizer pump"},
	{ID: 6, Name: "Air pump"},
	{ID: 7, Name: "Clean valve"},
	{ID: 8, Name: "Fertilizer valve"},
}

var relayTestErrors = map[int]string{
	http.StatusBadRequest:          "the relay does not exist",
	http.StatusUnauthorized:        "the API token is missing or invalid",
	http.StatusConflict:            "a clean cycle is running",
	http.StatusInternalServerError: "the GPIO failed, the relay may be in use by an irrigation (see journalctl -u water_system)",
}

func (a *App) Relays(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, &view{Template: "relays.html", Page: pageRelays, Title: "Relays", Data: relays, Refresh: true})
}

func (a *App) TestRelay(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	i := slices.IndexFunc(relays, func(rl relay) bool { return rl.ID == id })
	if err != nil || i < 0 {
		a.failure(w, r, "/relays", "Could not test the relay", errors.New("unknown relay"))
		return
	}
	rl := relays[i]
	if err := a.api.Relays.Test(r.Context(), rl.ID); err != nil {
		msg := "Could not test " + rl.Name
		if se, ok := errors.AsType[*api.StatusError](err); ok {
			if reason, known := relayTestErrors[se.Code]; known {
				err = fmt.Errorf("%s: %w", reason, err)
			}
		}
		a.failure(w, r, "/relays", msg, err)
		return
	}
	a.success(w, r, "/relays", rl.Name+" switched on and off correctly.")
}
