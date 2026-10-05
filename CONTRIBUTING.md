# Contributing

Thanks for helping out! This project aims to stay small, so please keep changes focused.

## Development

Requires Go 1.27.1 and, for the image, Docker.

```sh
go build ./...
go vet ./...
go test -race ./...
gofmt -l .        # should print nothing
```

## Pull requests

1. Fork and branch from `main`.
2. Add or update tests for behavior changes.
3. Make sure the commands above pass; CI runs the same checks.
4. Open a PR against `main` with a short description of what and why.

Use clear, imperative commit messages (e.g. `Add TIMEOUT setting`). Release notes are generated from them.

## Reporting bugs and requesting features

Open an issue using the templates. For security problems, see [SECURITY.md](SECURITY.md).

By contributing you agree that your contributions are licensed under the [MIT License](LICENSE).
