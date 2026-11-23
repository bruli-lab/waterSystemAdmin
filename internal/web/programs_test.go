package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/bruli-lab/waterSystemAdmin/internal/api"
)

func TestParseSeconds(t *testing.T) {
	tests := []struct {
		value   string
		want    int
		wantErr bool
	}{
		{value: "1", want: 1},
		{value: "120", want: 120},
		{value: "0", wantErr: true},
		{value: "-5", wantErr: true},
		{value: "121", wantErr: true},
		{value: "abc", wantErr: true},
		{value: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := parseSeconds(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseSeconds(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseSeconds(%q) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestErrMessage(t *testing.T) {
	down := fmt.Errorf("%w: request GET /zones: connection refused", api.ErrUnavailable)
	if got := errMessage("Could not read the zones", down); got != "Could not read the zones: the irrigation API is not reachable." {
		t.Errorf("unexpected message for unavailable API: %q", got)
	}
	if got := errMessage("Could not read the zones", errors.New("status code 500")); got != "Could not read the zones: status code 500" {
		t.Errorf("unexpected message: %q", got)
	}
}

func formRequest(t *testing.T, values url.Values) *http.Request {
	t.Helper()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

func TestParseProgramStatus(t *testing.T) {
	fert := url.Values{"seconds[]": {"10"}, "zones_0[]": {"z1"}}
	for _, tt := range []struct {
		name    string
		values  url.Values
		editing bool
		want    bool
	}{
		{name: "new program is enabled", values: fert, want: true},
		{name: "edit without switch disables", values: fert, editing: true, want: false},
		{
			name:    "edit with switch enables",
			values:  url.Values{"seconds[]": {"10"}, "zones_0[]": {"z1"}, "enabled": {"true"}},
			editing: true,
			want:    true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parseFertilizerProgram(formRequest(t, tt.values), "08:00", tt.editing)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Enabled != tt.want {
				t.Errorf("enabled = %v, want %v", p.Enabled, tt.want)
			}
		})
	}
}

func TestParseHourProgramsStatus(t *testing.T) {
	values := url.Values{
		"hours[]":                {"08:00", "09:00"},
		"executions_0_seconds[]": {"10"},
		"executions_0_zones_0[]": {"z1"},
		"executions_1_seconds[]": {"10"},
		"executions_1_zones_0[]": {"z1"},
		"hours_1_enabled":        {"true"},
	}
	edited, err := parseHourPrograms(formRequest(t, values), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if edited[0].Enabled || !edited[1].Enabled {
		t.Errorf("unexpected status when editing: %v %v", edited[0].Enabled, edited[1].Enabled)
	}
	created, err := parseHourPrograms(formRequest(t, values), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created[0].Enabled || !created[1].Enabled {
		t.Errorf("new programs must be enabled: %v %v", created[0].Enabled, created[1].Enabled)
	}
}
