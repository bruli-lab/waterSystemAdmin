package web

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/bruli-lab/waterSystemAdmin/internal/api"
	"github.com/bruli-lab/waterSystemAdmin/internal/disk"
	"github.com/bruli-lab/waterSystemAdmin/internal/model/programs"
	"github.com/bruli-lab/waterSystemAdmin/internal/model/status"
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
	pageRelays   = "relays"
	sessionName  = "session"
	flashSuccess = "flash_success"
	flashError   = "flash_error"
	maxSeconds   = 120
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
	Relays      *api.RelayRepository
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

type formButtons struct {
	Back     *link
	Label    string
	Disabled bool
}

type navLink struct {
	Active            bool
	Href, Icon, Label string
}

var assetVersions = func() map[string]string {
	versions := map[string]string{}
	err := fs.WalkDir(staticFS, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := staticFS.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		versions[strings.TrimPrefix(path, "static/")] = hex.EncodeToString(sum[:4])
		return nil
	})
	if err != nil {
		panic(err)
	}
	return versions
}()

func asset(path string) string {
	return "/static/" + path + "?v=" + assetVersions[path]
}

var templateFuncs = template.FuncMap{
	"asset": asset,
	"navItem": func(current, page, href, icon, label string) navLink {
		return navLink{Active: current == page, Href: href, Icon: icon, Label: label}
	},
	"join":       strings.Join,
	"lower":      strings.ToLower,
	"contains":   slices.Contains[[]int],
	"maxSeconds": func() int { return maxSeconds },
	"buttons": func(back *link, label string, disabled bool) formButtons {
		return formButtons{Back: back, Label: label, Disabled: disabled}
	},
	"menu": func(edit, del, label string) rowMenu {
		return rowMenu{Edit: edit, Delete: del, Label: label}
	},
	"relayItem": func(r relay, cleaning bool) relayView {
		return relayView{Relay: r, Cleaning: cleaning}
	},
	"named": func(v any, names map[string]string) named {
		return named{V: v, Names: names}
	},
	"percent":   percent,
	"dayOffset": dayOffset,
}

type named struct {
	V     any
	Names map[string]string
}

type rowMenu struct {
	Edit, Delete, Label string
}

type relayView struct {
	Relay    relay
	Cleaning bool
}

func percent(value, maxValue float64) float64 {
	return min(max(value/maxValue*100, 0), 100)
}

func dayOffset(h programs.Hour) float64 {
	t := time.Time(h)
	return float64(t.Hour()*60+t.Minute()) / (24 * 60) * 100
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

type link struct {
	Href, Label string
}

type pageData struct {
	Page    string
	Title   string
	Back    *link
	Status  *status.Status
	Error   string
	Success string
	Refresh bool
	Data    any
}

func (p *pageData) Dark() bool { return p.Status == nil || !p.Status.IsDay }

func (p *pageData) Cleaning() bool { return p.Status != nil && p.Status.Clean }

type view struct {
	Template    string
	Page        string
	Title       string
	Back        *link
	Root        string
	Data        any
	Error       string
	QuietStatus bool
	Offline     bool
	Refresh     bool
}

func (v *view) fail(msg string, err error) {
	v.Error = errMessage(msg, err)
	v.Offline = errors.Is(err, api.ErrUnavailable)
}

func errMessage(msg string, err error) string {
	if errors.Is(err, api.ErrUnavailable) {
		return msg + ": the irrigation API is not reachable."
	}
	return msg + ": " + err.Error()
}

func (a *App) loadStatus(r *http.Request, pd *pageData, quiet bool) {
	st, err := a.api.Status.Find(r.Context())
	if err != nil {
		a.log.ErrorContext(r.Context(), "error reading status", slog.String("error", err.Error()))
		if pd.Error == "" && !quiet {
			pd.Error = errMessage("Could not read the system status", err)
		}
		return
	}
	pd.Status = st
}

func (a *App) render(w http.ResponseWriter, r *http.Request, v *view) {
	ctx := r.Context()
	pd := pageData{Page: v.Page, Title: v.Title, Back: v.Back, Data: v.Data, Error: v.Error, Refresh: v.Refresh}

	success, flashErr := a.popFlashes(w, r)
	pd.Success = success
	if pd.Error == "" {
		pd.Error = flashErr
	}

	if !v.Offline {
		a.loadStatus(r, &pd, v.QuietStatus)
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
	if err := tpl.ExecuteTemplate(&buf, root, &pd); err != nil {
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
	http.Redirect(w, r, to, http.StatusSeeOther) //nolint:gosec // targets are internal paths built from validated values
}

func (a *App) success(w http.ResponseWriter, r *http.Request, to, msg string) {
	a.redirectWithFlash(w, r, to, flashSuccess, msg)
}

func (a *App) failure(w http.ResponseWriter, r *http.Request, to, msg string, err error) {
	a.log.ErrorContext(r.Context(), msg, slog.String("path", r.URL.Path), slog.String("error", err.Error()))
	a.redirectWithFlash(w, r, to, flashError, errMessage(msg, err))
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
