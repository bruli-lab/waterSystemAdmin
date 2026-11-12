package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/status"
	"github.com/bruli/waterSystemAdmin/internal/domain/vo"
)

type StatusRepository struct {
	cl *Client
}

func (s *StatusRepository) Find(ctx context.Context) (*status.Status, error) {
	url := fmt.Sprintf("http://%s/status", s.cl.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to read status: %w", err)
	}
	req.Header.Add("Authorization", s.cl.token)
	resp, err := s.cl.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed doing the request to read status: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to read status. Request no accepted. Status code : %v", resp.StatusCode)
	}
	var st Status
	data, _ := io.ReadAll(resp.Body)
	if err = json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("failed to unmarshall status: %w", err)
	}
	return s.buildStatus(st)
}

func (s *StatusRepository) Update(ctx context.Context) error {
	url := fmt.Sprintf("http://%s/weather", s.cl.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	req.Header.Add("Authorization", s.cl.token)
	resp, err := s.cl.cl.Do(req)
	if err != nil {
		return fmt.Errorf("failed doing the request to update status: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update status. Request not accepted: status code %d", resp.StatusCode)
	}
	return nil
}

func (s *StatusRepository) buildStatus(st Status) (*status.Status, error) {
	start, err := vo.ParseTimeFromUnix(st.SystemStartedAt)
	if err != nil {
		return nil, err
	}
	updated := start
	if st.UpdatedAt != "" {
		updated, err = vo.ParseTimeFromUnix(st.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	return &status.Status{
		SystemStartedAt: start,
		Temperature:     st.Temperature,
		Humidity:        st.Humidity,
		IsRaining:       st.IsRaining,
		IsDay:           st.IsDay,
		UpdatedAt:       updated,
		Active:          st.Active,
		Clean:           st.Clean,
	}, nil
}

func NewStatusRepository(cl *Client) *StatusRepository {
	return &StatusRepository{cl: cl}
}
