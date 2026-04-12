package api

import (
	"context"
	"fmt"
	"net/http"
)

type ActivateRepository struct {
	cl *Client
}

func (a ActivateRepository) Activate(ctx context.Context, activate bool) error {
	action := "activate"
	if !activate {
		action = "deactivate"
	}
	url := fmt.Sprintf("http://%s/status/%s", a.cl.apiURL, action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to execute %s action: %w", action, err)
	}
	req.Header.Add("Authorization", a.cl.token)
	resp, err := a.cl.cl.Do(req)
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		return fmt.Errorf("failed doing the request to execute %s acion: %w", action, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to execute %s action. Request no accepted: %w", action, err)
	}
	return nil
}

func NewActivateRepository(cl *Client) *ActivateRepository {
	return &ActivateRepository{cl: cl}
}
