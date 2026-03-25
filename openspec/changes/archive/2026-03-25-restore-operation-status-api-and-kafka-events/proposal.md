## Why

当前 Hyperflow 的异步操作合同不闭合：写接口返回 `Operation-Location: /api/pve/operations/{id}`，但实际仅暴露 `GET /api/pve/operations/{id}/watch` WebSocket 端点。与此同时，Hyperflow 的定位已经明确为 PVE 编排层，实时通知应面向门户后端的内部集成，而不是直接面向浏览器客户端。

## What Changes

- **BREAKING** 删除 `GET /api/pve/operations/{id}/watch` WebSocket 端点
- 恢复 `GET /api/pve/operations/{id}` 作为标准异步操作状态查询接口
- 将 operation 状态更新从“仅在读取时懒查询”调整为“后台主动跟踪 + 持久化更新”
- 新增 Kafka 事件发布能力，在 operation 进入终态时向内部消费者发布状态变化事件
- 新增事件持久化/发布保障机制，避免 operation 状态更新成功但事件丢失
- 新增 Kafka 相关配置项与内部发布链路文档

## Capabilities

### New Capabilities
- `operation-events`: 将异步操作终态事件发布到 Kafka，供门户后端等内部消费者订阅

### Modified Capabilities
- `lro-operations`: 恢复 `GET /api/pve/operations/{id}` 作为标准状态查询接口，并要求 operation 状态可被后台跟踪与持久化
- `operation-watch`: 移除基于 WebSocket 的单操作状态订阅能力
- `request-logging`: 移除 WebSocket 连接生命周期日志要求，并补充 operation 事件发布链路的日志要求

## Impact

- 影响 API：`GET /api/pve/operations/{id}` 恢复；`GET /api/pve/operations/{id}/watch` 删除
- 影响代码：`cmd/handlers.go`、`cmd/main.go`、`internal/operations/*`、`internal/logger/*`
- 可能新增内部组件：operation reconciler、event outbox、Kafka publisher
- 影响配置：新增 Kafka 连接与 topic 配置，移除 `gorilla/websocket` 依赖及相关文档
- 影响系统：门户后端改为消费 Kafka 事件，并通过自身 WebSocket 向浏览器推送任务完成通知
