package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bruli/waterSystemAdmin/internal/domain/logs"
	"github.com/bruli/waterSystemAdmin/internal/domain/vo"
)

type LogRepository struct {
	cl *Client
}

func (l LogRepository) Find(ctx context.Context) ([]logs.Log, error) {
	url := fmt.Sprintf("http://%s/logs", l.cl.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to read logs: %w", err)
	}
	req.Header.Add("Authorization", l.cl.token)
	resp, err := l.cl.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed sending the request to read logs: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to read logs. Request not accepted: %v", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var logsData []Log
	if err = json.Unmarshal(data, &logsData); err != nil {
		return nil, fmt.Errorf("failed to unmarshall logs: %w", err)
	}
	return l.buildLogs(logsData)
}

func (l LogRepository) buildLogs(data []Log) ([]logs.Log, error) {
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
