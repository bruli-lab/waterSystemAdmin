package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/bruli/waterSystemAdmin/internal/domain/programs"
)

type WeeklyRepository struct {
	cl *Client
}

func (w WeeklyRepository) Remove(ctx context.Context, day *programs.WeekDay) error {
	return w.cl.do(ctx, http.MethodDelete, "/programs/weekly/"+strings.ToLower(day.String()), nil, nil)
}

func (w WeeklyRepository) Save(ctx context.Context, p *programs.Weekly) error {
	dto := ProgramWeekly{WeekDay: p.WeekDay.String(), Programs: toProgramDTOs(p.Programs)}
	return w.cl.do(ctx, http.MethodPost, "/programs/weekly", dto, nil)
}

func NewWeeklyRepository(cl *Client) *WeeklyRepository {
	return &WeeklyRepository{cl: cl}
}
