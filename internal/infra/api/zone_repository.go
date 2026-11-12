package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/zones"
)

type ZoneRepository struct {
	cl *Client
}

func (r ZoneRepository) Delete(ctx context.Context, id string) error {
	url := fmt.Sprintf("http://%s/zones/%s", r.cl.apiURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}
	req.Header.Add("Authorization", r.cl.token)
	resp, err := r.cl.cl.Do(req)
	if err != nil {
		return fmt.Errorf("failed doing the request to delete zone: %s", err.Error())
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete zone. Request no accepted: Status code %v", resp.StatusCode)
	}
	return nil
}

func (r ZoneRepository) FindAll(ctx context.Context) ([]zones.Zone, error) {
	url := fmt.Sprintf("http://%s/zones", r.cl.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to find zones: %w", err)
	}
	req.Header.Add("Authorization", r.cl.token)
	resp, err := r.cl.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed doing the request to find zones: %s", err.Error())
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to find zones. Request no accepted: Status code %v", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var logsData []Zone
	if err = json.Unmarshal(data, &logsData); err != nil {
		return nil, fmt.Errorf("failed to unmarshall zones: %w", err)
	}
	return r.buildZones(logsData), nil
}

func (r ZoneRepository) Create(ctx context.Context, z *zones.Zone) error {
	url := fmt.Sprintf("http://%s/zones", r.cl.apiURL)
	zo := Zone{
		ID:     z.ID,
		Name:   z.Name,
		Relays: z.Relays,
	}
	body, _ := json.Marshal(zo)
	var buff bytes.Buffer
	buff.Write(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buff)
	if err != nil {
		return fmt.Errorf("failed to create zone: %w", err)
	}
	req.Header.Add("Authorization", r.cl.token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := r.cl.cl.Do(req)
	if err != nil {
		return fmt.Errorf("failed doing the request to create zone: %s", err.Error())
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			msg, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to create zone. Request no accepted: %s", string(msg))
		}
		return fmt.Errorf("failed to create zone. Request no accepted: Status code %v", resp.StatusCode)
	}
	return nil
}

func (r ZoneRepository) Update(ctx context.Context, z *zones.Zone) error {
	url := fmt.Sprintf("http://%s/zones/%s", r.cl.apiURL, z.ID)
	zo := UpdateZone{
		Name:   z.Name,
		Relays: z.Relays,
	}
	body, _ := json.Marshal(zo)
	var buff bytes.Buffer
	buff.Write(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &buff)
	if err != nil {
		return fmt.Errorf("failed to update zone: %w", err)
	}
	req.Header.Add("Authorization", r.cl.token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := r.cl.cl.Do(req)
	if err != nil {
		return fmt.Errorf("failed doing the request to update zone: %s", err.Error())
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			msg, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to updated zone. Request no accepted: %s", string(msg))
		}
		return fmt.Errorf("failed to updated zone. Request no accepted: Status code %v", resp.StatusCode)
	}
	return nil
}

func (r ZoneRepository) buildZones(data []Zone) []zones.Zone {
	zons := make([]zones.Zone, len(data))
	for i, dat := range data {
		zons[i] = r.buildZone(dat)
	}
	return zons
}

func (r ZoneRepository) buildZone(dat Zone) zones.Zone {
	return zones.Zone{
		ID:     dat.ID,
		Name:   dat.Name,
		Relays: dat.Relays,
	}
}

func NewZoneRepository(cl *Client) *ZoneRepository {
	return &ZoneRepository{cl: cl}
}
