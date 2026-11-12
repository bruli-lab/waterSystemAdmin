package api

import (
	"context"
	"net/http"

	"github.com/bruli-lab/waterSystemAdmin/internal/model/status"
	"github.com/bruli-lab/waterSystemAdmin/internal/model/vo"
)

type StatusRepository struct {
	cl *Client
}

func (s *StatusRepository) Find(ctx context.Context) (*status.Status, error) {
	var st Status
	if err := s.cl.do(ctx, http.MethodGet, "/status", nil, &st); err != nil {
		return nil, err
	}
	return s.buildStatus(&st)
}

func (s *StatusRepository) Update(ctx context.Context) error {
	return s.cl.do(ctx, http.MethodGet, "/weather", nil, nil)
}

func (s *StatusRepository) buildStatus(st *Status) (*status.Status, error) {
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
