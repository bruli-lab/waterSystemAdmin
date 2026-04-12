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

type TemperatureRepository struct {
	cl *Client
}

func (w TemperatureRepository) Remove(ctx context.Context, temperature int) error {
	url := fmt.Sprintf("http://%s/programs/temperature/%v", w.cl.apiURL, temperature)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request to delete temperature program: %w", err)
	}
	req.Header.Add("Authorization", w.cl.token)
	resp, err := w.cl.cl.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		return fmt.Errorf("failed sending the request to delete temperature program: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete temperature program. Request not accepted: %v. Error: %s", resp.StatusCode, string(data))
	}
	return nil
}

func (w TemperatureRepository) Save(ctx context.Context, p *programs.TemperatureProgram) error {
	url := fmt.Sprintf("http://%s/programs/temperature", w.cl.apiURL)

	body, _ := json.Marshal(w.buildTemperatureProgram(p))
	var buff bytes.Buffer
	buff.Write(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buff)
	if err != nil {
		return fmt.Errorf("failed to create request to create temperature program: %w", err)
	}
	req.Header.Add("Authorization", w.cl.token)
	resp, err := w.cl.cl.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		return fmt.Errorf("failed sending the request to create temperature program: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create temperature program. Request not accepted: %v. Error: %s", resp.StatusCode, string(data))
	}
	return nil
}

func (w TemperatureRepository) buildTemperatureProgram(p *programs.TemperatureProgram) TemperatureProgram {
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
	return TemperatureProgram{
		Programs:    prgms,
		Temperature: p.Temperature,
	}
}

func NewTemperatureRepository(cl *Client) *TemperatureRepository {
	return &TemperatureRepository{cl: cl}
}
