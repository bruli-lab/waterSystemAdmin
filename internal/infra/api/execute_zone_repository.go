package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/execution"
)

type ExecuteZoneRepository struct {
	cl *Client
}

func (e ExecuteZoneRepository) SendClean(ctx context.Context, id string) error {
	url := fmt.Sprintf("http://%s/zones/clean/%s/execute", e.cl.apiURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to execute clean zone: %w", err)
	}
	req.Header.Add("Authorization", e.cl.token)
	resp, err := e.cl.cl.Do(req)
	if err != nil {
		return fmt.Errorf("failed doing the request to execute clean zone: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to execute clean zone. Request not accepted: status code %d", resp.StatusCode)
	}
	return nil
}

func (e ExecuteZoneRepository) SendFertilizer(ctx context.Context, exe *execution.Execution) error {
	url := fmt.Sprintf("http://%s/zones/fertilizer/%s/execute", e.cl.apiURL, exe.ID)
	return e.execution(ctx, exe, url)
}

func (e ExecuteZoneRepository) execution(ctx context.Context, exe *execution.Execution, url string) error {
	seconds := ExecuteZone{Seconds: exe.Seconds}
	body, _ := json.Marshal(seconds)
	var buff bytes.Buffer
	buff.Write(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buff)
	if err != nil {
		return fmt.Errorf("failed to execute zone: %w", err)
	}
	req.Header.Add("Authorization", e.cl.token)
	resp, err := e.cl.cl.Do(req)
	if err != nil {
		return fmt.Errorf("failed doing the request to execute zone: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to execute zone. Request not accepted: status code %d", resp.StatusCode)
	}
	return nil
}

func (e ExecuteZoneRepository) SendExecution(ctx context.Context, exe *execution.Execution) error {
	url := fmt.Sprintf("http://%s/zones/%s/execute", e.cl.apiURL, exe.ID)
	return e.execution(ctx, exe, url)
}

func NewExecuteZoneRepository(cl *Client) *ExecuteZoneRepository {
	return &ExecuteZoneRepository{cl: cl}
}
