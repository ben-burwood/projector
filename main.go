package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "time/tzdata" // embed the timezone DB; the scratch image has none

	"projector/internal/api"
	"projector/internal/config"
	"projector/internal/provider"
	"projector/internal/store"
)

// version is overridable at build time via -ldflags "-X main.version=...".
var version = "dev"

// Fixed runtime settings. The app always runs in its container, so these are
// constants rather than configuration.
const (
	listenAddr = "[::]:8080"          // HTTP listen address
	dbPath     = "/data/projector.db" // SQLite path (persisted on the /data volume)
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	st, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()

	spa, err := spaHandler()
	if err != nil {
		return err
	}

	registry, releases := provider.Build(provider.Options{
		GitHubToken: cfg.GitHubToken,
		GitLabToken: cfg.GitLabToken,
		GiteaTokens: cfg.GiteaTokens,
	})

	apiHandler := api.New(st, registry, releases, version).Handler()

	root := http.NewServeMux()
	root.Handle("/api/", apiHandler)
	root.Handle("/", spa)

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           root,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", listenAddr, "version", version, "db", dbPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
