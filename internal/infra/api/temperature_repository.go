package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bruli/waterSystemAdmin/internal/domain/programs"
)

type TemperatureRepository struct {
	cl *Client
}

func (t TemperatureRepository) Remove(ctx context.Context, temperature int) error {
	return t.cl.do(ctx, http.MethodDelete, "/programs/temperature/"+strconv.Itoa(temperature), nil, nil)
}

func (t TemperatureRepository) Save(ctx context.Context, p *programs.TemperatureProgram) error {
	dto := TemperatureProgram{Programs: toProgramDTOs(p.Programs), Temperature: p.Temperature}
	return t.cl.do(ctx, http.MethodPost, "/programs/temperature", dto, nil)
}

func NewTemperatureRepository(cl *Client) *TemperatureRepository {
	return &TemperatureRepository{cl: cl}
}
