package api

import (
	"context"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/programs"
)

type ProgramsRepository struct {
	cl *Client
}

func (a ProgramsRepository) Remove(ctx context.Context, hour *programs.Hour, t *programs.TypeProgram) error {
	return a.cl.do(ctx, http.MethodDelete, "/programs/"+t.String()+"/"+hour.String(), nil, nil)
}

func (a ProgramsRepository) Save(ctx context.Context, p *programs.FertilizerProgram, t *programs.TypeProgram) error {
	dto := FertilizerProgram{
		Executions: toExecutionDTOs(p.Executions),
		Hour:       p.Hour.String(),
		Fertilizer: p.Fertilizer,
	}
	return a.cl.do(ctx, http.MethodPost, "/programs/"+t.String(), dto, nil)
}

func (a ProgramsRepository) FindAll(ctx context.Context) (*programs.Programs, error) {
	var prms Programs
	if err := a.cl.do(ctx, http.MethodGet, "/programs", nil, &prms); err != nil {
		return nil, err
	}
	return buildAllPrograms(&prms)
}

func buildAllPrograms(data *Programs) (*programs.Programs, error) {
	daily, err := buildFertilizerPrograms(data.Daily)
	if err != nil {
		return nil, err
	}
	odd, err := buildFertilizerPrograms(data.Odd)
	if err != nil {
		return nil, err
	}
	even, err := buildFertilizerPrograms(data.Even)
	if err != nil {
		return nil, err
	}
	temp, err := buildTemperaturePrograms(data.Temperature)
	if err != nil {
		return nil, err
	}
	weekly, err := buildWeeklyPrograms(data.Weekly)
	if err != nil {
		return nil, err
	}
	return &programs.Programs{
		Daily:       daily,
		Odd:         odd,
		Even:        even,
		Temperature: temp,
		Weekly:      weekly,
	}, nil
}

func buildExecutions(data []Execution) []programs.Execution {
	exec := make([]programs.Execution, len(data))
	for i, e := range data {
		exec[i] = programs.Execution{Seconds: e.Seconds, Zones: e.Zones}
	}
	return exec
}

func buildPrograms(data []Program) ([]programs.Program, error) {
	result := make([]programs.Program, len(data))
	for i, p := range data {
		hour, err := programs.ParseHour(p.Hour)
		if err != nil {
			return nil, err
		}
		result[i] = programs.Program{Executions: buildExecutions(p.Executions), Hour: hour}
	}
	return result, nil
}

func buildFertilizerPrograms(data []FertilizerProgram) ([]programs.FertilizerProgram, error) {
	result := make([]programs.FertilizerProgram, len(data))
	for i, p := range data {
		hour, err := programs.ParseHour(p.Hour)
		if err != nil {
			return nil, err
		}
		result[i] = programs.FertilizerProgram{
			Executions: buildExecutions(p.Executions),
			Hour:       hour,
			Fertilizer: p.Fertilizer,
		}
	}
	return result, nil
}

func buildTemperaturePrograms(data []TemperatureProgram) (programs.Temperature, error) {
	result := make(programs.Temperature, len(data))
	for i, t := range data {
		prgms, err := buildPrograms(t.Programs)
		if err != nil {
			return nil, err
		}
		result[i] = programs.TemperatureProgram{Programs: prgms, Temperature: t.Temperature}
	}
	return result, nil
}

func buildWeeklyPrograms(data []ProgramWeekly) ([]programs.Weekly, error) {
	result := make([]programs.Weekly, len(data))
	for i, w := range data {
		day, err := programs.ParseWeekDay(w.WeekDay)
		if err != nil {
			return nil, err
		}
		prgms, err := buildPrograms(w.Programs)
		if err != nil {
			return nil, err
		}
		result[i] = programs.Weekly{WeekDay: day, Programs: prgms}
	}
	return result, nil
}

func toExecutionDTOs(execs []programs.Execution) []Execution {
	result := make([]Execution, len(execs))
	for i, e := range execs {
		result[i] = Execution{Seconds: e.Seconds, Zones: e.Zones}
	}
	return result
}

func toProgramDTOs(prgms []programs.Program) []Program {
	result := make([]Program, len(prgms))
	for i, p := range prgms {
		result[i] = Program{Executions: toExecutionDTOs(p.Executions), Hour: p.Hour.String()}
	}
	return result
}

func NewAllProgramsRepository(cl *Client) *ProgramsRepository {
	return &ProgramsRepository{cl: cl}
}
