# docker-volume-exporter

[![CI](https://github.com/yansetiaji/docker-volume-exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/yansetiaji/docker-volume-exporter/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/yansetiaji/docker-volume-exporter)](https://github.com/yansetiaji/docker-volume-exporter/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/yansetiaji/docker-volume-exporter)](https://goreportcard.com/report/github.com/yansetiaji/docker-volume-exporter)

Minimal Prometheus exporter for Docker volume sizes. Sizes come from the Docker Engine API (`/system/df?type=volume`).

## Install

Docker image (linux/amd64, linux/arm64):

```sh
docker pull ghcr.io/yansetiaji/docker-volume-exporter:latest
```

Or download a binary (linux/darwin, amd64/arm64) from the [Releases](https://github.com/yansetiaji/docker-volume-exporter/releases) page.

A background goroutine refreshes the data every `INTERVAL`, so `/metrics` always responds instantly from cache. If a refresh fails, the error is logged and the last good data keeps being served.

## Run

```sh
docker compose up -d
curl localhost:9101/metrics
```

Or from source (Go 1.27.1):

```sh
go run .
```

## Configuration

| Env | Default | Description |
|---|---|---|
| `PORT` | `9101` | HTTP listen port |
| `INTERVAL` | `60s` | Refresh interval (Go duration) |
| `TIMEOUT` | `5m` | Max time for one Docker API call. Refreshes never overlap; a tick that fires during a slow call is skipped |
| `DOCKER_SOCKET` | `/var/run/docker.sock` | Docker unix socket path |

## Metrics

```
docker_volume_size_bytes{volume="name",driver="local"} 12345
docker_volume_ref_count{volume="name"} 1
docker_volume_exporter_last_refresh_timestamp_seconds 1791197093
```

Volumes whose size Docker has not computed (-1) are skipped.

## Build

```sh
docker build -t ghcr.io/yansetiaji/docker-volume-exporter:0.1 .
```

For a specific platform:

```sh
docker buildx build --platform linux/amd64 -t ghcr.io/yansetiaji/docker-volume-exporter:0.1 --load .
```

## Security

Access to the Docker socket is equivalent to root on the host. Mount it read-only, as in the compose example, and do not expose the exporter port to untrusted networks. See [SECURITY.md](SECURITY.md).

## Releasing

Tag and push; the release workflow publishes binaries, checksums and the multi-arch image to ghcr.io:

```sh
git tag v0.2.0 && git push origin v0.2.0
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Licensed under the [MIT License](LICENSE).

## Prometheus

```yaml
scrape_configs:
  - job_name: docker-volumes
    static_configs:
      - targets: ["docker-volume-exporter:9101"]
```
