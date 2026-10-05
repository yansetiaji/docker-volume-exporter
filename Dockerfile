FROM --platform=$BUILDPLATFORM golang:1.27 AS build
ARG TARGETOS TARGETARCH VERSION=dev
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w -X main.version=$VERSION" -o /exporter ./cmd/docker-volume-exporter

FROM scratch
LABEL org.opencontainers.image.source="https://github.com/yansetiaji/docker-volume-exporter" \
      org.opencontainers.image.description="Prometheus exporter for Docker volume sizes" \
      org.opencontainers.image.licenses="MIT"
COPY --from=build /exporter /exporter
ENTRYPOINT ["/exporter"]
