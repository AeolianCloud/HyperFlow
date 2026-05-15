## Purpose

定义 HyperFlow 对已有 PVE 虚拟机的数据盘生命周期管理行为（查询、挂载、卸载、删除）。

## ADDED Requirements

### Requirement: 查询 VM 磁盘列表

系统 SHALL 提供接口返回指定虚拟机的所有磁盘信息，包含接口名、大小、存储位置、磁盘格式等。

#### Scenario: 查询运行中虚拟机的磁盘列表

- **WHEN** 客户端发送 `GET /api/pve/nodes/{node}/vms/{vmid}/disks`
- **THEN** 系统 SHALL 返回 200 状态码及磁盘数组，每条记录包含 `diskId`（如 `scsi1`）、`size`（GB）、`storage`、`format` 等字段

#### Scenario: 虚拟机不存在时返回 404

- **WHEN** 客户端发送 `GET /api/pve/nodes/{node}/vms/{vmid}/disks`，且 VMID 不存在
- **THEN** 系统 SHALL 返回 404 状态码及标准错误响应

### Requirement: 挂载数据盘到已有 VM

系统 SHALL 提供接口给指定虚拟机挂载一块新的空数据盘，支持热插拔。请求体包含 `size`（必填，GB）和 `storage`（必填，存储池名称）。系统自动分配下一个可用的 scsi 接口名。

#### Scenario: 挂载数据盘成功

- **WHEN** 客户端发送 `POST /api/pve/nodes/{node}/vms/{vmid}/disks`，请求体包含 `size: 100`、`storage: "local-lvm"`
- **THEN** 系统 SHALL 返回 202 状态码，响应头包含 `Operation-Location`，异步完成后 VM 新增一块 100GB 的数据盘

#### Scenario: 缺少 size 字段

- **WHEN** 客户端发送的请求体缺少 `size` 字段
- **THEN** 系统 SHALL 返回 400 状态码及标准错误响应

#### Scenario: 缺少 storage 字段

- **WHEN** 客户端发送的请求体缺少 `storage` 字段
- **THEN** 系统 SHALL 返回 400 状态码及标准错误响应

#### Scenario: 并发挂载时不会冲突

- **WHEN** 两个客户端同时向同一 VM 发送挂载磁盘请求
- **THEN** 两个请求串行执行，第二个请求等到第一个完成后自动分配下一个接口名

### Requirement: 卸载数据盘

系统 SHALL 提供接口从 VM 配置中移除指定数据盘，保留存储卷不被删除。

#### Scenario: 卸载数据盘成功

- **WHEN** 客户端发送 `DELETE /api/pve/nodes/{node}/vms/{vmid}/disks/{diskId}`，且不传 purge 参数
- **THEN** 系统 SHALL 返回 202 状态码及 `Operation-Location` header，异步完成后该磁盘从 VM 配置中移除，存储卷保留

### Requirement: 卸载并销毁数据盘

系统 SHALL 提供接口从 VM 配置中移除指定数据盘的同时，销毁对应的存储卷。

#### Scenario: 卸载并销毁数据盘成功

- **WHEN** 客户端发送 `DELETE /api/pve/nodes/{node}/vms/{vmid}/disks/{diskId}?purge=true`
- **THEN** 系统 SHALL 返回 202 状态码及 `Operation-Location` header，异步完成后该磁盘从 VM 配置中移除且存储卷被销毁

#### Scenario: 卸载不存在的磁盘

- **WHEN** 客户端发送 `DELETE .../disks/scsi99`，且该接口名不存在
- **THEN** 系统 SHALL 返回 404 状态码及标准错误响应

### Requirement: 并发安全

系统 SHALL 确保同一 VM 的挂载和卸载磁盘操作互斥执行，防止接口名冲突。

#### Scenario: 加盘和拆盘同时发生

- **WHEN** 一个客户端挂载磁盘的同时另一个客户端对同一 VM 卸载磁盘
- **THEN** 两个操作串行执行，不会出现接口名分配冲突

#### Scenario: 锁超时返回 409

- **WHEN** 一个磁盘操作持有锁超过预期时间，另一个请求等待锁超时
- **THEN** 系统 SHALL 返回 409 状态码，错误信息包含 "Another disk operation is in progress on this VM"
