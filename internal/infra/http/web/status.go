package web

import (
	"log/slog"
	"net/http"
)

func (a *App) Status(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, &view{Template: "status.html", Page: "status"})
}

func (a *App) Activate(w http.ResponseWriter, r *http.Request) {
	if err := a.api.Activate.Activate(r.Context(), true); err != nil {
		a.failure(w, r, "/", "Could not activate the system", err)
		return
	}
	a.success(w, r, "/", "System activated.")
}

func (a *App) Deactivate(w http.ResponseWriter, r *http.Request) {
	if err := a.api.Activate.Activate(r.Context(), false); err != nil {
		a.failure(w, r, "/", "Could not deactivate the system", err)
		return
	}
	a.success(w, r, "/", "System deactivated.")
}

func (a *App) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if err := a.api.Status.Update(r.Context()); err != nil {
		a.failure(w, r, "/", "Could not update the status", err)
		return
	}
	a.success(w, r, "/", "Status updated.")
}

func (a *App) Logs(w http.ResponseWriter, r *http.Request) {
	v := view{Template: "logs.html", Page: "logs"}
	list, err := a.api.Logs.Find(r.Context())
	if err != nil {
		a.log.ErrorContext(r.Context(), "error finding logs", slog.String("error", err.Error()))
		v.Error = "Could not read the logs: " + err.Error()
	}
	v.Data = list
	a.render(w, r, &v)
}
