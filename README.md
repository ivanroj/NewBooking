# Coworking booking system

Telegram Mini App for managing university coworking spaces.

- **Backend:** Go 1.25+, REST API (`cmd/server`), SQL migrations (`migrations/`), migration CLI (`cmd/migrate`)
- **Frontend:** Vanilla HTML / CSS / JS in `frontend/` (ESLint + Prettier via npm)
- **Database:** PostgreSQL 15+
- **CI:** GitHub Actions — unit tests, `golangci-lint`, integration tests (`-tags=integration`), frontend ESLint

## Local development

### PostgreSQL via Docker Compose

```bash
docker compose up -d db
docker compose run --rm migrate
```

Compose runs the `migrate` service automatically before `api` when using `docker compose up` for the whole stack.

### Running the API on the host machine

Defaults expect Postgres on **`127.0.0.1:5432`**, DB `coworking_db`, user/password `postgres` / `postgres` (same as Compose port mapping).

```bash
export DB_DSN='postgres://postgres:postgres@127.0.0.1:5432/coworking_db?sslmode=disable'
export PORT=8080
go run ./cmd/migrate -command up -dir migrations
go run ./cmd/server
curl -sf http://127.0.0.1:8080/health
```

### Migrating manually

```bash
export DB_DSN='postgres://postgres:postgres@127.0.0.1:5432/coworking_db?sslmode=disable'
go run ./cmd/migrate -command up   -dir migrations
go run ./cmd/migrate -command down -dir migrations
```

## Tests & quality

```bash
go test ./...
golangci-lint run ./...
```

Integration tests (destructive migrations on the connected database — use a disposable DB):

```bash
export RUN_INTEGRATION=1
export INTEGRATION_DB_DSN="$DB_DSN"   # optional override
go test -tags=integration ./...
```

In GitHub Actions, `CI=true` is set automatically and Postgres is started as a service; no `RUN_INTEGRATION` export is needed.

Frontend (requires [Node.js](https://nodejs.org/) / npm):

```bash
cd frontend
npm install
npm run lint
npm run format
```

## Структура (кратко)

| Path | Purpose |
|------|---------|
| `cmd/server` | HTTP API entrypoint |
| `cmd/migrate` | Apply migrations using `DB_DSN` |
| `internal/` | Config, DB, router, repositories, models |
| `migrations/` | `golang-migrate` SQL files |
| `frontend/` | TMA статика |
