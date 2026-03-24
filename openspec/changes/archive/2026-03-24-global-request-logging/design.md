## Context

当前 Hyperflow 使用 `gin.Default()` 内置的访问日志（stdout），无结构化字段，无法按请求 ID 聚合日志，无法跨层（HTTP → PVE 调用 → Operation 变更）追踪单一请求链路。`pve.Client` 的所有方法不携带 `context.Context`，request_id 无法自然地向下传递。

已有基础设施：
- Gin HTTP 框架
- MySQL（已有 `operations` 表）
- WebSocket（gorilla/websocket）

## Goals / Non-Goals

**Goals:**
- 每个 HTTP 请求生成唯一 `request_id`，贯穿整个处理链路
- HTTP 请求、PVE 出站调用、WebSocket 连接/断开、Operation 状态变更均写入 `logs` 表
- Operation 使用其创建者的 `request_id`，而非触发查询者的
- 日志异步写入，不阻塞主链路

**Non-Goals:**
- 分布式 tracing（OpenTelemetry 等）
- 日志查询 API
- 日志轮转或归档
- 对外暴露 request_id 给客户端

## Decisions

### 1. request_id 生成位置：Gin 中间件

在 Gin 中间件中生成 request_id 并存入 `gin.Context`，所有 handler 从 context 取用。

替代方案：在各 handler 内部单独生成 → 无法统一，容易遗漏。

### 2. request_id 向下传递：context.Context

为 `pve.Client.do()` 及所有上层 service 方法增加 `context.Context` 第一参数，从中读取 request_id 写日志。这是 Go 惯用做法，也为将来接入 tracing 留好接口。

替代方案：全局变量或 goroutine-local → Go 中不可行且危险。

### 3. Operation 日志归属：creator_request_id

`operations` 表新增 `creator_request_id` 列，在 `CreateOperation` 时写入。后续 `GetOperation` 触发状态变更时，使用该值写日志而非当前请求的 request_id。

这样同一个 Operation 从创建到终态的所有日志都挂在同一个 request_id 下，方便聚合查询。

### 4. 异步写入：buffered channel + 单一 goroutine

```
Logger.Log(entry) → channel (buffer=1000) → goroutine → MySQL
```

- buffer 满时丢弃（记 stderr），不阻塞请求
- 进程退出时 graceful drain（context cancel + 超时）

替代方案：同步写入 → 增加请求延迟；消息队列 → 引入额外依赖。

### 5. 不引入新日志库

使用 `log/slog`（Go 1.21 stdlib）格式化日志条目，`database/sql` 写入，无新依赖。

### 6. WebSocket 日志粒度：只记连接和断开

每次 tick 写日志会产生大量低价值记录。只在 upgrade 成功和连接关闭时各写一条，并记录 operation_id。

## Risks / Trade-offs

- **破坏性 API 变更**：所有 pve service 方法签名变化，handlers.go 改动量较大 → 单次 PR 完成，无向后兼容需求
- **异步丢日志**：channel 满时新日志被丢弃 → buffer=1000 在正常负载下充裕；可通过监控 stderr 感知
- **MySQL 写压力**：高 QPS 时日志表增长快 → 当前项目规模可接受；后续可加索引或分区
- **context.Context 改动量大**：涉及所有 pve/*.go 和 handlers.go → 改动规整，单一模式，测试覆盖可验证

## Migration Plan

1. 创建 `logs` 表（新增，无影响）
2. `operations` 表 ALTER ADD COLUMN `creator_request_id`（nullable，存量数据为 NULL，无影响）
3. 部署新代码（一次性切换，无灰度需求）

回滚：还原代码部署；`logs` 表和新列可保留（不影响功能）。

## Open Questions

- `logs` 表是否需要保留策略（如按时间分区、定期清理）？当前暂不实现，留作后续。
