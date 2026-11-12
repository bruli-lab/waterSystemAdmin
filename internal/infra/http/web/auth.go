package web

import (
	"log/slog"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/model/password"
)

func (a *App) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		exists, err := a.password.Exists(ctx)
		if err != nil {
			a.log.ErrorContext(ctx, "error checking if password exists", slog.String("error", err.Error()))
			http.Error(w, "error checking password", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Redirect(w, r, "/password", http.StatusSeeOther)
			return
		}
		session, err := a.sessions.Get(r, sessionName)
		if err != nil {
			a.log.WarnContext(ctx, "invalid session cookie, redirecting to login", slog.String("error", err.Error()))
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if auth, _ := session.Values["authenticated"].(bool); !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	v := view{Template: "login.html", Root: "auth"}
	if r.Method != http.MethodPost {
		a.render(w, r, &v)
		return
	}
	ctx := r.Context()
	stored, err := a.password.Read(ctx)
	if err != nil {
		a.log.ErrorContext(ctx, "error reading password", slog.String("error", err.Error()))
		v.Error = "Could not read the password."
		a.render(w, r, &v)
		return
	}
	if stored.Compare(r.FormValue("password")) != nil {
		v.Error = "Invalid password"
		a.render(w, r, &v)
		return
	}
	session, _ := a.sessions.Get(r, sessionName)
	session.Values["authenticated"] = true
	if err = session.Save(r, w); err != nil {
		a.log.ErrorContext(ctx, "error saving session", slog.String("error", err.Error()))
		http.Error(w, "error saving session", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := a.sessions.Get(r, sessionName)
	session.Values["authenticated"] = false
	if err := session.Save(r, w); err != nil {
		a.log.ErrorContext(r.Context(), "error saving session", slog.String("error", err.Error()))
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *App) Password(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	exists, err := a.password.Exists(ctx)
	if err != nil {
		a.log.ErrorContext(ctx, "error checking if password exists", slog.String("error", err.Error()))
		http.Error(w, "error checking password", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	v := view{Template: "password.html", Root: "auth", NoStatus: true}
	if r.Method != http.MethodPost {
		a.render(w, r, &v)
		return
	}
	pass, rePass := r.FormValue("password"), r.FormValue("re-password")
	switch {
	case pass == "":
		v.Error = "Password is required"
	case pass != rePass:
		v.Error = "Passwords do not match"
	}
	if v.Error != "" {
		a.render(w, r, &v)
		return
	}
	hashed, err := password.NewPassword(pass)
	if err == nil {
		err = a.password.Save(ctx, hashed)
	}
	if err != nil {
		a.log.ErrorContext(ctx, "error creating password", slog.String("error", err.Error()))
		v.Error = "Could not save the password."
		a.render(w, r, &v)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
