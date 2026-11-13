package api

import (
	"context"
	"net/http"
	"net/url"

	"github.com/bruli-lab/waterSystemAdmin/internal/model/zones"
)

type ExecuteZoneRepository struct {
	cl *Client
}

func (e ExecuteZoneRepository) SendClean(ctx context.Context, id string) error {
	return e.cl.do(ctx, http.MethodPost, "/zones/clean/"+url.PathEscape(id)+"/execute", nil, nil)
}

func (e ExecuteZoneRepository) SendFertilizer(ctx context.Context, exe *zones.Execution) error {
	return e.cl.do(ctx, http.MethodPost, "/zones/fertilizer/"+url.PathEscape(exe.ID)+"/execute", ExecuteZone{Seconds: exe.Seconds}, nil)
}

func (e ExecuteZoneRepository) SendExecution(ctx context.Context, exe *zones.Execution) error {
	return e.cl.do(ctx, http.MethodPost, "/zones/"+url.PathEscape(exe.ID)+"/execute", ExecuteZone{Seconds: exe.Seconds}, nil)
}

func NewExecuteZoneRepository(cl *Client) *ExecuteZoneRepository {
	return &ExecuteZoneRepository{cl: cl}
}
