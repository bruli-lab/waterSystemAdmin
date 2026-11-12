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

	"github.com/bruli/waterSystemAdmin/internal/config"
	"github.com/bruli/waterSystemAdmin/internal/infra/api"
	"github.com/bruli/waterSystemAdmin/internal/infra/disk"
	"github.com/bruli/waterSystemAdmin/internal/infra/http/web"
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

	auth := app.RequireAuth
	http.Handle("GET /static/", web.Static())

	http.HandleFunc("/login", app.Login)
	http.HandleFunc("GET /logout", app.Logout)
	http.HandleFunc("/password", app.Password)

	http.HandleFunc("GET /{$}", auth(app.Status))
	http.HandleFunc("POST /activate", auth(app.Activate))
	http.HandleFunc("POST /deactivate", auth(app.Deactivate))
	http.HandleFunc("POST /update-status", auth(app.UpdateStatus))
	http.HandleFunc("GET /logs", auth(app.Logs))

	http.HandleFunc("GET /zones", auth(app.Zones))
	http.HandleFunc("/zones/new", auth(app.CreateZone))
	http.HandleFunc("/zones/{id}/edit", auth(app.UpdateZone))
	http.HandleFunc("POST /zones/{id}/delete", auth(app.DeleteZone))
	http.HandleFunc("/execution/{id}", auth(app.Execute))
	http.HandleFunc("/fertilizer/{id}", auth(app.Fertilizer))
	http.HandleFunc("POST /clean/{id}", auth(app.Clean))

	http.HandleFunc("GET /programs", auth(app.Programs))
	http.HandleFunc("/programs/new", auth(app.CreateProgram))
	http.HandleFunc("POST /programs/delete/{hour}/{type}", auth(app.RemoveProgram))
	http.HandleFunc("/programs/weekly/new", auth(app.CreateWeeklyProgram))
	http.HandleFunc("POST /programs/weekly/delete/{weekday}", auth(app.RemoveWeeklyProgram))
	http.HandleFunc("/programs/temperature/new", auth(app.CreateTemperatureProgram))
	http.HandleFunc("POST /programs/temperature/delete/{temperature}", auth(app.RemoveTemperatureProgram))

	srv := &http.Server{
		Addr:    conf.ServerURL,
		Handler: nil,
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
