package docker

import (
	"context"
	"net"
	"net/http"
	"path/filepath"
	"testing"
)

// serve starts handler h on a temporary unix socket and returns a client for it.
func serve(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	sock := filepath.Join(t.TempDir(), "docker.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: h}
	go func() { _ = srv.Serve(l) }()
	t.Cleanup(func() { _ = srv.Close() })
	return NewClient(sock)
}

func TestVolumes(t *testing.T) {
	c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1.43/system/df" || r.URL.Query().Get("type") != "volume" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"Volumes":[{"Name":"a","Driver":"local","UsageData":{"Size":42,"RefCount":1}}]}`))
	})

	got, err := c.Volumes(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "a" || got[0].UsageData.Size != 42 || got[0].UsageData.RefCount != 1 {
		t.Fatalf("unexpected volumes: %+v", got)
	}
}

func TestVolumesNon200(t *testing.T) {
	c := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	if _, err := c.Volumes(context.Background()); err == nil {
		t.Fatal("expected error on non-200")
	}
}
