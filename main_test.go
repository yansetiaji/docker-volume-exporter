package main

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func usage(size, ref int64) *struct{ Size, RefCount int64 } {
	return &struct{ Size, RefCount int64 }{size, ref}
}

func TestMetrics(t *testing.T) {
	mu.Lock()
	volumes = []volume{
		{Name: "data", Driver: "local", UsageData: usage(1234, 2)},
		{Name: "uncomputed", Driver: "local", UsageData: usage(-1, -1)},
		{Name: "no-usage", Driver: "local"},
		{Name: "we\"ird\\na\nme", Driver: "local", UsageData: usage(1, 0)},
	}
	mu.Unlock()

	rec := httptest.NewRecorder()
	metrics(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := rec.Body.String()

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

// serve starts handler h on a temporary unix socket and returns a client for it.
func serve(t *testing.T, h http.HandlerFunc) *http.Client {
	t.Helper()
	sock := filepath.Join(t.TempDir(), "docker.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: h}
	go srv.Serve(l)
	t.Cleanup(func() { srv.Close() })
	return newClient(sock)
}

func TestFetch(t *testing.T) {
	c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1.43/system/df" || r.URL.Query().Get("type") != "volume" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Write([]byte(`{"Volumes":[{"Name":"a","Driver":"local","UsageData":{"Size":42,"RefCount":1}}]}`))
	})

	got, err := fetch(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "a" || got[0].UsageData.Size != 42 || got[0].UsageData.RefCount != 1 {
		t.Fatalf("unexpected volumes: %+v", got)
	}
}

func TestFetchNon200(t *testing.T) {
	c := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	if _, err := fetch(context.Background(), c); err == nil {
		t.Fatal("expected error on non-200")
	}
}
