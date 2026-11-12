package api

import (
	"context"
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
	return a.cl.do(ctx, http.MethodPatch, "/status/"+action, nil, nil)
}

func NewActivateRepository(cl *Client) *ActivateRepository {
	return &ActivateRepository{cl: cl}
}
