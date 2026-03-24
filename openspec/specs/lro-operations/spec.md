## ADDED Requirements

### Requirement: 查询长时间运行操作状态
系统 SHALL 提供端点返回异步操作的当前状态，遵循 Microsoft REST API Guidelines LRO 模式，屏蔽底层 PVE UPID 细节。

#### Scenario: 操作进行中
- **WHEN** 客户端发送 `GET /api/pve/operations/{id}`，且操作仍在执行
- **THEN** 系统 SHALL 返回 200 状态码及 `{"id": "...", "status": "Running"}`

#### Scenario: 操作成功完成
- **WHEN** 客户端发送 `GET /api/pve/operations/{id}`，且 PVE 任务已成功
- **THEN** 系统 SHALL 返回 200 状态码及 `{"id": "...", "status": "Succeeded", "resourceLocation": "..."}`

#### Scenario: 操作失败
- **WHEN** 客户端发送 `GET /api/pve/operations/{id}`，且 PVE 任务失败
- **THEN** 系统 SHALL 返回 200 状态码及 `{"id": "...", "status": "Failed", "error": {"code": "...", "message": "..."}}`

#### Scenario: 操作不存在
- **WHEN** 客户端发送 `GET /api/pve/operations/{id}`，且 ID 不存在
- **THEN** 系统 SHALL 返回 404 状态码及标准错误响应

#### Scenario: 重启后操作记录仍可查询
- **WHEN** 服务重启后客户端查询重启前创建的操作
- **THEN** 系统 SHALL 返回该操作的状态记录（从 MySQL 读取）
