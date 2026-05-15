# AGENTS.md

## Project Structure

```
cmd/            — Go entrypoint (main.go), Gin routes, handlers, swagger annotations
internal/
  pve/          — Proxmox VE API client + nodes/vms/storage services
  operations/   — async LRO operations store (MySQL) + outbox event publisher (Kafka)
  logger/       — async MySQL request log writer
  timeutil/     — fixed Asia/Shanghai timezone
docs/           — generated Swagger (don't hand-edit)
openspec/       — spec-driven proposal/design/task artifacts
bin/            — local build output
```

## Commands

| Action | Command |
|---|---|
| Serve | `go run ./cmd` (port 8080) |
| Build all | `go build -v ./...` |
| Build binary | `go build -o bin/hyperflow ./cmd` |
| Test all | `go test -v ./...` |
| Test single | `go test -v ./internal/operations/...` |
| Regenerate Swagger | `swag init -g cmd/main.go` (install: `go install github.com/swaggo/swag/cmd/swag@latest`) |
| Config | `cp .env.example .env` |

CI order: `go build -v ./...` → `go test -v ./...`

## Required Env

`PVE_HOST`, `PVE_TOKEN_ID`, `PVE_TOKEN_SECRET`, `MYSQL_DSN`, `KAFKA_BROKERS`, `KAFKA_OPERATION_EVENTS_TOPIC`

Optional: `PVE_INSECURE` (skip SSL verify), `PVE_SNIPPETS_WEBDAV_URL` / `WEBDAV_USER` / `WEBDAV_PASSWORD` (cloud-init snippet upload)

`.env` loading is lenient (godotenv, missing file ignored).  
`MYSQL_DSN` is auto-normalized: forced `parseTime=true`, `time_zone='+08:00'`, `Loc=Asia/Shanghai`.

## Architecture

- **Gin** on `:8080`, all routes under `/api/pve`, Swagger at `/swagger/*any`
- **PVE Client**: API Token auth (`PVEAPIToken=...`), auto-unwraps `{"data": ...}`, logs every call as `pve.call` to MySQL
- **Operations outbox**: `Reconciler` polls PVE task status → updates MySQL + writes `operation_events_outbox`; `OutboxPublisher` drains outbox → Kafka. Both are background goroutines with graceful shutdown
- **Time**: always `Asia/Shanghai` via `timeutil` package
- **Responses**: success unwrapped (no `{"data": ...}`), error `{"error": {"code":"PascalCase","message":"..."}}`
- **Async**: 202 + `Operation-Location` header
- **Graceful shutdown** order: cancel app context → HTTP Shutdown → drain Reconciler + OutboxPublisher → drain MySQLLogger
- **request_id** (32 hex chars) generated per request, propagated via `requestContextFromGin()` into context; operation IDs are 16 hex chars

## Testing

Tests use in-memory fakes (`fakeStore`, `captureLogger`, `fakeProducer`, `fakeQuerier`) — no MySQL/PVE/Kafka needed. Prefer table-driven tests. Test files: `cmd/handlers_test.go`, `internal/operations/*_test.go`.

## Conventions

- Standard Go: `gofmt`, tabs, lowercased packages, exported only when cross-package
- Conventional Commits: `feat:`, `fix:`, `docs:`, `chore:`
- Transport in `cmd/`, domain in `internal/`
- Never hand-edit `docs/docs.go`, `swagger.json`, `swagger.yaml`
- Never commit real `.env` values
- OpenSpec changes under `openspec/changes/`; archive with date prefix
