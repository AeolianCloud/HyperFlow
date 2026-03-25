## REMOVED Requirements

### Requirement: 通过 WebSocket 订阅操作状态变更
**Reason**: Hyperflow 作为基础设施编排层不再直接提供面向客户端的 operation WebSocket 订阅接口。
**Migration**: 使用 `GET /api/pve/operations/{id}` 进行状态查询，并由门户后端消费 Kafka operation 事件后再向浏览器推送实时通知。
