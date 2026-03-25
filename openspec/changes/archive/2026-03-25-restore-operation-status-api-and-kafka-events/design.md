## Context

当前 Hyperflow 的异步写接口已经返回 `202 + Operation-Location`，但 `Operation-Location` 指向的 REST 状态接口已被删除，只剩 `GET /api/pve/operations/{id}/watch` WebSocket 端点。与此同时，Hyperflow 已明确定位为门户后端调用的基础设施编排层，而不是直接服务浏览器的实时通知网关。

现有 operation 状态推进依赖 `GetOperation` 按读取时懒查询 PVE，这对单客户端轮询可用，但不适合作为 Kafka 事件发布的基础：如果没有人读取 operation，状态不会推进，事件也不会产生。要支持门户后端通过 Kafka 获知任务完成，Hyperflow 需要具备后台主动跟踪和可靠发布能力。

## Goals / Non-Goals

**Goals:**
- 恢复 `GET /api/pve/operations/{id}` 作为标准 LRO 状态查询接口
- 删除 `GET /api/pve/operations/{id}/watch` WebSocket 端点
- 在无客户端读取的情况下后台跟踪 `Running` operation 并持久化终态
- 在 operation 进入终态后可靠地向 Kafka 发布事件
- 保持 operation 记录为唯一真相源，Kafka 仅作为传播层

**Non-Goals:**
- 直接向浏览器推送消息
- 在 Hyperflow 中引入用户、租户、浏览器会话等前端身份模型
- 提供 exactly-once 跨系统语义；本次仅保证 at-least-once 发布与可补偿查询
- 扩展为通用资源事件总线；本次仅覆盖 operation 终态事件

## Decisions

### D1: `GET /operations/{id}` 作为权威状态查询接口

**决策**：恢复 `GET /api/pve/operations/{id}`，并将其定义为异步操作状态的标准查询接口；删除 `GET /api/pve/operations/{id}/watch`。

**理由**：
- 与现有 `Operation-Location` 头保持一致，补齐 LRO 合同
- 为门户后端提供稳定的补偿查询口，即使 Kafka 通知延迟或丢失也可回查
- 符合 Hyperflow 作为基础设施服务的边界，不再暴露浏览器导向的实时接口

**替代方案**：
- 保留 WebSocket-only：无法与 `Operation-Location` 对齐，且不适合作为服务到服务补偿口
- 同时长期保留 REST + WebSocket：职责重叠，会让 Hyperflow 持续承担浏览器实时集成负担

### D2: 后台 reconciler 负责推进 operation 状态

**决策**：新增后台 reconciler 周期性扫描持久化的 `Running` operations，查询对应 PVE task 状态，并在进入终态时更新 operation 记录。`GET /operations/{id}` 读取持久化结果，不再承担主状态推进职责。

**理由**：
- 无需依赖外部查询即可推进状态，Kafka 事件能够自动产生
- 将“读状态”和“推进状态”分离，减少行为歧义
- 进程重启后可从数据库恢复未完成 operation 的跟踪

**替代方案**：
- 继续仅靠 on-read 懒查询：没有读取就没有状态推进，不适合作为事件源
- 为每个 operation 启动独立 goroutine：短期可行，但重启恢复、并发控制和多实例场景更复杂

### D3: 使用 transactional outbox 发布 Kafka 事件

**决策**：当 operation 从 `Running` 进入 `Succeeded` 或 `Failed` 时，在同一事务内完成两件事：更新 operation 终态；写入一条待发布 outbox 记录。独立 publisher 负责从 outbox 读取待发布事件并发送到 Kafka，成功后标记已发布。

**理由**：
- 避免“operation 已成功更新但 Kafka 事件未发出”的状态裂缝
- 允许 Kafka 暂时不可用时延迟重试，不影响 operation 状态查询
- 使事件发布成为可观测、可恢复的后台流程

**替代方案**：
- 直接在状态更新后同步发 Kafka：实现简单，但一旦进程在更新后崩溃会丢事件
- 以 Kafka 为真相源：不符合本服务的职责边界，也会削弱 `GET /operations/{id}` 的补偿能力

### D4: Kafka 事件采用单 topic、终态事件、JSON 负载

**决策**：向可配置的单个 Kafka topic 发布 operation 终态事件，message key 使用 `operationId`，负载至少包含 `eventId`、`operationId`、`status`、`resourceLocation`、`error`、`providerTaskRef`、`occurredAt`。

**理由**：
- 单 topic 便于门户后端订阅和治理
- `operationId` 作为 key 可保持同一 operation 的顺序语义
- `eventId` 支持消费者去重，满足 at-least-once 交付下的幂等处理

**替代方案**：
- 为成功/失败拆分多个 topic：路由更复杂，收益有限
- 直接推送完整资源对象：耦合门户读模型，且容易导致数据陈旧

### D5: 调整日志模型，围绕 HTTP / operation / event 三条链路

**决策**：删除 WebSocket 生命周期日志要求，新增 Kafka 事件发布日志。保留现有 `http.request`、`pve.call`、`operation.change`；新增 `operation.event.publish`，在成功或失败时均记录，并继续使用 operation 的 `creator_request_id` 作为关联 `request_id`。

**理由**：
- 删除已废弃接口对应的日志噪音
- 在跨服务排障时，可以从创建请求一路追踪到 operation 终态和 Kafka 发布结果

**替代方案**：
- 不记录 Kafka 发布日志：跨服务链路断在 Hyperflow 与 Portal Backend 之间，问题定位困难

## Risks / Trade-offs

- [Risk] Kafka 不可用时事件无法实时到达门户后端 → 使用 outbox 持久化并重试；门户后端仍可通过 `GET /operations/{id}` 补偿查询
- [Risk] reconciler 轮询间隔过长会导致 `GET /operations/{id}` 短暂滞后 → 通过可配置的轮询间隔与批量扫描控制时效性
- [Risk] 多实例部署时多个 reconciler 可能同时轮询同一 operation → 允许重复查询 PVE，但通过条件更新和 outbox 去重确保仅有一个终态转换生效
- [Risk] 删除 WebSocket 是 API breaking change → 当前已确认无客户端依赖；同步更新 Swagger 和文档即可

## Migration Plan

1. 新增 `GET /api/pve/operations/{id}` handler 与文档，删除 `GET /api/pve/operations/{id}/watch`
2. 为 operation 终态事件引入 outbox 持久化表与 Kafka 配置
3. 在应用启动时初始化 reconciler 与 Kafka publisher
4. 门户后端接入 Kafka topic，继续将浏览器 WebSocket 保持在门户层
5. 部署后通过 `GET /operations/{id}` 验证状态查询，通过 Kafka 消费验证终态事件
6. 回滚时可停用 Kafka publisher 并保留 `GET /operations/{id}` 作为唯一状态获取方式；operation 数据不需迁移回滚

## Open Questions

- Kafka topic 的最终命名、认证方式和 broker 地址命名规范需与部署环境约定，但不影响本次接口与能力设计
