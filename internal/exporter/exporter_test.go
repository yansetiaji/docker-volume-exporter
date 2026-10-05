package exporter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yansetiaji/docker-volume-exporter/internal/docker"
)

type fakeLister struct {
	vols []docker.Volume
	err  error
}

func (f fakeLister) Volumes(context.Context) ([]docker.Volume, error) { return f.vols, f.err }

func usage(size, ref int64) *docker.UsageData { return &docker.UsageData{Size: size, RefCount: ref} }

func scrape(e *Exporter) string {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	return rec.Body.String()
}

func TestMetrics(t *testing.T) {
	e := New(fakeLister{vols: []docker.Volume{
		{Name: "data", Driver: "local", UsageData: usage(1234, 2)},
		{Name: "uncomputed", Driver: "local", UsageData: usage(-1, -1)},
		{Name: "no-usage", Driver: "local"},
		{Name: "we\"ird\\na\nme", Driver: "local", UsageData: usage(1, 0)},
	}}, time.Minute, time.Second)
	e.Refresh(context.Background())
	body := scrape(e)

	for _, want := range []string{
		`docker_volume_size_bytes{volume="data",driver="local"} 1234`,
		`docker_volume_ref_count{volume="data"} 2`,
		`docker_volume_size_bytes{volume="we\"ird\\na\nme",driver="local"} 1`,
		`docker_volume_exporter_last_refresh_timestamp_seconds`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
	for _, bad := range []string{"uncomputed", "no-usage"} {
		if strings.Contains(body, bad) {
			t.Errorf("unexpected %q in:\n%s", bad, body)
		}
	}
}

func TestRefreshFailureKeepsCache(t *testing.T) {
	l := &fakeLister{vols: []docker.Volume{{Name: "keep", Driver: "local", UsageData: usage(1, 1)}}}
	e := New(l, time.Minute, time.Second)
	e.Refresh(context.Background())
	l.vols, l.err = nil, errors.New("down")
	e.Refresh(context.Background())
	if !strings.Contains(scrape(e), `volume="keep"`) {
		t.Fatal("cache dropped after failed refresh")
	}
}

func TestRunStopsOnCancel(t *testing.T) {
	e := New(fakeLister{}, time.Millisecond, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { e.Run(ctx); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not stop after cancel")
	}
}
