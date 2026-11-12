package api

import (
	"context"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/logs"
	"github.com/bruli/waterSystemAdmin/internal/domain/vo"
)

type LogRepository struct {
	cl *Client
}

func (l LogRepository) Find(ctx context.Context) ([]logs.Log, error) {
	var data []Log
	if err := l.cl.do(ctx, http.MethodGet, "/logs", nil, &data); err != nil {
		return nil, err
	}
	result := make([]logs.Log, len(data))
	for i, d := range data {
		executedAt, err := vo.ParseTimeFromUnix(d.ExecutedAt)
		if err != nil {
			return nil, err
		}
		result[i] = logs.Log{
			ExecutedAt: executedAt,
			Seconds:    d.Seconds,
			Zone:       d.ZoneName,
		}
	}
	return result, nil
}

func NewLogRepository(cl *Client) *LogRepository {
	return &LogRepository{cl: cl}
}
