package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/programs"
)

type ProgramsRepository struct {
	cl *Client
}

func (a ProgramsRepository) Remove(ctx context.Context, hour *programs.Hour, t *programs.TypeProgram) error {
	url := fmt.Sprintf("http://%s/programs/%s/%s", a.cl.apiURL, t.String(), hour.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request to delete program: %w", err)
	}
	req.Header.Add("Authorization", a.cl.token)
	resp, err := a.cl.cl.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		return fmt.Errorf("failed sending the request to delete program: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete program. Request not accepted: %v", resp.StatusCode)
	}
	return nil
}

func (a ProgramsRepository) Save(ctx context.Context, p *programs.Program, t *programs.TypeProgram) error {
	url := fmt.Sprintf("http://%s/programs/%s", a.cl.apiURL, t.String())

	body, _ := json.Marshal(a.buildProgramDTO(p))
	var buff bytes.Buffer
	buff.Write(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buff)
	if err != nil {
		return fmt.Errorf("failed to create request to create program: %w", err)
	}
	req.Header.Add("Authorization", a.cl.token)
	resp, err := a.cl.cl.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		return fmt.Errorf("failed sending the request to create program: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create program. Request not accepted: %v", resp.StatusCode)
	}
	return nil
}

func (a ProgramsRepository) FindAll(ctx context.Context) (*programs.Programs, error) {
	url := fmt.Sprintf("http://%s/programs", a.cl.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to read programs: %w", err)
	}
	req.Header.Add("Authorization", a.cl.token)
	resp, err := a.cl.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed sending the request to read programs: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to read programs. Request not accepted: %v", resp.StatusCode)
	}
	var prms Programs
	data, _ := io.ReadAll(resp.Body)
	if err = json.Unmarshal(data, &prms); err != nil {
		return nil, fmt.Errorf("failed to unmarshall programs: %w", err)
	}
	return a.buildAllPrograms(&prms)
}

func (a ProgramsRepository) buildAllPrograms(data *Programs) (*programs.Programs, error) {
	daily, err := a.buildPrograms(data.Daily)
	if err != nil {
		return nil, err
	}
	temp, err := a.buildTemperaturePrograms(data.Temperature)
	if err != nil {
		return nil, err
	}
	odd, err := a.buildPrograms(data.Odd)
	if err != nil {
		return nil, err
	}
	even, err := a.buildPrograms(data.Even)
	if err != nil {
		return nil, err
	}
	weekly, err := a.buildWeeklyPrograms(data.Weekly)
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

func (a ProgramsRepository) buildPrograms(data []Program) ([]programs.Program, error) {
	daily := make(programs.Daily, len(data))
	for i, day := range data {
		hour, err := programs.ParseHour(day.Hour)
		if err != nil {
			return nil, err
		}
		exec := make([]programs.Execution, len(day.Executions))
		for j, e := range day.Executions {
			exec[j] = programs.Execution{
				Seconds: e.Seconds,
				Zones:   e.Zones,
			}
		}
		daily[i] = programs.Program{
			Executions: exec,
			Hour:       hour,
		}
	}
	return daily, nil
}

func (a ProgramsRepository) buildTemperaturePrograms(progs []TemperatureProgram) (programs.Temperature, error) {
	temp := make(programs.Temperature, len(progs))
	for i, prog := range progs {
		progm, err := a.buildPrograms(prog.Programs)
		if err != nil {
			return nil, err
		}
		temp[i] = programs.TemperatureProgram{
			Programs:    progm,
			Temperature: prog.Temperature,
		}
	}
	return temp, nil
}

func (a ProgramsRepository) buildWeeklyPrograms(progms []ProgramWeekly) ([]programs.Weekly, error) {
	week := make([]programs.Weekly, len(progms))
	for i, we := range progms {
		day, err := programs.ParseWeekDay(we.WeekDay)
		if err != nil {
			return nil, err
		}
		progm, err := a.buildPrograms(we.Programs)
		if err != nil {
			return nil, err
		}
		week[i] = programs.Weekly{
			WeekDay:  day,
			Programs: progm,
		}
	}
	return week, nil
}

func (a ProgramsRepository) buildProgramDTO(p *programs.Program) Program {
	exec := make([]Execution, len(p.Executions))
	for i, e := range p.Executions {
		exec[i] = Execution{
			Seconds: e.Seconds,
			Zones:   e.Zones,
		}
	}
	return Program{
		Executions: exec,
		Hour:       p.Hour.String(),
	}
}

func NewAllProgramsRepository(cl *Client) *ProgramsRepository {
	return &ProgramsRepository{cl: cl}
}
