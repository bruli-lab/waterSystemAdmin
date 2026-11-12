package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bruli/waterSystemAdmin/internal/model/programs"
	"github.com/bruli/waterSystemAdmin/internal/model/zones"
)

const testToken = "secret-token"

type recorded struct {
	method, path, body, auth, contentType string
}

type fakeAPI struct {
	t        *testing.T
	status   int
	response string
	got      []recorded
}

func newFakeAPI(t *testing.T, status int, response string) (*fakeAPI, *Client) {
	t.Helper()
	f := &fakeAPI{t: t, status: status, response: response}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading body: %v", err)
		}
		f.got = append(f.got, recorded{
			method:      r.Method,
			path:        r.URL.EscapedPath(),
			body:        string(body),
			auth:        r.Header.Get("Authorization"),
			contentType: r.Header.Get("Content-Type"),
		})
		w.WriteHeader(f.status)
		_, _ = io.WriteString(w, f.response)
	}))
	t.Cleanup(srv.Close)
	return f, NewClient(testToken, strings.TrimPrefix(srv.URL, "http://"), time.Second)
}

func (f *fakeAPI) last() recorded {
	f.t.Helper()
	if len(f.got) != 1 {
		f.t.Fatalf("expected 1 request, got %d", len(f.got))
	}
	return f.got[0]
}

func mustHour(t *testing.T, s string) programs.Hour {
	t.Helper()
	h, err := programs.ParseHour(s)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestClientDo(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		resp    string
		in      any
		wantErr string
		wantCT  string
	}{
		{name: "ok without body", status: http.StatusOK},
		{name: "ok with body", status: http.StatusOK, in: ExecuteZone{Seconds: 5}, wantCT: "application/json"},
		{
			name:    "error with message",
			status:  http.StatusBadRequest,
			resp:    " zone not found\n",
			wantErr: "request GET /x: status code 400: zone not found",
		},
		{name: "error without message", status: http.StatusInternalServerError, wantErr: "request GET /x: status code 500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, cl := newFakeAPI(t, tt.status, tt.resp)
			err := cl.do(t.Context(), http.MethodGet, "/x", tt.in, nil)
			switch {
			case tt.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.wantErr != "" && (err == nil || err.Error() != tt.wantErr):
				t.Fatalf("expected error %q, got %v", tt.wantErr, err)
			}
			got := f.last()
			if got.auth != testToken {
				t.Errorf("expected Authorization %q, got %q", testToken, got.auth)
			}
			if got.contentType != tt.wantCT {
				t.Errorf("expected Content-Type %q, got %q", tt.wantCT, got.contentType)
			}
		})
	}
}

func TestClientDoDecodeError(t *testing.T) {
	_, cl := newFakeAPI(t, http.StatusOK, "not json")
	var out Status
	err := cl.do(t.Context(), http.MethodGet, "/status", nil, &out)
	if err == nil || !strings.HasPrefix(err.Error(), "decoding response GET /status") {
		t.Fatalf("expected decoding error, got %v", err)
	}
}

func TestClientDoUnreachable(t *testing.T) {
	cl := NewClient(testToken, "127.0.0.1:1", time.Second)
	if err := cl.do(t.Context(), http.MethodGet, "/status", nil, nil); err == nil {
		t.Fatal("expected error when the API is unreachable")
	}
}

