## Why

当前客户端必须轮询 `GET /api/pve/operations/{id}` 来获知任务完成，存在延迟与无效请求。WebSocket 推送可在操作状态变更时即时通知客户端，消除轮询开销。

## What Changes

- 新增 WebSocket 端点 `GET /api/pve/operations/{id}/watch`，客户端连接后服务端推送操作状态变更事件
- 操作进入终态（Succeeded / Failed）后服务端主动关闭连接
- **BREAKING** 删除原 REST 轮询接口 `GET /api/pve/operations/{id}`，客户端须改用 WebSocket 端点

## Capabilities

### New Capabilities
- `operation-watch`: 通过 WebSocket 订阅单个异步操作的状态变更，服务端在操作终态时推送最终状态并关闭连接

### Modified Capabilities
- `lro-operations`：删除 REST 轮询查询 requirement，替换为 WebSocket 订阅

## Impact

- `cmd/handlers.go`：删除 `getOperation` handler，新增 WebSocket 处理器 `watchOperation`
- `cmd/main.go`：更新路由注册，引入 gorilla/websocket 依赖
- `internal/operations/service.go`：新增主动轮询 PVE 并广播状态的逻辑（ticker + channel）
- 新增依赖：`github.com/gorilla/websocket`
- **Breaking change**：所有依赖 `GET /api/pve/operations/{id}` 的客户端须迁移至 WebSocket 端点
