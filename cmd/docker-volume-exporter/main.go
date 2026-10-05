// Command docker-volume-exporter exposes Docker volume sizes as Prometheus metrics.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/yansetiaji/docker-volume-exporter/internal/config"
	"github.com/yansetiaji/docker-volume-exporter/internal/docker"
	"github.com/yansetiaji/docker-volume-exporter/internal/exporter"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("docker-volume-exporter %s", version)
	exp := exporter.New(docker.NewClient(cfg.Socket), cfg.Interval, cfg.Timeout)
	go exp.Run(ctx)

	mux := http.NewServeMux()
	mux.Handle("/metrics", exp)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
