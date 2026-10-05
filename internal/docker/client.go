// Package docker is a minimal Docker Engine API client over a unix socket.
package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

const apiVersion = "v1.43"

// Volume is a Docker volume with its usage data, if Docker computed it.
type Volume struct {
	Name      string
	Driver    string
	UsageData *UsageData
}

// UsageData is the per-volume usage reported by /system/df.
type UsageData struct {
	Size     int64
	RefCount int64
}

// Client talks to the Docker Engine API.
type Client struct {
	http *http.Client
}

// NewClient returns a Client that dials the given unix socket.
func NewClient(socket string) *Client {
	return &Client{http: &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socket)
		},
	}}}
}

// Volumes returns all volumes with usage data (/system/df?type=volume).
func (c *Client) Volumes(ctx context.Context) ([]Volume, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/"+apiVersion+"/system/df?type=volume", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("docker api: %s", resp.Status)
	}
	var out struct{ Volumes []Volume }
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Volumes, nil
}
