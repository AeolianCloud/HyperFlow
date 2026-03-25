## ADDED Requirements

### Requirement: Kafka 操作事件发布日志
系统 SHALL 在向 Kafka 发布 operation 终态事件时写入结构化日志，事件名为 `operation.event.publish`，并使用该 operation 的 `creator_request_id` 作为 `request_id` 关联字段。

#### Scenario: 发布成功时写日志
- **WHEN** 某个 operation 的终态事件被 Kafka 成功确认
- **THEN** 系统 SHALL 写入 `level=INFO` 的 `operation.event.publish` 日志，包含 `request_id`、`operation_id` 和 topic 信息

#### Scenario: 发布失败时写日志
- **WHEN** 某个 operation 的终态事件发布到 Kafka 失败
- **THEN** 系统 SHALL 写入 `level=ERROR` 的 `operation.event.publish` 日志，包含 `request_id`、`operation_id` 和错误详情

## REMOVED Requirements

### Requirement: WebSocket 连接生命周期日志
**Reason**: `GET /api/pve/operations/{id}/watch` WebSocket 端点被移除，不再存在对应连接生命周期。
**Migration**: 通过 `http.request`、`operation.change` 和 `operation.event.publish` 日志跟踪 operation 的查询、状态变化和事件发布链路。
