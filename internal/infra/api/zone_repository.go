package api

import (
	"context"
	"net/http"
	"net/url"

	"github.com/bruli/waterSystemAdmin/internal/domain/zones"
)

type ZoneRepository struct {
	cl *Client
}

func (r ZoneRepository) Delete(ctx context.Context, id string) error {
	return r.cl.do(ctx, http.MethodDelete, "/zones/"+url.PathEscape(id), nil, nil)
}

func (r ZoneRepository) FindAll(ctx context.Context) ([]zones.Zone, error) {
	var data []Zone
	if err := r.cl.do(ctx, http.MethodGet, "/zones", nil, &data); err != nil {
		return nil, err
	}
	result := make([]zones.Zone, len(data))
	for i, z := range data {
		result[i] = zones.Zone{ID: z.ID, Name: z.Name, Relays: z.Relays}
	}
	return result, nil
}

func (r ZoneRepository) Create(ctx context.Context, z *zones.Zone) error {
	return r.cl.do(ctx, http.MethodPost, "/zones", Zone{ID: z.ID, Name: z.Name, Relays: z.Relays}, nil)
}

func (r ZoneRepository) Update(ctx context.Context, z *zones.Zone) error {
	return r.cl.do(ctx, http.MethodPut, "/zones/"+url.PathEscape(z.ID), UpdateZone{Name: z.Name, Relays: z.Relays}, nil)
}

func NewZoneRepository(cl *Client) *ZoneRepository {
	return &ZoneRepository{cl: cl}
}
