# Security Policy

## Supported versions

Only the latest release receives security fixes.

## Reporting a vulnerability

Please do not open a public issue. Use GitHub's
[private vulnerability reporting](https://github.com/yansetiaji/docker-volume-exporter/security/advisories/new)
to send a report. You can expect an initial response within a few days.

## Note on the Docker socket

The exporter talks to the Docker Engine API through the Docker socket, and
access to that socket is equivalent to root on the host. Mount it read-only
(`/var/run/docker.sock:/var/run/docker.sock:ro`), do not expose the exporter
port to untrusted networks, and consider a socket proxy that only allows
`GET /system/df`.
