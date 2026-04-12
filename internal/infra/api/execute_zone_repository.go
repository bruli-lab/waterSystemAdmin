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

func (e ExecuteZoneRepository) SendExecution(ctx context.Context, exe *execution.Execution) error {
	url := fmt.Sprintf("http://%s/zones/%s/execute", e.cl.apiURL, exe.ID)
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
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		return fmt.Errorf("failed doing the request to execute zone: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to execute zone. Reques no accepted: %w", err)
	}
	return nil
}

func NewExecuteZoneRepository(cl *Client) *ExecuteZoneRepository {
	return &ExecuteZoneRepository{cl: cl}
}
