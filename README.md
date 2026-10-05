# docker-volume-exporter

Minimal Prometheus exporter for Docker volume sizes. Sizes come from the Docker Engine API (`/system/df?type=volume`), not `du`.

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

On macOS Docker Desktop / Colima, set `DOCKER_SOCKET` to your socket path.

## Configuration

| Env | Default | Description |
|---|---|---|
| `PORT` | `9101` | HTTP listen port |
| `INTERVAL` | `60s` | Refresh interval (Go duration) |
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
docker buildx build --platform linux/amd64 -t ghcr.io/yansetiaji/docker-volume-exporter:0.1 --load .
```

The Dockerfile cross-compiles on the build host, so it works on Apple Silicon without emulating the Go compiler.

## Prometheus

```yaml
scrape_configs:
  - job_name: docker-volumes
    static_configs:
      - targets: ["docker-volume-exporter:9101"]
```
