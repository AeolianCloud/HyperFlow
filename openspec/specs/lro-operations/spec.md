## Purpose

定义 Hyperflow 长时间运行操作的持久化、查询与 WebSocket 状态订阅行为。

## Requirements

### Requirement: Operation 记录创建者请求 ID
Operation 记录 SHALL 存储 `creator_request_id` 字段，在 `CreateOperation` 时从调用方的 `context.Context` 中读取并持久化到数据库。该字段用于将 Operation 生命周期内的所有日志归属到创建该 Operation 的原始请求。

#### Scenario: 创建 Operation 时记录 request_id
- **WHEN** `CreateOperation` 被调用，且 context 中包含有效 `request_id`
- **THEN** 新建的 Operation 记录 SHALL 包含该 `request_id` 作为 `creator_request_id`

#### Scenario: 状态变更日志使用创建者 request_id
- **WHEN** Operation 状态从 Running 变为 Succeeded 或 Failed
- **THEN** 写入的日志条目 SHALL 使用 `creator_request_id`，而非触发查询的请求 ID

### Requirement: 通过 WebSocket 订阅操作状态变更
系统 SHALL 提供 WebSocket 端点，客户端连接后服务端周期性查询操作状态并推送事件；操作进入终态后服务端推送最终状态并主动关闭连接。WebSocket 连接建立和关闭 SHALL 分别写入日志。

#### Scenario: 连接时操作仍在进行
- **WHEN** 客户端建立 WebSocket 连接至 `GET /api/pve/operations/{id}/watch`，且操作状态为 Running
- **THEN** 系统 SHALL 完成 WebSocket 握手，并以固定间隔（≤2s）推送 `{"id":"...","status":"Running"}` 消息，直至状态变更

#### Scenario: 操作成功完成时推送并关闭
- **WHEN** 服务端检测到操作状态变为 Succeeded
- **THEN** 系统 SHALL 推送 `{"id":"...","status":"Succeeded","resourceLocation":"..."}` 并以正常关闭帧（code 1000）关闭 WebSocket 连接

#### Scenario: 操作失败时推送并关闭
- **WHEN** 服务端检测到操作状态变为 Failed
- **THEN** 系统 SHALL 推送 `{"id":"...","status":"Failed","error":{"code":"...","message":"..."}}` 并以正常关闭帧（code 1000）关闭 WebSocket 连接

#### Scenario: 操作 ID 不存在时拒绝升级
- **WHEN** 客户端请求 `GET /api/pve/operations/{id}/watch`，且 ID 不存在
- **THEN** 系统 SHALL 返回 HTTP 404 拒绝 WebSocket 升级，响应体为标准错误格式

#### Scenario: 客户端主动断开
- **WHEN** 客户端在操作终态前关闭 WebSocket 连接
- **THEN** 系统 SHALL 停止该连接的轮询协程，不产生资源泄漏

#### Scenario: 重启后操作记录仍可查询
- **WHEN** 服务重启后客户端连接重启前创建的操作
- **THEN** 系统 SHALL 返回该操作的状态记录（从 MySQL 读取）

#### Scenario: WebSocket 连接建立时写日志
- **WHEN** WebSocket 升级握手成功
- **THEN** 系统 SHALL 写入 `event=ws.connect` 日志，包含 `request_id` 和 `operation_id`

#### Scenario: WebSocket 连接关闭时写日志
- **WHEN** WebSocket 连接因任意原因关闭
- **THEN** 系统 SHALL 写入 `event=ws.disconnect` 日志，包含 `request_id` 和 `operation_id`
