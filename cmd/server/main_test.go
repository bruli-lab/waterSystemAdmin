package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/bruli/waterSystemAdmin/internal/disk"
	"github.com/bruli/waterSystemAdmin/internal/web"
	"github.com/gorilla/sessions"
)

func TestRoutesRejectCrossOriginPosts(t *testing.T) {
	passwordRepo := disk.NewPasswordRepository(filepath.Join(t.TempDir(), "password"))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app, err := web.NewApp(web.API{}, passwordRepo, sessions.NewCookieStore([]byte("test-key")), logger)
	if err != nil {
		t.Fatal(err)
	}
	handler := routes(app)

	tests := []struct {
		name         string
		secFetchSite string
		wantStatus   int
	}{
		{name: "cross-site", secFetchSite: "cross-site", wantStatus: http.StatusForbidden},
		{name: "same-origin", secFetchSite: "same-origin", wantStatus: http.StatusSeeOther},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/activate", http.NoBody)
			req.Header.Set("Sec-Fetch-Site", tt.secFetchSite)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestHealthz(t *testing.T) {
	passwordRepo := disk.NewPasswordRepository(filepath.Join(t.TempDir(), "password"))
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app, err := web.NewApp(web.API{}, passwordRepo, sessions.NewCookieStore([]byte("test-key")), logger)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	routes(app).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", http.NoBody))
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
