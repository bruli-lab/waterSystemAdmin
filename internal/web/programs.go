package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/bruli-lab/waterSystemAdmin/internal/model/programs"
	"github.com/bruli-lab/waterSystemAdmin/internal/model/zones"
)

type programSection struct {
	Title, Type string
	Programs    []programs.FertilizerProgram
}

type programsPage struct {
	Sections    []programSection
	Weekly      []programs.Weekly
	Temperature programs.Temperature
	Loaded      bool
}

func (a *App) Programs(w http.ResponseWriter, r *http.Request) {
	v := view{Template: "programs.html", Page: pagePrograms, Title: "Programs"}
	all, err := a.api.Programs.FindAll(r.Context())
	if err != nil {
		a.log.ErrorContext(r.Context(), "error finding programs", slog.String("error", err.Error()))
		v.fail("Could not read the programs", err)
		v.Data = programsPage{}
		a.render(w, r, &v)
		return
	}
	v.Data = programsPage{
		Sections: []programSection{
			{Title: "Daily", Type: programs.DailyProgramType.String(), Programs: all.Daily},
			{Title: "Odd", Type: programs.OddProgramType.String(), Programs: all.Odd},
			{Title: "Even", Type: programs.EvenProgramType.String(), Programs: all.Even},
		},
		Weekly:      all.Weekly,
		Temperature: all.Temperature,
		Loaded:      true,
	}
	a.render(w, r, &v)
}

type execInput struct {
	Seconds int      `json:"seconds"`
	Zones   []string `json:"zones"`
}

type hourInput struct {
	Hour       string      `json:"hour"`
	Executions []execInput `json:"executions"`
}

type programForm struct {
	Kind       string
	Zones      []zones.Zone
	Action     string
	Key        string
	Fertilizer bool
	Executions []execInput
	Hours      []hourInput
}

func programsURL(section string) string {
	return "/programs#" + section
}

func (a *App) formView(w http.ResponseWriter, r *http.Request, template string, form *programForm) (view, bool) {
	list, err := a.api.Zones.FindAll(r.Context())
	if err != nil {
		a.failure(w, r, programsURL(form.Kind), "Could not read the zones", err)
		return view{}, false
	}
	form.Zones = list
	v := view{
		Template: template,
		Page:     pagePrograms,
		Title:    "New " + form.Kind + " program",
		Back:     &link{Href: programsURL(form.Kind), Label: "Programs"},
		Data:     form,
	}
	if len(list) == 0 {
		v.Error = "There are no zones configured."
	}
	return v, true
}

func (a *App) CreateProgram(w http.ResponseWriter, r *http.Request) {
	progType, err := programs.ParseProgramType(r.URL.Query().Get("type"))
	if err != nil {
		a.failure(w, r, "/programs", "Invalid program type", err)
		return
	}
	form := &programForm{Kind: progType.String(), Action: "/programs/new?type=" + progType.String()}
	v, ok := a.formView(w, r, "program_form.html", form)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		a.render(w, r, &v)
		return
	}
	p, err := parseFertilizerProgram(r, r.FormValue("hour"))
	if err == nil {
		err = a.api.Programs.Save(r.Context(), p, progType)
	}
	if err != nil {
		a.log.ErrorContext(r.Context(), "error creating program", slog.String("error", err.Error()))
		v.fail("Could not create the program", err)
		a.render(w, r, &v)
		return
	}
	a.success(w, r, programsURL(progType.String()), fmt.Sprintf("%s program at %s created.", progType, p.Hour))
}

func (a *App) CreateWeeklyProgram(w http.ResponseWriter, r *http.Request) {
	v, ok := a.formView(w, r, "program_hours_form.html", &programForm{Kind: "weekly", Action: "/programs/weekly/new"})
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		a.render(w, r, &v)
		return
	}
	day, err := programs.ParseWeekDay(r.FormValue("day"))
	var prgms []programs.Program
	if err == nil {
		prgms, err = parseHourPrograms(r)
	}
	if err == nil {
		err = a.api.Weekly.Save(r.Context(), &programs.Weekly{WeekDay: day, Programs: prgms})
	}
	if err != nil {
		a.log.ErrorContext(r.Context(), "error creating weekly program", slog.String("error", err.Error()))
		v.fail("Could not create the program", err)
		a.render(w, r, &v)
		return
	}
	a.success(w, r, programsURL("weekly"), fmt.Sprintf("Weekly program for %s created.", day))
}

func (a *App) CreateTemperatureProgram(w http.ResponseWriter, r *http.Request) {
	v, ok := a.formView(w, r, "program_hours_form.html", &programForm{Kind: "temperature", Action: "/programs/temperature/new"})
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		a.render(w, r, &v)
		return
	}
	temp, err := strconv.Atoi(r.FormValue("temperature"))
	var prgms []programs.Program
	if err != nil {
		err = errors.New("invalid temperature")
	} else {
		prgms, err = parseHourPrograms(r)
	}
	if err == nil {
		err = a.api.Temperature.Save(r.Context(), &programs.TemperatureProgram{Temperature: temp, Programs: prgms})
	}
	if err != nil {
		a.log.ErrorContext(r.Context(), "error creating temperature program", slog.String("error", err.Error()))
		v.fail("Could not create the program", err)
		a.render(w, r, &v)
		return
	}
	a.success(w, r, programsURL("temperature"), fmt.Sprintf("Program for %d °C created.", temp))
}

