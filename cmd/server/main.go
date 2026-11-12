package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bruli-lab/waterSystemAdmin/internal/api"
	"github.com/bruli-lab/waterSystemAdmin/internal/config"
	"github.com/bruli-lab/waterSystemAdmin/internal/disk"
	"github.com/bruli-lab/waterSystemAdmin/internal/web"
	"github.com/gorilla/sessions"
)

const serviceName = "waterSystemAdmin"

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conf, err := config.New()
	if err != nil {
		fmt.Printf("failed to load config: %s\n", err.Error())
		return err
	}
	log := buildLog(conf.LogLevel)

	sessionStore, err := buildSessionStore(conf.SessionKey, log)
	if err != nil {
		log.Error("failed to build session store", slog.String("error", err.Error()))
		return err
	}

	cl := api.NewClient(conf.ApiAuthKey, conf.ApiUrl, 5*time.Second)
	app, err := web.NewApp(web.API{
		Status:      api.NewStatusRepository(cl),
		Activate:    api.NewActivateRepository(cl),
		Zones:       api.NewZoneRepository(cl),
		Execute:     api.NewExecuteZoneRepository(cl),
		Logs:        api.NewLogRepository(cl),
		Programs:    api.NewAllProgramsRepository(cl),
		Weekly:      api.NewWeeklyRepository(cl),
		Temperature: api.NewTemperatureRepository(cl),
	}, disk.NewPasswordRepository(conf.PasswordFile), sessionStore, log)
	if err != nil {
		log.Error("failed to build web app", slog.String("error", err.Error()))
		return err
	}

	srv := &http.Server{
		Addr:              conf.ServerURL,
		Handler:           routes(app),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error)
	defer close(errCh)

	go func() {
		if err := runServer(ctx, log, conf, srv); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.InfoContext(ctx, "received interrupt signal")

		if err = srv.Shutdown(context.Background()); err != nil {
			log.ErrorContext(ctx, "failed to shutdown server", slog.String("error", err.Error()))
		}
	}

	return nil
}

func routes(app *web.App) http.Handler {
	mux := http.NewServeMux()
	auth := app.RequireAuth
	mux.Handle("GET /static/", web.Static())
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	mux.HandleFunc("/login", app.Login)
	mux.HandleFunc("GET /logout", app.Logout)
	mux.HandleFunc("/password", app.Password)

	mux.HandleFunc("GET /{$}", auth(app.Status))
	mux.HandleFunc("POST /activate", auth(app.Activate))
	mux.HandleFunc("POST /deactivate", auth(app.Deactivate))
	mux.HandleFunc("POST /update-status", auth(app.UpdateStatus))
	mux.HandleFunc("GET /logs", auth(app.Logs))

	mux.HandleFunc("GET /zones", auth(app.Zones))
	mux.HandleFunc("/zones/new", auth(app.CreateZone))
	mux.HandleFunc("/zones/{id}/edit", auth(app.UpdateZone))
	mux.HandleFunc("POST /zones/{id}/delete", auth(app.DeleteZone))
	mux.HandleFunc("/execution/{id}", auth(app.Execute))
	mux.HandleFunc("/fertilizer/{id}", auth(app.Fertilizer))
	mux.HandleFunc("POST /clean/{id}", auth(app.Clean))

	mux.HandleFunc("GET /programs", auth(app.Programs))
	mux.HandleFunc("/programs/new", auth(app.CreateProgram))
	mux.HandleFunc("POST /programs/delete/{hour}/{type}", auth(app.RemoveProgram))
	mux.HandleFunc("/programs/weekly/new", auth(app.CreateWeeklyProgram))
	mux.HandleFunc("POST /programs/weekly/delete/{weekday}", auth(app.RemoveWeeklyProgram))
	mux.HandleFunc("/programs/temperature/new", auth(app.CreateTemperatureProgram))
	mux.HandleFunc("POST /programs/temperature/delete/{temperature}", auth(app.RemoveTemperatureProgram))
	return http.NewCrossOriginProtection().Handler(mux)
}

func runServer(ctx context.Context, log *slog.Logger, conf *config.Config, srv *http.Server) error {
	log.InfoContext(ctx, "starting server", slog.String("url", conf.ServerURL))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.ErrorContext(ctx, "failed to start server", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func buildSessionStore(key string, log *slog.Logger) (*sessions.CookieStore, error) {
	k := []byte(key)
	if len(k) == 0 {
		log.Warn("SESSION_KEY not set, using a random key: sessions will not survive restarts")
		k = make([]byte, 32)
		if _, err := rand.Read(k); err != nil {
			return nil, err
		}
	}
	store := sessions.NewCookieStore(k)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	return store, nil
}

func buildLog(level string) *slog.Logger {
	levels := map[string]slog.Level{
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
		"error": slog.LevelError,
		"warn":  slog.LevelWarn,
	}
	l := levels[level]

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: l,
	})

	log := slog.New(handler)
	return log.With("service", serviceName)
}
