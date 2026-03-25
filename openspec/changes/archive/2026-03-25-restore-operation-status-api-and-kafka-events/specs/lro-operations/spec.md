## ADDED Requirements

### Requirement: 查询长时间运行操作状态
系统 SHALL 提供 `GET /api/pve/operations/{id}` 端点返回异步操作的当前状态，遵循 Microsoft REST API Guidelines 的 LRO 模式，屏蔽底层 PVE task 标识细节。

#### Scenario: 操作进行中
- **WHEN** 客户端发送 `GET /api/pve/operations/{id}`，且操作仍在执行
- **THEN** 系统 SHALL 返回 200 状态码及 `{"id":"...","status":"Running"}`

#### Scenario: 操作成功完成
- **WHEN** 客户端发送 `GET /api/pve/operations/{id}`，且操作已完成且成功
- **THEN** 系统 SHALL 返回 200 状态码及 `{"id":"...","status":"Succeeded","resourceLocation":"..."}`

#### Scenario: 操作失败
- **WHEN** 客户端发送 `GET /api/pve/operations/{id}`，且操作已完成且失败
- **THEN** 系统 SHALL 返回 200 状态码及 `{"id":"...","status":"Failed","error":{"code":"...","message":"..."}}`

#### Scenario: 操作不存在
- **WHEN** 客户端发送 `GET /api/pve/operations/{id}`，且 ID 不存在
- **THEN** 系统 SHALL 返回 404 状态码及标准错误响应

#### Scenario: 重启后操作记录仍可查询
- **WHEN** 服务重启后客户端查询重启前创建的操作
- **THEN** 系统 SHALL 返回该操作的持久化状态记录

### Requirement: 后台跟踪运行中的操作
系统 SHALL 在无客户端查询的情况下持续跟踪 `Running` 状态的 operation，并在底层 PVE 任务进入终态后持久化更新 operation 状态。

#### Scenario: 无客户端读取时仍更新终态
- **WHEN** 一个 operation 处于 `Running` 状态且底层 PVE 任务已经完成
- **THEN** 系统 SHALL 在后台将该 operation 更新为 `Succeeded` 或 `Failed`

#### Scenario: 服务重启后恢复跟踪
- **WHEN** 服务启动时存在重启前遗留的 `Running` operations
- **THEN** 系统 SHALL 恢复对这些 operation 的后台跟踪，直到进入终态

## REMOVED Requirements

### Requirement: 通过 WebSocket 订阅操作状态变更
**Reason**: Hyperflow 不再提供浏览器导向的 operation WebSocket 订阅能力；实时状态传播改由 Kafka 事件供门户后端消费。
**Migration**: 使用 `GET /api/pve/operations/{id}` 查询操作状态，并通过门户后端消费 Kafka operation 事件获取实时通知。
