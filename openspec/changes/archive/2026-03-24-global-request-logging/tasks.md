## 1. 数据库变更

- [x] 1.1 在 `operations` 包的 `store.go` 中为 `operations` 表添加 `creator_request_id` 字段（`CreateTable` DDL 和 `Insert` 语句）
- [x] 1.2 在 `internal/logger` 包中实现 `CreateTable` 方法，创建 `logs` 表（含所有字段和索引）

## 2. 日志核心模块

- [x] 2.1 创建 `internal/logger/logger.go`，定义 `Entry` 结构体（所有日志字段）和 `Logger` 接口
- [x] 2.2 实现 `MySQLLogger`：buffered channel（容量 1000）+ 单一后台 goroutine 异步写入
- [x] 2.3 实现 channel 满时丢弃并向 stderr 输出告警的逻辑
- [x] 2.4 实现 `Shutdown(ctx context.Context)` 方法，等待 channel 排空或超时后退出

## 3. PVE Client 引入 context.Context

- [x] 3.1 修改 `pve/client.go` 的 `do()` 方法，增加 `ctx context.Context` 参数，使用 `http.NewRequestWithContext`
- [x] 3.2 修改 `pve/client.go` 的 `doMultipart()` 方法，增加 `ctx context.Context` 参数
- [x] 3.3 在 `pve/client.go` 的 `do()` 中调用 logger，写入 `pve.call` 日志（含 request_id、method、path、status_code、duration_ms）
- [x] 3.4 修改 `pve/nodes.go` 所有公开方法签名，增加 `ctx context.Context` 第一参数并透传
- [x] 3.5 修改 `pve/vms.go` 所有公开方法签名，增加 `ctx context.Context` 第一参数并透传
- [x] 3.6 修改 `pve/storage.go` 所有公开方法签名，增加 `ctx context.Context` 第一参数并透传

## 4. Operations 层变更

- [x] 4.1 在 `operations/store.go` 的 `Operation` 结构体中添加 `CreatorRequestID string` 字段
- [x] 4.2 更新 `operations/store.go` 的 `Insert`、`GetByID` 方法以处理 `creator_request_id` 字段
- [x] 4.3 修改 `operations/service.go` 的 `CreateOperation` 方法，从 `context.Context` 中读取 request_id 并存入 Operation
- [x] 4.4 在 `operations/service.go` 的 `GetOperation` 状态变更逻辑中，使用 `creator_request_id` 调用 logger 写入 `operation.change` 日志

## 5. Gin 中间件与 Handler 改造

- [x] 5.1 在 `cmd/main.go` 中初始化 `MySQLLogger` 并注入到 Gin、pve Client 和 operations Service
- [x] 5.2 创建 Gin 中间件函数，为每个请求生成 `request_id`（16 字节随机 hex）并存入 `gin.Context`
- [x] 5.3 在中间件中注册 `c.Next()` 后的回调，写入 `http.request` 日志（method、path、status_code、duration_ms）
- [x] 5.4 在 `cmd/main.go` 中替换 `gin.Default()` 为 `gin.New()` 并挂载新中间件（移除内置 Logger）
- [x] 5.5 修改 `cmd/handlers.go` 所有 handler，从 `gin.Context` 提取 request_id 并通过 `context.WithValue` 构造 ctx，传入 pve service 调用

## 6. WebSocket 日志埋点

- [x] 6.1 在 `cmd/handlers.go` 的 `watchOperation` 中，WebSocket 升级成功后写入 `ws.connect` 日志
- [x] 6.2 在 `watchOperation` 连接关闭时（defer 或退出循环后）写入 `ws.disconnect` 日志

## 7. 应用关闭处理

- [x] 7.1 在 `cmd/main.go` 中监听 OS 信号（SIGINT/SIGTERM），收到信号后调用 `logger.Shutdown` 完成日志排空再退出
