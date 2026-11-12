package web

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/bruli/waterSystemAdmin/internal/domain/status"
	"github.com/bruli/waterSystemAdmin/internal/infra/api"
	"github.com/bruli/waterSystemAdmin/internal/infra/disk"
	"github.com/gorilla/sessions"
)

//go:embed views/*.html
var viewsFS embed.FS

//go:embed static
var staticFS embed.FS

const (
	pageStatus   = "status"
	pageLogs     = "logs"
	pageZones    = "zones"
	pagePrograms = "programs"
	sessionName  = "session"
	flashSuccess = "flash_success"
	flashError   = "flash_error"
)

type API struct {
	Status      *api.StatusRepository
	Activate    *api.ActivateRepository
	Zones       *api.ZoneRepository
	Execute     *api.ExecuteZoneRepository
	Logs        *api.LogRepository
	Programs    *api.ProgramsRepository
	Weekly      *api.WeeklyRepository
	Temperature *api.TemperatureRepository
}

type App struct {
	api      API
	password *disk.PasswordRepository
	sessions *sessions.CookieStore
	log      *slog.Logger
	pages    map[string]*template.Template
}

func NewApp(apiClients API, passwordRepo *disk.PasswordRepository, store *sessions.CookieStore, log *slog.Logger) (*App, error) {
	pages, err := parsePages()
	if err != nil {
		return nil, err
	}
	return &App{
		api:      apiClients,
		password: passwordRepo,
		sessions: store,
		log:      log,
		pages:    pages,
	}, nil
}

func Static() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return http.StripPrefix("/static/", http.FileServer(http.FS(sub)))
}

type navLink struct {
	Active            bool
	Href, Icon, Label string
}

var templateFuncs = template.FuncMap{
	"navItem": func(current, page, href, icon, label string) navLink {
		return navLink{Active: current == page, Href: href, Icon: icon, Label: label}
	},
	"join":     strings.Join,
	"lower":    strings.ToLower,
	"contains": slices.Contains[[]int],
}

func parsePages() (map[string]*template.Template, error) {
	files, err := fs.Glob(viewsFS, "views/*.html")
	if err != nil {
		return nil, err
	}
	pages := make(map[string]*template.Template)
	for _, f := range files {
		name := strings.TrimPrefix(f, "views/")
		if name == "layout.html" {
			continue
		}
		tpl, err := template.New(name).Funcs(templateFuncs).ParseFS(viewsFS, "views/layout.html", f)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", name, err)
		}
		pages[name] = tpl
	}
	return pages, nil
}

type pageData struct {
	Page    string
	Status  *status.Status
	Error   string
	Success string
	Data    any
}

func (p pageData) Dark() bool { return p.Status == nil || !p.Status.IsDay }

func (p pageData) Cleaning() bool { return p.Status != nil && p.Status.Clean }

type view struct {
	Template string
	Page     string
	Root     string
	Data     any
	Error    string
	NoStatus bool
}

func (a *App) loadStatus(r *http.Request, pd *pageData) {
	st, err := a.api.Status.Find(r.Context())
	if err != nil {
		a.log.ErrorContext(r.Context(), "error reading status", slog.String("error", err.Error()))
		if pd.Error == "" {
			pd.Error = "Could not read the system status."
		}
		return
	}
	pd.Status = st
}

func (a *App) render(w http.ResponseWriter, r *http.Request, v *view) {
	ctx := r.Context()
	pd := pageData{Page: v.Page, Data: v.Data, Error: v.Error}

	if !v.NoStatus {
		a.loadStatus(r, &pd)
	}

	success, flashErr := a.popFlashes(w, r)
	pd.Success = success
	if pd.Error == "" {
		pd.Error = flashErr
	}

	tpl, ok := a.pages[v.Template]
	if !ok {
		a.log.ErrorContext(ctx, "unknown page", slog.String("template", v.Template))
		http.Error(w, "page not found", http.StatusInternalServerError)
		return
	}
	root := v.Root
	if root == "" {
		root = "layout"
	}
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, root, pd); err != nil {
		a.log.ErrorContext(ctx, "error rendering page", slog.String("template", v.Template), slog.String("error", err.Error()))
		http.Error(w, "error rendering page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

func (a *App) redirectWithFlash(w http.ResponseWriter, r *http.Request, to, key, msg string) {
	session, _ := a.sessions.Get(r, sessionName)
	session.AddFlash(msg, key)
	if err := session.Save(r, w); err != nil {
		a.log.ErrorContext(r.Context(), "error saving flash", slog.String("error", err.Error()))
	}
	http.Redirect(w, r, to, http.StatusSeeOther)
}

func (a *App) success(w http.ResponseWriter, r *http.Request, to, msg string) {
	a.redirectWithFlash(w, r, to, flashSuccess, msg)
}

func (a *App) failure(w http.ResponseWriter, r *http.Request, to, msg string, err error) {
	a.log.ErrorContext(r.Context(), msg, slog.String("path", r.URL.Path), slog.String("error", err.Error()))
	a.redirectWithFlash(w, r, to, flashError, msg+": "+err.Error())
}

func (a *App) popFlashes(w http.ResponseWriter, r *http.Request) (success, errMsg string) {
	session, err := a.sessions.Get(r, sessionName)
	if err != nil {
		return "", ""
	}
	s, e := session.Flashes(flashSuccess), session.Flashes(flashError)
	if len(s) == 0 && len(e) == 0 {
		return "", ""
	}
	if err := session.Save(r, w); err != nil {
		a.log.ErrorContext(r.Context(), "error clearing flashes", slog.String("error", err.Error()))
	}
	return lastString(s), lastString(e)
}

func lastString(values []any) string {
	if len(values) == 0 {
		return ""
	}
	s, _ := values[len(values)-1].(string)
	return s
}
