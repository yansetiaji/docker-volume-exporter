// Package exporter caches Docker volume usage and serves it as Prometheus metrics.
package exporter

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/yansetiaji/docker-volume-exporter/internal/docker"
)

// VolumeLister fetches the current Docker volumes.
type VolumeLister interface {
	Volumes(ctx context.Context) ([]docker.Volume, error)
}

// Exporter refreshes volume data in the background and serves it from cache.
type Exporter struct {
	lister   VolumeLister
	interval time.Duration
	timeout  time.Duration

	mu          sync.RWMutex
	volumes     []docker.Volume
	lastRefresh time.Time
}

// New returns an Exporter that refreshes every interval, allowing each call up to timeout.
func New(l VolumeLister, interval, timeout time.Duration) *Exporter {
	return &Exporter{lister: l, interval: interval, timeout: timeout}
}

// Run refreshes immediately and then every interval until ctx is cancelled.
// Refreshes never overlap; a tick that fires during a slow call is skipped.
func (e *Exporter) Run(ctx context.Context) {
	e.Refresh(ctx)
	t := time.NewTicker(e.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			e.Refresh(ctx)
		}
	}
}

// Refresh fetches volumes once. On failure the error is logged and the cache is kept.
func (e *Exporter) Refresh(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()
	v, err := e.lister.Volumes(ctx)
	if err != nil {
		log.Printf("refresh failed: %v", err)
		return
	}
	e.mu.Lock()
	e.volumes, e.lastRefresh = v, time.Now()
	e.mu.Unlock()
}

var esc = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)

// ServeHTTP writes the cached data in Prometheus text format.
func (e *Exporter) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var b strings.Builder
	b.WriteString("# TYPE docker_volume_size_bytes gauge\n")
	for _, v := range e.volumes {
		if v.UsageData != nil && v.UsageData.Size >= 0 {
			fmt.Fprintf(&b, "docker_volume_size_bytes{volume=\"%s\",driver=\"%s\"} %d\n", esc.Replace(v.Name), esc.Replace(v.Driver), v.UsageData.Size)
		}
	}
	b.WriteString("# TYPE docker_volume_ref_count gauge\n")
	for _, v := range e.volumes {
		if v.UsageData != nil && v.UsageData.RefCount >= 0 {
			fmt.Fprintf(&b, "docker_volume_ref_count{volume=\"%s\"} %d\n", esc.Replace(v.Name), v.UsageData.RefCount)
		}
	}
	b.WriteString("# TYPE docker_volume_exporter_last_refresh_timestamp_seconds gauge\n")
	fmt.Fprintf(&b, "docker_volume_exporter_last_refresh_timestamp_seconds %d\n", e.lastRefresh.Unix())
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = io.WriteString(w, b.String())
}