func TestRequests(t *testing.T) {
	exe := &zones.Execution{ID: "front yard", Seconds: 10}
	zone := &zones.Zone{ID: "z1", Name: "Lawn", Relays: []int{1, 3}}
	weekly := &programs.Weekly{
		WeekDay:  programs.WeekDay(time.Monday),
		Programs: []programs.Program{{Hour: mustHour(t, "09:00"), Executions: []programs.Execution{{Seconds: 7, Zones: []string{"z2"}}}}},
	}
	daily := programs.DailyProgramType
	tests := []struct {
		name string
		call func(context.Context, *Client) error
		want recorded
	}{
		{
			name: "activate",
			call: func(ctx context.Context, cl *Client) error { return NewActivateRepository(cl).Activate(ctx, true) },
			want: recorded{method: http.MethodPatch, path: "/status/activate"},
		},
		{
			name: "deactivate",
			call: func(ctx context.Context, cl *Client) error { return NewActivateRepository(cl).Activate(ctx, false) },
			want: recorded{method: http.MethodPatch, path: "/status/deactivate"},
		},
		{
			name: "update status",
			call: func(ctx context.Context, cl *Client) error { return NewStatusRepository(cl).Update(ctx) },
			want: recorded{method: http.MethodGet, path: "/weather"},
		},
		{
			name: "execute zone",
			call: func(ctx context.Context, cl *Client) error {
				return NewExecuteZoneRepository(cl).SendExecution(ctx, exe)
			},
			want: recorded{method: http.MethodPost, path: "/zones/front%20yard/execute", body: `{"seconds":10}`},
		},
		{
			name: "fertilizer",
			call: func(ctx context.Context, cl *Client) error {
				return NewExecuteZoneRepository(cl).SendFertilizer(ctx, exe)
			},
			want: recorded{method: http.MethodPost, path: "/zones/fertilizer/front%20yard/execute", body: `{"seconds":10}`},
		},
		{
			name: "clean",
			call: func(ctx context.Context, cl *Client) error { return NewExecuteZoneRepository(cl).SendClean(ctx, "z1") },
			want: recorded{method: http.MethodPost, path: "/zones/clean/z1/execute"},
		},
		{
			name: "create zone",
			call: func(ctx context.Context, cl *Client) error { return NewZoneRepository(cl).Create(ctx, zone) },
			want: recorded{method: http.MethodPost, path: "/zones", body: `{"id":"z1","name":"Lawn","relays":[1,3]}`},
		},
		{
			name: "update zone",
			call: func(ctx context.Context, cl *Client) error { return NewZoneRepository(cl).Update(ctx, zone) },
			want: recorded{method: http.MethodPut, path: "/zones/z1", body: `{"name":"Lawn","relays":[1,3]}`},
		},
		{
			name: "delete zone",
			call: func(ctx context.Context, cl *Client) error { return NewZoneRepository(cl).Delete(ctx, "z1") },
			want: recorded{method: http.MethodDelete, path: "/zones/z1"},
		},
		{
			name: "save daily program",
			call: func(ctx context.Context, cl *Client) error {
				p := &programs.FertilizerProgram{
					Hour:       mustHour(t, "08:00"),
					Fertilizer: true,
					Executions: []programs.Execution{{Seconds: 5, Zones: []string{"z1"}}},
				}
				return NewAllProgramsRepository(cl).Save(ctx, p, &daily)
			},
			want: recorded{
				method: http.MethodPost,
				path:   "/programs/daily",
				body:   `{"executions":[{"seconds":5,"zones":["z1"]}],"hour":"08:00","fertilizer":true}`,
			},
		},
		{
			name: "remove daily program",
			call: func(ctx context.Context, cl *Client) error {
				h := mustHour(t, "08:00")
				return NewAllProgramsRepository(cl).Remove(ctx, &h, &daily)
			},
			want: recorded{method: http.MethodDelete, path: "/programs/daily/08:00"},
		},
		{
			name: "save weekly program",
			call: func(ctx context.Context, cl *Client) error { return NewWeeklyRepository(cl).Save(ctx, weekly) },
			want: recorded{
				method: http.MethodPost,
				path:   "/programs/weekly",
				body:   `{"week_day":"Monday","programs":[{"executions":[{"seconds":7,"zones":["z2"]}],"hour":"09:00"}]}`,
			},
		},
		{
			name: "remove weekly program",
			call: func(ctx context.Context, cl *Client) error {
				return NewWeeklyRepository(cl).Remove(ctx, &weekly.WeekDay)
			},
			want: recorded{method: http.MethodDelete, path: "/programs/weekly/monday"},
		},
		{
			name: "save temperature program",
			call: func(ctx context.Context, cl *Client) error {
				p := &programs.TemperatureProgram{Temperature: 30, Programs: weekly.Programs}
				return NewTemperatureRepository(cl).Save(ctx, p)
			},
			want: recorded{
				method: http.MethodPost,
				path:   "/programs/temperature",
				body:   `{"programs":[{"executions":[{"seconds":7,"zones":["z2"]}],"hour":"09:00"}],"temperature":30}`,
			},
		},
		{
			name: "remove temperature program",
			call: func(ctx context.Context, cl *Client) error { return NewTemperatureRepository(cl).Remove(ctx, 30) },
			want: recorded{method: http.MethodDelete, path: "/programs/temperature/30"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, cl := newFakeAPI(t, http.StatusOK, "")
			if err := tt.call(t.Context(), cl); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := f.last()
			if got.method != tt.want.method || got.path != tt.want.path || got.body != tt.want.body {
				t.Errorf("expected %s %s %s, got %s %s %s", tt.want.method, tt.want.path, tt.want.body, got.method, got.path, got.body)
			}
		})
	}
}

