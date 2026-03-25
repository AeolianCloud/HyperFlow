## Purpose

定义 Hyperflow 请求日志、PVE 调用日志、operation 状态日志与 Kafka 发布日志的结构化记录要求。

## Requirements

### Requirement: 请求 ID 生成与传递
系统 SHALL 为每个入站 HTTP 请求生成唯一的 `request_id`（16 字节随机 hex 字符串），并将其存入 `gin.Context`，供整个请求处理链路使用。

#### Scenario: 每个请求获得唯一 ID
- **WHEN** 任意 HTTP 请求到达服务器
- **THEN** 系统 SHALL 生成唯一 `request_id` 并注入当前请求的 `gin.Context`

#### Scenario: 同一请求的所有日志共享同一 ID
- **WHEN** 一个请求触发多条日志（HTTP + PVE 调用 + 后续 Operation 变更 / 事件发布）
- **THEN** 所有日志记录的 `request_id` 字段 SHALL 相同

### Requirement: logs 数据库表
系统 SHALL 维护 `logs` 表，用于持久化所有结构化日志条目。表结构 SHALL 包含以下字段：`id`（自增主键）、`request_id`（VARCHAR 32）、`timestamp`（DATETIME(3)）、`level`（VARCHAR 10，INFO/WARN/ERROR）、`event`（VARCHAR 100）、`method`（VARCHAR 10，可空）、`path`（VARCHAR 500，可空）、`status_code`（INT，可空）、`duration_ms`（INT，可空）、`operation_id`（VARCHAR 32，可空）、`node`（VARCHAR 100，可空）、`message`（TEXT，可空）。`request_id` 和 `timestamp` SHALL 各有索引。

#### Scenario: 表自动创建
- **WHEN** 应用启动
- **THEN** 系统 SHALL 确保 `logs` 表存在，若不存在则自动创建

### Requirement: HTTP 请求日志
系统 SHALL 在每个 HTTP 请求完成后写入一条 `event=http.request` 日志，包含 `request_id`、`method`、`path`、`status_code`、`duration_ms`。

#### Scenario: 成功请求记录
- **WHEN** HTTP 请求处理完成（任意状态码）
- **THEN** 系统 SHALL 写入包含完整字段的 `http.request` 日志条目

### Requirement: PVE 出站调用日志
系统 SHALL 在每次向 PVE API 发出请求时写入 `event=pve.call` 日志，包含 `request_id`、`method`、`path`（PVE API 路径）、`status_code`（PVE 返回）、`duration_ms`。失败时 `level=ERROR`，成功时 `level=INFO`。

#### Scenario: PVE 调用成功
- **WHEN** PVE API 返回 2xx 响应
- **THEN** 系统 SHALL 写入 `level=INFO` 的 `pve.call` 日志，包含状态码和耗时

#### Scenario: PVE 调用失败
- **WHEN** PVE API 返回非 2xx 或网络错误
- **THEN** 系统 SHALL 写入 `level=ERROR` 的 `pve.call` 日志，message 包含错误详情

### Requirement: Operation 状态变更日志
系统 SHALL 在 Operation 状态从 Running 变为 Succeeded 或 Failed 时，使用该 Operation 的 `creator_request_id` 写入 `event=operation.change` 日志，包含 `operation_id`、新状态、以及错误信息（若失败）。

#### Scenario: Operation 成功完成
- **WHEN** Operation 状态更新为 Succeeded
- **THEN** 系统 SHALL 以 `creator_request_id` 为 `request_id` 写入 `level=INFO` 的 `operation.change` 日志

#### Scenario: Operation 执行失败
- **WHEN** Operation 状态更新为 Failed
- **THEN** 系统 SHALL 以 `creator_request_id` 为 `request_id` 写入 `level=ERROR` 的 `operation.change` 日志，message 包含错误码和错误信息

### Requirement: Kafka 操作事件发布日志
系统 SHALL 在向 Kafka 发布 operation 终态事件时写入结构化日志，事件名为 `operation.event.publish`，并使用该 operation 的 `creator_request_id` 作为 `request_id` 关联字段。

#### Scenario: 发布成功时写日志
- **WHEN** 某个 operation 的终态事件被 Kafka 成功确认
- **THEN** 系统 SHALL 写入 `level=INFO` 的 `operation.event.publish` 日志，包含 `request_id`、`operation_id` 和 topic 信息

#### Scenario: 发布失败时写日志
- **WHEN** 某个 operation 的终态事件发布到 Kafka 失败
- **THEN** 系统 SHALL 写入 `level=ERROR` 的 `operation.event.publish` 日志，包含 `request_id`、`operation_id` 和错误详情

### Requirement: 异步写入不阻塞主链路
日志写入 SHALL 通过 buffered channel 异步执行，主请求处理路径不等待数据库写入完成。channel 满时 SHALL 丢弃新日志条目并向 stderr 输出告警，不影响请求处理。

#### Scenario: 正常负载下日志写入
- **WHEN** 系统在正常负载下处理请求
- **THEN** 日志写入 SHALL 不增加 HTTP 响应延迟

#### Scenario: 日志 channel 满
- **WHEN** 异步写入 channel 已满
- **THEN** 系统 SHALL 丢弃新日志条目并向 stderr 输出告警，请求处理 SHALL 正常继续

#### Scenario: 应用关闭时排空日志
- **WHEN** 应用收到关闭信号
- **THEN** 系统 SHALL 等待 channel 中已有日志写入完成（有超时限制），再退出
