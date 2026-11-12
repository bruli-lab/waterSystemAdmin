package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bruli/waterSystemAdmin/internal/domain/programs"
)

type WeeklyRepository struct {
	cl *Client
}

func (w WeeklyRepository) Remove(ctx context.Context, day *programs.WeekDay) error {
	url := fmt.Sprintf("http://%s/programs/weekly/%s", w.cl.apiURL, strings.ToLower(day.String()))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request to delete weekly program: %w", err)
	}
	req.Header.Add("Authorization", w.cl.token)
	resp, err := w.cl.cl.Do(req)
	if err != nil {
		return fmt.Errorf("failed sending the request to delete weekly program: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete weekly program. Request not accepted: %v. Error: %s", resp.StatusCode, string(data))
	}
	return nil
}

func (w WeeklyRepository) Save(ctx context.Context, p *programs.Weekly) error {
	url := fmt.Sprintf("http://%s/programs/weekly", w.cl.apiURL)

	body, _ := json.Marshal(w.buildProgramDTO(p))
	var buff bytes.Buffer
	buff.Write(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buff)
	if err != nil {
		return fmt.Errorf("failed to create request to create weekly program: %w", err)
	}
	req.Header.Add("Authorization", w.cl.token)
	resp, err := w.cl.cl.Do(req)
	if err != nil {
		return fmt.Errorf("failed sending the request to create weekly program: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create weekly program. Request not accepted: %v. Error: %s", resp.StatusCode, string(data))
	}
	return nil
}

func (w WeeklyRepository) buildProgramDTO(p *programs.Weekly) ProgramWeekly {
	prgms := make([]Program, len(p.Programs))
	for i, program := range p.Programs {
		exec := make([]Execution, len(program.Executions))
		for n, execution := range program.Executions {
			exec[n] = Execution{
				Seconds: execution.Seconds,
				Zones:   execution.Zones,
			}
		}
		prgms[i] = Program{
			Executions: exec,
			Hour:       program.Hour.String(),
		}
	}
	return ProgramWeekly{
		WeekDay:  p.WeekDay.String(),
		Programs: prgms,
	}
}

func NewWeeklyRepository(cl *Client) *WeeklyRepository {
	return &WeeklyRepository{cl: cl}
}