func (a *App) RemoveProgram(w http.ResponseWriter, r *http.Request) {
	hour, err := programs.ParseHour(r.PathValue("hour"))
	if err != nil {
		a.failure(w, r, "/programs", "Invalid hour", err)
		return
	}
	progType, err := programs.ParseProgramType(r.PathValue("type"))
	if err != nil {
		a.failure(w, r, "/programs", "Invalid program type", err)
		return
	}
	if err = a.api.Programs.Remove(r.Context(), &hour, progType); err != nil {
		a.failure(w, r, programsURL(progType.String()), "Could not delete the program", err)
		return
	}
	a.success(w, r, programsURL(progType.String()), fmt.Sprintf("%s program at %s deleted.", progType, hour))
}

func (a *App) RemoveWeeklyProgram(w http.ResponseWriter, r *http.Request) {
	day, err := programs.ParseWeekDay(r.PathValue("weekday"))
	if err != nil {
		a.failure(w, r, programsURL("weekly"), "Invalid day", err)
		return
	}
	if err = a.api.Weekly.Remove(r.Context(), &day); err != nil {
		a.failure(w, r, programsURL("weekly"), "Could not delete the program", err)
		return
	}
	a.success(w, r, programsURL("weekly"), fmt.Sprintf("Weekly program for %s deleted.", day))
}

func (a *App) RemoveTemperatureProgram(w http.ResponseWriter, r *http.Request) {
	temp, err := strconv.Atoi(r.PathValue("temperature"))
	if err != nil {
		a.failure(w, r, programsURL("temperature"), "Invalid temperature", err)
		return
	}
	if err = a.api.Temperature.Remove(r.Context(), temp); err != nil {
		a.failure(w, r, programsURL("temperature"), "Could not delete the program", err)
		return
	}
	a.success(w, r, programsURL("temperature"), fmt.Sprintf("Program for %d °C deleted.", temp))
}

func parseFertilizerProgram(r *http.Request, hourValue string) (*programs.FertilizerProgram, error) {
	if err := r.ParseForm(); err != nil {
		return nil, errors.New("invalid form")
	}
	hour, err := programs.ParseHour(hourValue)
	if err != nil {
		return nil, errors.New("invalid hour")
	}
	fertilizer, _ := strconv.ParseBool(r.FormValue("fertilizer"))
	execs, err := parseExecutions(r.Form["seconds[]"], func(i int) []string {
		return r.Form[fmt.Sprintf("zones_%d[]", i)]
	})
	if err != nil {
		return nil, err
	}
	return &programs.FertilizerProgram{Executions: execs, Hour: hour, Fertilizer: fertilizer}, nil
}

func parseHourPrograms(r *http.Request) ([]programs.Program, error) {
	if err := r.ParseForm(); err != nil {
		return nil, errors.New("invalid form")
	}
	hours := r.Form["hours[]"]
	if len(hours) == 0 {
		return nil, errors.New("add at least one hour")
	}
	prgms := make([]programs.Program, len(hours))
	for h, hourValue := range hours {
		hour, err := programs.ParseHour(hourValue)
		if err != nil {
			return nil, fmt.Errorf("invalid hour: %q", hourValue)
		}
		execs, err := parseExecutions(r.Form[fmt.Sprintf("executions_%d_seconds[]", h)], func(n int) []string {
			return r.Form[fmt.Sprintf("executions_%d_zones_%d[]", h, n)]
		})
		if err != nil {
			return nil, fmt.Errorf("hour %s: %w", hourValue, err)
		}
		prgms[h] = programs.Program{Executions: execs, Hour: hour}
	}
	return prgms, nil
}

func parseExecutions(seconds []string, zonesOf func(i int) []string) ([]programs.Execution, error) {
	if len(seconds) == 0 {
		return nil, errors.New("add at least one execution")
	}
	execs := make([]programs.Execution, len(seconds))
	for i, sec := range seconds {
		s, err := parseSeconds(sec)
		if err != nil {
			return nil, fmt.Errorf("execution %d: %w", i+1, err)
		}
		zoneIDs := zonesOf(i)
		if len(zoneIDs) == 0 {
			return nil, fmt.Errorf("execution %d: select at least one zone", i+1)
		}
		execs[i] = programs.Execution{Seconds: s, Zones: zoneIDs}
	}
	return execs, nil
}

func parseSeconds(value string) (int, error) {
	s, err := strconv.Atoi(value)
	if err != nil || s < 1 || s > maxSeconds {
		return 0, fmt.Errorf("seconds must be between 1 and %d", maxSeconds)
	}
	return s, nil
}
