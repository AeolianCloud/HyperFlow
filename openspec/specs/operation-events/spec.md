## Purpose

定义 Hyperflow 在 operation 进入终态后向 Kafka 发布事件的格式与可靠投递要求。

## Requirements

### Requirement: 发布操作终态事件到 Kafka
系统 SHALL 在 operation 状态从 `Running` 进入 `Succeeded` 或 `Failed` 时，向配置的 Kafka topic 发布一条 JSON 事件。

#### Scenario: 成功操作发布终态事件
- **WHEN** 某个 operation 状态变为 `Succeeded`
- **THEN** 系统 SHALL 发布包含 `eventId`、`operationId`、`status="Succeeded"`、`resourceLocation`、`providerTaskRef` 和 `occurredAt` 的事件

#### Scenario: 失败操作发布终态事件
- **WHEN** 某个 operation 状态变为 `Failed`
- **THEN** 系统 SHALL 发布包含 `eventId`、`operationId`、`status="Failed"`、`error.code`、`error.message`、`providerTaskRef` 和 `occurredAt` 的事件

### Requirement: 操作事件可靠投递
系统 SHALL 在持久化 operation 终态后持久化待发布事件，并在 Kafka 暂时不可用时重试发布，直到 Kafka 确认或运维介入。

#### Scenario: Kafka 暂时不可用
- **WHEN** operation 已进入终态但 Kafka 当前不可用
- **THEN** 系统 SHALL 保留待发布事件的持久化记录，并在后续重试时继续发布，而不丢失 operation 终态

#### Scenario: 服务重启后恢复未发布事件
- **WHEN** 服务重启时仍存在未发布的 operation 终态事件
- **THEN** 系统 SHALL 从持久化记录中恢复这些事件并继续发布

### Requirement: 事件支持消费者幂等处理
系统 SHALL 使用 `operationId` 作为 Kafka message key，并在事件负载中包含稳定的 `eventId`，以支持消费者去重和幂等处理。

#### Scenario: 发布事件时携带稳定标识
- **WHEN** 系统向 Kafka 发布 operation 终态事件
- **THEN** Kafka message key SHALL 等于该 `operationId`，且事件负载 SHALL 包含唯一的 `eventId`
