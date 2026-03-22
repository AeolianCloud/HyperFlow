## ADDED Requirements

### Requirement: 查询集群节点列表
系统 SHALL 提供接口返回 PVE 集群中所有节点的列表及其基本状态信息（节点名、在线状态、CPU 使用率、内存使用率）。

#### Scenario: 获取节点列表成功
- **WHEN** 客户端发送 `GET /api/pve/nodes`
- **THEN** 系统 SHALL 返回 200 状态码及节点数组，每个节点包含 `node`、`status`、`cpu`、`mem`、`maxmem` 字段

#### Scenario: PVE 不可达时返回错误
- **WHEN** PVE 服务器无法连接
- **THEN** 系统 SHALL 返回 502 状态码及标准错误响应

### Requirement: 查询单个节点详情
系统 SHALL 提供接口返回指定节点的详细状态信息，包括 CPU、内存、磁盘、运行时间等。

#### Scenario: 获取存在的节点详情
- **WHEN** 客户端发送 `GET /api/pve/nodes/:node`，且节点存在
- **THEN** 系统 SHALL 返回 200 状态码及该节点的详细信息

#### Scenario: 节点不存在时返回 404
- **WHEN** 客户端发送 `GET /api/pve/nodes/:node`，且节点不存在
- **THEN** 系统 SHALL 返回 404 状态码及标准错误响应
