package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type volume struct {
	Name      string
	Driver    string
	UsageData *struct {
		Size     int64
		RefCount int64
	}
}

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

var (
	mu          sync.RWMutex
	volumes     []volume
	lastRefresh time.Time
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func fetch(ctx context.Context, c *http.Client) ([]volume, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/v1.43/system/df?type=volume", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker api: %s", resp.Status)
	}
	var out struct{ Volumes []volume }
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Volumes, nil
}

func refresh(c *http.Client, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	v, err := fetch(ctx, c)
	if err != nil {
		log.Printf("refresh failed: %v", err)
		return
	}
	mu.Lock()
	volumes, lastRefresh = v, time.Now()
	mu.Unlock()
}

var esc = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`)

func metrics(w http.ResponseWriter, _ *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	var b strings.Builder
	b.WriteString("# TYPE docker_volume_size_bytes gauge\n")
	for _, v := range volumes {
		if v.UsageData != nil && v.UsageData.Size >= 0 {
			fmt.Fprintf(&b, "docker_volume_size_bytes{volume=\"%s\",driver=\"%s\"} %d\n", esc.Replace(v.Name), esc.Replace(v.Driver), v.UsageData.Size)
		}
	}
	b.WriteString("# TYPE docker_volume_ref_count gauge\n")
	for _, v := range volumes {
		if v.UsageData != nil && v.UsageData.RefCount >= 0 {
			fmt.Fprintf(&b, "docker_volume_ref_count{volume=\"%s\"} %d\n", esc.Replace(v.Name), v.UsageData.RefCount)
		}
	}
	b.WriteString("# TYPE docker_volume_exporter_last_refresh_timestamp_seconds gauge\n")
	fmt.Fprintf(&b, "docker_volume_exporter_last_refresh_timestamp_seconds %d\n", lastRefresh.Unix())
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprint(w, b.String())
}

func newClient(socket string) *http.Client {
	return &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socket)
		},
	}}
}

func main() {
	socket := env("DOCKER_SOCKET", "/var/run/docker.sock")
	interval, err := time.ParseDuration(env("INTERVAL", "60s"))
	if err != nil || interval <= 0 {
		log.Fatalf("invalid INTERVAL: %v", err)
	}
	timeout, err := time.ParseDuration(env("TIMEOUT", "5m"))
	if err != nil || timeout <= 0 {
		log.Fatalf("invalid TIMEOUT: %v", err)
	}
	client := newClient(socket)

	log.Printf("docker-volume-exporter %s", version)
	refresh(client, timeout)
	go func() {
		for range time.Tick(interval) {
			refresh(client, timeout)
		}
	}()

	http.HandleFunc("/metrics", metrics)
	addr := ":" + env("PORT", "9101")
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
