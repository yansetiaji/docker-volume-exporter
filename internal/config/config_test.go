package config

import (
	"testing"
	"time"
)

func fake(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadDefaults(t *testing.T) {
	c, err := load(fake(nil))
	if err != nil {
		t.Fatal(err)
	}
	want := Config{Socket: "/var/run/docker.sock", Port: "9101", Interval: time.Minute, Timeout: 5 * time.Minute}
	if c != want {
		t.Fatalf("got %+v, want %+v", c, want)
	}
}

func TestLoadOverrides(t *testing.T) {
	c, err := load(fake(map[string]string{"PORT": "8080", "INTERVAL": "10s", "TIMEOUT": "1m", "DOCKER_SOCKET": "/x.sock"}))
	if err != nil {
		t.Fatal(err)
	}
	want := Config{Socket: "/x.sock", Port: "8080", Interval: 10 * time.Second, Timeout: time.Minute}
	if c != want {
		t.Fatalf("got %+v, want %+v", c, want)
	}
}

func TestLoadInvalid(t *testing.T) {
	for _, env := range []map[string]string{
		{"INTERVAL": "abc"},
		{"INTERVAL": "0s"},
		{"TIMEOUT": "-1m"},
		{"TIMEOUT": "nope"},
	} {
		if _, err := load(fake(env)); err == nil {
			t.Errorf("expected error for %v", env)
		}
	}
}
