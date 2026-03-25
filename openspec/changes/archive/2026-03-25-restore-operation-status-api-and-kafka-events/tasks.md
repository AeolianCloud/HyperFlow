## 1. 配置与依赖

- [x] 1.1 为 Hyperflow 引入 Kafka 客户端依赖，并在 `.env.example` 中新增 broker、topic 等配置项
- [x] 1.2 从依赖和文档中移除 `gorilla/websocket` 及 `GET /api/pve/operations/{id}/watch` 相关说明

## 2. Operation 状态推进

- [x] 2.1 扩展 `internal/operations` 持久化层，增加终态事件 outbox 存储及必要的建表逻辑
- [x] 2.2 重构 operation 状态转换逻辑，使终态更新与 outbox 写入在同一事务内完成
- [x] 2.3 实现后台 reconciler，周期性扫描 `Running` operations、查询 PVE task 状态并推进终态
- [x] 2.4 在应用启动和关闭流程中接入 reconciler 生命周期管理

## 3. Kafka 事件发布

- [x] 3.1 定义 operation 终态 Kafka 事件负载、message key 和序列化格式
- [x] 3.2 实现 outbox publisher，将待发布事件发送到 Kafka 并在成功后标记已发布
- [x] 3.3 为 Kafka 发布成功与失败补充 `operation.event.publish` 结构化日志

## 4. API 与文档调整

- [x] 4.1 在 `cmd/handlers.go` 中恢复 `GET /api/pve/operations/{id}` handler，并返回标准 operation 状态响应
- [x] 4.2 在路由注册中删除 `GET /api/pve/operations/{id}/watch`，清理对应 handler 与模型
- [x] 4.3 更新 Swagger 注释与生成文档，确保 `Operation-Location` 指向可查询的 REST 状态接口

## 5. 验证

- [x] 5.1 为 `GET /api/pve/operations/{id}`、后台 reconciler 和 outbox publisher 添加测试覆盖
- [x] 5.2 验证 Kafka 不可用、服务重启后恢复未发布事件等失败路径
- [x] 5.3 运行 `go test -v ./...`、`go build -v ./...`，并检查 OpenSpec/Swagger 产物一致性
