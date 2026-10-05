// Package config loads exporter settings from environment variables.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds the exporter settings.
type Config struct {
	Socket   string
	Port     string
	Interval time.Duration
	Timeout  time.Duration
}

// Load reads the configuration from the environment.
func Load() (Config, error) {
	return load(os.Getenv)
}

func load(get func(string) string) (Config, error) {
	env := func(key, def string) string {
		if v := get(key); v != "" {
			return v
		}
		return def
	}
	c := Config{
		Socket: env("DOCKER_SOCKET", "/var/run/docker.sock"),
		Port:   env("PORT", "9101"),
	}
	var err error
	if c.Interval, err = duration(env("INTERVAL", "60s")); err != nil {
		return Config{}, fmt.Errorf("invalid INTERVAL: %w", err)
	}
	if c.Timeout, err = duration(env("TIMEOUT", "5m")); err != nil {
		return Config{}, fmt.Errorf("invalid TIMEOUT: %w", err)
	}
	return c, nil
}

func duration(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	if d <= 0 {
		return 0, fmt.Errorf("%q must be positive", s)
	}
	return d, nil
}
