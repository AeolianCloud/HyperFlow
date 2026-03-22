## ADDED Requirements

### Requirement: 查询存储池列表
系统 SHALL 提供接口返回 PVE 集群中所有存储池的列表，包含存储名称、类型、总容量、已用容量、可用容量及状态。

#### Scenario: 获取存储池列表成功
- **WHEN** 客户端发送 `GET /api/pve/storage`
- **THEN** 系统 SHALL 返回 200 状态码及存储池数组，每条记录包含 `storage`、`type`、`total`、`used`、`avail`、`active` 字段

#### Scenario: PVE 不可达时返回错误
- **WHEN** PVE 服务器无法连接
- **THEN** 系统 SHALL 返回 502 状态码及标准错误响应
