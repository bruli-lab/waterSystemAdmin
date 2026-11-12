package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bruli/waterSystemAdmin/internal/domain/password"
	"github.com/bruli/waterSystemAdmin/internal/infra/disk"

	"github.com/gorilla/sessions"

	"github.com/bruli/waterSystemAdmin/internal/domain/execution"
	"github.com/bruli/waterSystemAdmin/internal/domain/logs"
	"github.com/bruli/waterSystemAdmin/internal/domain/programs"
	"github.com/bruli/waterSystemAdmin/internal/domain/zones"
	"github.com/bruli/waterSystemAdmin/internal/infra/http/templates"

	"github.com/bruli/waterSystemAdmin/internal/config"
	"github.com/bruli/waterSystemAdmin/internal/domain/status"
	"github.com/bruli/waterSystemAdmin/internal/infra/api"
	"github.com/bruli/waterSystemAdmin/internal/infra/http/controller"
	pongo2 "github.com/flosch/pongo2/v6"
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

	cl := api.NewClient(conf.ApiAuthKey, conf.ApiUrl, 5*time.Second)
	statusRepo := api.NewStatusRepository(cl)
	logsRepo := api.NewLogRepository(cl)
	programsRepo := api.NewAllProgramsRepository(cl)
	executionRepo := api.NewExecuteZoneRepository(cl)
	activateRepo := api.NewActivateRepository(cl)
	zoneRepo := api.NewZoneRepository(cl)
	weeklyRepo := api.NewWeeklyRepository(cl)
	tempRepo := api.NewTemperatureRepository(cl)
	passwordRepo := disk.NewPasswordRepository(conf.PasswordFile)

	findZones := zones.NewFindZones(zoneRepo)
	passwordExists := password.NewExists(passwordRepo)

	tplSet := pongo2.NewSet("embed", templates.NewEmbedLoader(templates.FS, "."))

	sessionStore, err := buildSessionStore(conf.SessionKey, log)
	if err != nil {
		log.Error("failed to build session store", slog.String("error", err.Error()))
		return err
	}

	authMdw := controller.AuthMiddleware(sessionStore, passwordExists, log)

	staticFiles, _ := fs.Sub(templates.FS, "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))

	findStatus := status.NewFindStatus(statusRepo)

	http.HandleFunc("/{$}", authMdw(controller.FindStatus(tplSet, findStatus, log)))
	http.HandleFunc("/logs", authMdw(controller.FindLogs(tplSet, logs.NewFindLogs(logsRepo), findStatus, log)))
	http.HandleFunc("/update-status", authMdw(controller.UpdateStatus(status.NewUpdateStatus(statusRepo), log)))
	http.HandleFunc("/programs", authMdw(controller.Programs(tplSet, programs.NewFindAllPrograms(programsRepo), findStatus, log)))
	http.HandleFunc("/programs/new", authMdw(controller.CreateProgram(tplSet, findZones, programs.NewCreate(programsRepo), findStatus, log)))
	http.HandleFunc("/programs/delete/{hour}/{type}", authMdw(controller.RemoveProgram(tplSet, findStatus, programs.NewRemove(programsRepo), log)))
	http.HandleFunc("/programs/weekly/new", authMdw(controller.CreateWeeklyProgram(tplSet, findZones, findStatus, programs.NewCreateWeekly(weeklyRepo), log)))
	http.HandleFunc("/programs/weekly/delete/{weekday}", authMdw(controller.RemoveWeeklyProgram(tplSet, findStatus, programs.NewRemoveWeekly(weeklyRepo), log)))
	http.HandleFunc("/programs/temperature/new", authMdw(controller.CreateTemperatureProgram(tplSet, findZones, findStatus, programs.NewCreateTemperature(tempRepo), log)))
	http.HandleFunc("/programs/temperature/delete/{temperature}", authMdw(controller.RemoveTemperatureProgram(tplSet, findStatus, programs.NewRemoveTemperature(tempRepo), log)))
	http.HandleFunc("/zones", authMdw(controller.Zones(tplSet, findZones, findStatus, log)))
	http.HandleFunc("/execution/{id}", authMdw(controller.Execution(tplSet, execution.NewExecuteZone(executionRepo), findStatus, log)))
	http.HandleFunc("/fertilizer/{id}", authMdw(controller.Fertilizer(tplSet, execution.NewExecuteZone(executionRepo), findStatus, log)))
	http.HandleFunc("/clean/{id}", authMdw(controller.Clean(execution.NewClean(executionRepo), log)))
	http.HandleFunc("/deactivate", authMdw(controller.Deactivate(status.NewActivateDeactivate(activateRepo), log)))
	http.HandleFunc("/activate", authMdw(controller.Activate(status.NewActivateDeactivate(activateRepo), log)))
	http.HandleFunc("/zones/new", authMdw(controller.CreateZone(tplSet, zones.NewCreate(zoneRepo), findStatus, log)))
	http.HandleFunc("/zones/{id}/delete", authMdw(controller.DeleteZone(zones.NewDelete(zoneRepo), log)))
	http.HandleFunc("/zones/{id}/edit", authMdw(controller.UpdateZone(tplSet, findZones, zones.NewUpdate(zoneRepo), findStatus, log)))
	http.HandleFunc("/login", controller.Login(tplSet, sessionStore, password.NewCheck(passwordRepo), findStatus, log))
	http.HandleFunc("/logout", controller.Logout(sessionStore, log))
	http.HandleFunc("/password", controller.Password(tplSet, password.NewCreate(passwordRepo), passwordExists, log))

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
