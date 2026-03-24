## Why

当前系统没有结构化的请求日志，无法追踪单个请求在 HTTP 处理、PVE 出站调用、Operation 状态变更等各层之间的完整链路。需要以 request_id 为线索将所有事件串联并持久化到数据库，供运维排查和审计使用。

## What Changes

- 新增 `logs` 数据库表，存储结构化日志记录
- 新增 Gin 中间件，为每个请求生成唯一 `request_id` 并注入 `gin.Context`
- 新增异步日志写入器（channel + goroutine），不阻塞主链路
- **BREAKING** `pve.Client` 所有方法引入 `context.Context` 参数，用于透传 `request_id`
- `operations` 表新增 `creator_request_id` 字段，记录创建该 Operation 的请求 ID
- WebSocket `/operations/:id/watch` 在连接建立和断开时各写一条日志
- Operation 状态变更（Running → Succeeded/Failed）使用 `creator_request_id` 写日志

## Capabilities

### New Capabilities

- `request-logging`: 结构化请求日志能力，包括 logs 表定义、异步写入器、Gin 中间件、以及各层的日志埋点规范

### Modified Capabilities

- `pve-client`: 所有公开方法签名增加 `context.Context` 第一参数（破坏性变更）
- `lro-operations`: operations 表新增 `creator_request_id` 字段；Operation 状态变更日志使用创建者的 request_id

## Impact

- **数据库**：新增 `logs` 表；`operations` 表新增列
- **代码**：`internal/pve/*.go` 所有 service 方法签名变更；`internal/operations/service.go` 和 `store.go` 变更；`cmd/handlers.go` 所有 handler 变更
- **依赖**：无新增外部依赖，使用 `context` stdlib 和已有 MySQL 连接
