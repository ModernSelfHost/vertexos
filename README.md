# Vertex OS

A simplified home NAS operating system: Debian base + OpenZFS, managed by a Go
daemon (`vertexd`) with a React/TypeScript web dashboard.

See [design.md](./design.md) for the full design doc and [Todo.md](./Todo.md) for
the task tracker.

## Development

Prerequisites: Go 1.27+, Node 22+, (optional) golangci-lint.

```sh
make build        # build vertexd + vertexctl
make test         # run Go tests
make lint         # golangci-lint
cd web && npm install && npm run dev   # web dev server
```

## Layout

- `cmd/` — entrypoints (`vertexd`, `vertexctl`)
- `internal/` — daemon packages
- `web/` — React/TypeScript SPA
- `installer/` — image build pipeline (Milestone 3)
- `docs/` — documentation
