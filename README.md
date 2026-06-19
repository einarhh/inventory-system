# Inventory

A nested-location inventory system. Go backend + PostgreSQL, web UI to follow.

See `AGENTS.md` for architecture and scope, `docs/SPEC.md` for MVP requirements.

## Prerequisites
- Go 1.22+
- PostgreSQL 15+
- `goose` (migrations): `go install github.com/pressly/goose/v3/cmd/goose@latest`
- `sqlc` (codegen): `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

## Setup
```bash
go mod init github.com/einarhh/inventory   # adjust module path as you like
createdb inventory
export DATABASE_URL="postgres://localhost:5432/inventory?sslmode=disable"
goose -dir db/migrations postgres "$DATABASE_URL" up
```

## Status
Bootstrap only. Implemented:
- [x] Core schema (customers, locations, items, files)

Next (build in this order, one increment at a time — see AGENTS.md):
- [ ] `go.mod`, `cmd/server/main.go` that connects to DB and serves `/healthz`
- [ ] sqlc config + first queries (locations)
- [ ] Location CRUD endpoints + tests
- [ ] Item CRUD endpoints + tests
- [ ] File upload + tests
- [ ] Search + statistics
- [ ] Web UI — an installable PWA (React + Vite + TypeScript). A native client is
      deferred until BLE/NFC features are scheduled.

## First task for the agent
> Read AGENTS.md and docs/SPEC.md. Initialize the Go module, add pgx, and create
> cmd/server/main.go that reads DATABASE_URL, connects with a pgxpool, and serves
> GET /healthz returning 200 with the DB ping result. Add one test. Stop there.
