## Context

当前 Hyperflow 使用 Microsoft LRO 规范：异步操作（VM 创建、启停、删除）返回 202 + `Operation-Location`，客户端需反复轮询 `GET /api/pve/operations/{id}` 直到状态脱离 Running。服务端已有懒查询机制（`Service.GetOperation` 按需查询 PVE UPID）。框架为 Gin，无现有 WebSocket 基础设施。

## Goals / Non-Goals

**Goals:**
- 新增 `GET /api/pve/operations/{id}/watch` WebSocket 端点，在操作状态变更时即时推送
- 操作进入终态后服务端主动关闭连接（code 1000）
- 保持原 REST 轮询接口不变

**Non-Goals:**
- 多操作批量订阅
- 身份鉴权（与现有 REST 接口保持一致，当前无鉴权）
- 操作历史回放

## Decisions

### D1: 使用 gorilla/websocket
**决策**：引入 `github.com/gorilla/websocket`。
**理由**：Gin 无内置 WebSocket 支持；gorilla/websocket 是 Go 生态事实标准，API 稳定，无额外运行时依赖。
**替代方案**：`nhooyr.io/websocket` 更轻量，但社区熟悉度低；直接用 `net/http` 手动升级协议复杂度高且易出错。

### D2: 服务端定时轮询 PVE，不引入消息队列
**决策**：每个 WebSocket 连接独立启动 goroutine，以 1s ticker 调用现有 `Service.GetOperation` 查询状态并推送。
**理由**：当前并发操作数量有限（单集群运维场景），轮询模型简单可控；`GetOperation` 已封装 PVE 懒查询与 DB 更新，可直接复用。
**替代方案**：在 `Service` 层引入发布/订阅（channel fan-out）可减少对 PVE 的重复查询，但增加架构复杂度，现阶段不必要。

### D3: 连接建立前先校验操作存在性
**决策**：在 WebSocket 升级前调用 `Service.GetOperation`，若操作不存在返回 HTTP 404（不升级连接）；若操作已为终态则推送一条终态消息后立即关闭。
**理由**：避免升级后立即关闭带来的协议噪音；HTTP 层错误更易被标准客户端捕获。

### D4: 推送消息格式与 REST 响应一致
**决策**：WebSocket 消息为 JSON，结构与 `OperationResponse` 相同（`id`, `status`, `resourceLocation`, `error`）。
**理由**：客户端无需为 WS 实现独立的反序列化逻辑。

## Risks / Trade-offs

- **PVE 查询放大**：N 个客户端订阅同一操作时会产生 N 次/s 的 PVE 查询。当前场景下可接受；如需优化可改为 fan-out 广播。→ 文档记录限制，后续按需优化。
- **连接泄漏**：客户端异常断开时 goroutine 需通过 `context` 或 write 错误检测退出。→ 使用 `ctx, cancel` + `conn.SetCloseHandler` 确保协程退出。
- **Gin 路由冲突**：`/:id` 与 `/:id/watch` 在同一 RouterGroup 下共存。→ Gin 的静态路径优先于参数路径，`/watch` 作为独立注册路由无冲突。

## Migration Plan

1. `go get github.com/gorilla/websocket` 更新 go.mod/go.sum
2. 在 `cmd/handlers.go` 新增 `watchOperation` handler
3. 在 `registerOperationsRoutes` 中注册 `GET /:id/watch`
4. 更新 swag 注释，重新生成 `docs/`
5. 无数据库变更，无需迁移脚本
6. 回滚：移除路由注册与 handler，不影响现有接口

## Open Questions

- 是否需要对 WebSocket 连接数设置上限（如每操作最多 N 个订阅者）？当前暂不限制。
