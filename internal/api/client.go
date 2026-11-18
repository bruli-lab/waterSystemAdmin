package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrUnavailable = errors.New("api unavailable")

type StatusError struct {
	Method, Path, Message string
	Code                  int
}

func (e *StatusError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("request %s %s: status code %d: %s", e.Method, e.Path, e.Code, e.Message)
	}
	return fmt.Sprintf("request %s %s: status code %d", e.Method, e.Path, e.Code)
}

type Client struct {
	cl     *http.Client
	token  string
	apiURL string
}

func NewClient(token, apiURL string, timeout time.Duration) *Client {
	return &Client{token: token, cl: &http.Client{Timeout: timeout}, apiURL: apiURL}
}

func (c *Client) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader = http.NoBody
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("encoding request %s %s: %w", method, path, err)
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://"+c.apiURL+path, body)
	if err != nil {
		return fmt.Errorf("building request %s %s: %w", method, path, err)
	}
	req.Header.Set("Authorization", c.token)
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.cl.Do(req)
	if err != nil {
		return fmt.Errorf("%w: request %s %s: %w", ErrUnavailable, method, path, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return &StatusError{Method: method, Path: path, Code: resp.StatusCode, Message: strings.TrimSpace(string(data))}
	}
	if out == nil {
		return nil
	}
	if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decoding response %s %s: %w", method, path, err)
	}
	return nil
}
