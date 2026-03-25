# Repository Guidelines

## Project Structure & Module Organization
`cmd/` contains the HTTP entrypoint, route registration, request logging, and Swagger annotations. Core logic lives under `internal/`: `pve/` wraps Proxmox VE access, `operations/` persists async operation state, `logger/` writes request logs, and `timeutil/` centralizes Shanghai time handling. `docs/` holds generated Swagger artifacts plus API standards, `templates/` stores documentation templates, and `openspec/` tracks proposal/spec/design/task artifacts for spec-driven changes. `bin/` is the local build output directory.

## Build, Test, and Development Commands
Use Go’s standard toolchain:

- `cp .env.example .env` to create local configuration for PVE and MySQL.
- `go run ./cmd` to start the API locally on `:8080`.
- `go build -v ./...` to compile all packages exactly as CI does.
- `go build -o bin/hyperflow ./cmd` to produce the local binary.
- `go test -v ./...` to run the full test suite.
- `swag init -g cmd/main.go` after changing handlers, request/response models, or Swagger annotations.

## Coding Style & Naming Conventions
Follow standard Go formatting with `gofmt`; use tabs, lowercase package names, and exported identifiers only where cross-package access is needed. Keep transport concerns in `cmd/` and domain/service code in `internal/`. Do not hand-edit generated Swagger files such as `docs/docs.go`. For API work, keep success responses unwrapped, return errors as `{"error": ...}`, and preserve async `202` flows with `Operation-Location` headers.

## Testing Guidelines
The repository currently has no committed `_test.go` files, so new work should add tests alongside the package it changes, for example `internal/pve/vms_test.go`. Prefer table-driven tests for handlers, request validation, and failure paths around MySQL or PVE calls. Before opening a PR, run `go test -v ./...` and `go build -v ./...`.

## Commit & Pull Request Guidelines
Recent history follows Conventional Commit prefixes such as `feat:` and `fix:`; keep subjects short and imperative. Pull requests should summarize behavior changes, link the related issue or OpenSpec change, and call out any config, schema, or generated-doc updates. If an API changes, include a Swagger screenshot or a sample request/response and commit regenerated files under `docs/`.

## Security & Workflow Notes
Never commit real values from `.env`; only update `.env.example`. For scoped feature work, record the change under `openspec/changes/` and keep proposal, design, and task files in sync with implementation.

## API设计规范
- 严格按照定义实现API
- 遵循 Microsoft REST API Guidelines（https://github.com/microsoft/api-guidelines/blob/vNext/azure/Guidelines.md）
- 每次修改增加或删除接口或者相关文件必须补全完整注释，不得影响swag文档阅读
- 成功响应直接返回资源对象，不使用 `{"data": ...}` 包装层
- 异步操作返回 202 + `Operation-Location` header，不直接暴露底层任务 ID
- 错误响应统一格式：`{"error": {"code": "PascalCase", "message": "..."}}`