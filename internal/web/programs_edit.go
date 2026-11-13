package web

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	"github.com/bruli-lab/waterSystemAdmin/internal/model/programs"
)

func (a *App) loadPrograms(w http.ResponseWriter, r *http.Request, section string) (*programs.Programs, bool) {
	all, err := a.api.Programs.FindAll(r.Context())
	if err != nil {
		a.failure(w, r, programsURL(section), "Could not read the programs", err)
		return nil, false
	}
	return all, true
}

func fertilizerPrograms(all *programs.Programs, t programs.TypeProgram) []programs.FertilizerProgram {
	switch t {
	case programs.OddProgramType:
		return all.Odd
	case programs.EvenProgramType:
		return all.Even
	default:
		return all.Daily
	}
}

func toExecInputs(execs []programs.Execution) []execInput {
	result := make([]execInput, len(execs))
	for i, e := range execs {
		result[i] = execInput{Seconds: e.Seconds, Zones: e.Zones}
	}
	return result
}

func toHourInputs(prgms []programs.Program) []hourInput {
	result := make([]hourInput, len(prgms))
	for i, p := range prgms {
		result[i] = hourInput{Hour: p.Hour.String(), Executions: toExecInputs(p.Executions)}
	}
	return result
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request, section, what string) {
	a.redirectWithFlash(w, r, programsURL(section), flashError, what+" does not exist.")
}

func (a *App) EditProgram(w http.ResponseWriter, r *http.Request) {
	progType, err := programs.ParseProgramType(r.PathValue("type"))
	if err != nil {
		a.failure(w, r, "/programs", "Invalid program type", err)
		return
	}
	section := progType.String()
	hour, err := programs.ParseHour(r.PathValue("hour"))
	if err != nil {
		a.failure(w, r, programsURL(section), "Invalid hour", err)
		return
	}
	all, ok := a.loadPrograms(w, r, section)
	if !ok {
		return
	}
	list := fertilizerPrograms(all, *progType)
	i := slices.IndexFunc(list, func(p programs.FertilizerProgram) bool { return p.Hour.String() == hour.String() })
	if i < 0 {
		a.notFound(w, r, section, fmt.Sprintf("The %s program at %s", section, hour))
		return
	}
	form := &programForm{
		Kind:       section,
		Action:     fmt.Sprintf("/programs/edit/%s/%s", section, hour),
		Key:        hour.String(),
		Fertilizer: list[i].Fertilizer,
		Executions: toExecInputs(list[i].Executions),
	}
	v, ok := a.formView(w, r, "program_form.html", form)
	if !ok {
		return
	}
	v.Title = fmt.Sprintf("Edit %s program at %s", section, hour)
	if r.Method != http.MethodPost {
		a.render(w, r, &v)
		return
	}
	p, err := parseFertilizerProgram(r, hour.String())
	if err == nil {
		err = a.api.Programs.Update(r.Context(), p, progType)
	}
	if err != nil {
		a.log.ErrorContext(r.Context(), "error updating program", slog.String("error", err.Error()))
		v.fail("Could not update the program", err)
		a.render(w, r, &v)
		return
	}
	a.success(w, r, programsURL(section), fmt.Sprintf("%s program at %s updated.", progType, hour))
}

func (a *App) EditWeeklyProgram(w http.ResponseWriter, r *http.Request) {
	day, err := programs.ParseWeekDay(r.PathValue("weekday"))
	if err != nil {
		a.failure(w, r, programsURL("weekly"), "Invalid day", err)
		return
	}
	all, ok := a.loadPrograms(w, r, "weekly")
	if !ok {
		return
	}
	i := slices.IndexFunc(all.Weekly, func(p programs.Weekly) bool { return p.WeekDay == day })
	if i < 0 {
		a.notFound(w, r, "weekly", fmt.Sprintf("The weekly program for %s", day))
		return
	}
	form := &programForm{
		Kind:   "weekly",
		Action: "/programs/weekly/edit/" + day.String(),
		Key:    day.String(),
		Hours:  toHourInputs(all.Weekly[i].Programs),
	}
	v, ok := a.formView(w, r, "program_hours_form.html", form)
	if !ok {
		return
	}
	v.Title = "Edit weekly program for " + day.String()
	if r.Method != http.MethodPost {
		a.render(w, r, &v)
		return
	}
	prgms, err := parseHourPrograms(r)
	if err == nil {
		err = a.api.Weekly.Update(r.Context(), &programs.Weekly{WeekDay: day, Programs: prgms})
	}
	if err != nil {
		a.log.ErrorContext(r.Context(), "error updating weekly program", slog.String("error", err.Error()))
		v.fail("Could not update the program", err)
		a.render(w, r, &v)
		return
	}
	a.success(w, r, programsURL("weekly"), fmt.Sprintf("Weekly program for %s updated.", day))
}

func (a *App) EditTemperatureProgram(w http.ResponseWriter, r *http.Request) {
	temp, err := strconv.Atoi(r.PathValue("temperature"))
	if err != nil {
		a.failure(w, r, programsURL("temperature"), "Invalid temperature", err)
		return
	}
	all, ok := a.loadPrograms(w, r, "temperature")
	if !ok {
		return
	}
	i := slices.IndexFunc(all.Temperature, func(p programs.TemperatureProgram) bool { return p.Temperature == temp })
	if i < 0 {
		a.notFound(w, r, "temperature", fmt.Sprintf("The program for %d °C", temp))
		return
	}
	form := &programForm{
		Kind:   "temperature",
		Action: "/programs/temperature/edit/" + strconv.Itoa(temp),
		Key:    strconv.Itoa(temp),
		Hours:  toHourInputs(all.Temperature[i].Programs),
	}
	v, ok := a.formView(w, r, "program_hours_form.html", form)
	if !ok {
		return
	}
	v.Title = fmt.Sprintf("Edit program for %d °C", temp)
	if r.Method != http.MethodPost {
		a.render(w, r, &v)
		return
	}
	prgms, err := parseHourPrograms(r)
	if err == nil {
		err = a.api.Temperature.Update(r.Context(), &programs.TemperatureProgram{Temperature: temp, Programs: prgms})
	}
	if err != nil {
		a.log.ErrorContext(r.Context(), "error updating temperature program", slog.String("error", err.Error()))
		v.fail("Could not update the program", err)
		a.render(w, r, &v)
		return
	}
	a.success(w, r, programsURL("temperature"), fmt.Sprintf("Program for %d °C updated.", temp))
}