func TestFindStatus(t *testing.T) {
	_, cl := newFakeAPI(t, http.StatusOK, `{
		"system_started_at":"1700000000","updated_at":"1700000600",
		"Temperature":21.5,"Humidity":40,"is_raining":true,"is_day":true,"Active":true
	}`)
	st, err := NewStatusRepository(cl).Find(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.SystemStartedAt.Unix() != 1700000000 || st.UpdatedAt.Unix() != 1700000600 {
		t.Errorf("unexpected times: %v %v", st.SystemStartedAt, st.UpdatedAt)
	}
	if st.Temperature != 21.5 || st.Humidity != 40 || !st.IsRaining || !st.IsDay || !st.Active || st.Clean {
		t.Errorf("unexpected status: %+v", st)
	}
}

func TestFindStatusWithoutUpdate(t *testing.T) {
	_, cl := newFakeAPI(t, http.StatusOK, `{"system_started_at":"1700000000"}`)
	st, err := NewStatusRepository(cl).Find(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !st.UpdatedAt.Equal(st.SystemStartedAt) {
		t.Errorf("expected updated at to default to started at, got %v", st.UpdatedAt)
	}
}

func TestFindZones(t *testing.T) {
	_, cl := newFakeAPI(t, http.StatusOK, `[{"id":"z1","name":"Lawn","relays":[1,2]}]`)
	got, err := NewZoneRepository(cl).FindAll(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ID != "z1" || got[0].Name != "Lawn" || len(got[0].Relays) != 2 {
		t.Errorf("unexpected zones: %+v", got)
	}
}

func TestFindLogs(t *testing.T) {
	_, cl := newFakeAPI(t, http.StatusOK, `[{"executed_at":"1700000000","seconds":15,"zone_name":"Lawn"}]`)
	got, err := NewLogRepository(cl).Find(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ExecutedAt.Unix() != 1700000000 || got[0].Seconds != 15 || got[0].Zone != "Lawn" {
		t.Errorf("unexpected logs: %+v", got)
	}
}

func TestFindPrograms(t *testing.T) {
	_, cl := newFakeAPI(t, http.StatusOK, `{
		"daily":[{"hour":"08:00","fertilizer":true,"executions":[{"seconds":5,"zones":["z1"]}]}],
		"odd":[],"even":[{"hour":"10:00","executions":[]}],
		"weekly":[{"week_day":"Sunday","programs":[{"hour":"09:00","executions":[{"seconds":7,"zones":["z2"]}]}]}],
		"temperature":[{"temperature":30,"programs":[{"hour":"17:00","executions":[]}]}]
	}`)
	got, err := NewAllProgramsRepository(cl).FindAll(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Daily) != 1 || got.Daily[0].Hour.String() != "08:00" || !got.Daily[0].Fertilizer ||
		got.Daily[0].Executions[0].Seconds != 5 {
		t.Errorf("unexpected daily: %+v", got.Daily)
	}
	if len(got.Odd) != 0 || len(got.Even) != 1 || got.Even[0].Hour.String() != "10:00" {
		t.Errorf("unexpected odd/even: %+v %+v", got.Odd, got.Even)
	}
	if len(got.Weekly) != 1 || got.Weekly[0].WeekDay.String() != "Sunday" || got.Weekly[0].Programs[0].Hour.String() != "09:00" {
		t.Errorf("unexpected weekly: %+v", got.Weekly)
	}
	if len(got.Temperature) != 1 || got.Temperature[0].Temperature != 30 {
		t.Errorf("unexpected temperature: %+v", got.Temperature)
	}
}

func TestFindProgramsInvalidData(t *testing.T) {
	tests := map[string]string{
		"invalid hour":     `{"daily":[{"hour":"25:00"}]}`,
		"invalid week day": `{"weekly":[{"week_day":"Funday"}]}`,
	}
	for name, resp := range tests {
		t.Run(name, func(t *testing.T) {
			_, cl := newFakeAPI(t, http.StatusOK, resp)
			_, err := NewAllProgramsRepository(cl).FindAll(t.Context())
			if !errors.Is(err, programs.ErrInvalidExecutionHour) && !errors.Is(err, programs.ErrInvalidWeekDay) {
				t.Fatalf("expected a parse error, got %v", err)
			}
		})
	}
}
