## REMOVED Requirements

### Requirement: 查询长时间运行操作状态
**Reason**: 替换为 WebSocket 订阅端点，客户端可实时获得状态推送，无需轮询。
**Migration**: 使用 `GET /api/pve/operations/{id}/watch`（WebSocket）替代原 `GET /api/pve/operations/{id}` REST 接口。

## ADDED Requirements

### Requirement: 通过 WebSocket 订阅操作状态变更
系统 SHALL 提供 WebSocket 端点，客户端连接后服务端周期性查询操作状态并推送事件；操作进入终态后服务端推送最终状态并主动关闭连接。

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
