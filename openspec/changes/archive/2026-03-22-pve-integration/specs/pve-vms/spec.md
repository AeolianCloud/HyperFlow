## ADDED Requirements

### Requirement: 查询节点上的虚拟机列表
系统 SHALL 提供接口返回指定节点上所有 QEMU/KVM 虚拟机的列表，包含 VMID、名称、状态、CPU、内存信息。

#### Scenario: 获取虚拟机列表成功
- **WHEN** 客户端发送 `GET /api/pve/nodes/:node/vms`
- **THEN** 系统 SHALL 返回 200 状态码及虚拟机数组，每条记录包含 `vmid`、`name`、`status`、`cpus`、`mem`、`maxmem` 字段

### Requirement: 查询虚拟机详情
系统 SHALL 提供接口返回指定虚拟机的当前状态详情。

#### Scenario: 获取存在的虚拟机详情
- **WHEN** 客户端发送 `GET /api/pve/nodes/:node/vms/:vmid`，且虚拟机存在
- **THEN** 系统 SHALL 返回 200 状态码及该虚拟机详细状态

#### Scenario: 虚拟机不存在时返回 404
- **WHEN** 客户端发送 `GET /api/pve/nodes/:node/vms/:vmid`，且 VMID 不存在
- **THEN** 系统 SHALL 返回 404 状态码及标准错误响应

### Requirement: 启动虚拟机
系统 SHALL 提供接口触发指定虚拟机的启动操作，返回 PVE 任务 ID（UPID）。

#### Scenario: 启动虚拟机成功
- **WHEN** 客户端发送 `POST /api/pve/nodes/:node/vms/:vmid/start`
- **THEN** 系统 SHALL 返回 202 状态码及包含 `upid` 字段的响应体

#### Scenario: 虚拟机已在运行中
- **WHEN** 客户端对已运行的虚拟机发送启动请求
- **THEN** 系统 SHALL 返回 409 状态码及标准错误响应

### Requirement: 停止虚拟机
系统 SHALL 提供接口触发指定虚拟机的停止操作，返回 PVE 任务 ID（UPID）。

#### Scenario: 停止虚拟机成功
- **WHEN** 客户端发送 `POST /api/pve/nodes/:node/vms/:vmid/stop`
- **THEN** 系统 SHALL 返回 202 状态码及包含 `upid` 字段的响应体

### Requirement: 删除虚拟机
系统 SHALL 提供接口删除指定虚拟机，虚拟机 SHALL 处于停止状态方可删除，返回 PVE 任务 ID（UPID）。

#### Scenario: 删除已停止的虚拟机
- **WHEN** 客户端发送 `DELETE /api/pve/nodes/:node/vms/:vmid`，且虚拟机已停止
- **THEN** 系统 SHALL 返回 202 状态码及包含 `upid` 字段的响应体

#### Scenario: 删除运行中的虚拟机被拒绝
- **WHEN** 客户端对运行中的虚拟机发送删除请求
- **THEN** 系统 SHALL 返回 409 状态码及标准错误响应
