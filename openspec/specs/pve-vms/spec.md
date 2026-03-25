## Purpose

定义 Hyperflow 对 PVE 虚拟机资源的查询、创建和生命周期操作行为。

## Requirements

### Requirement: 新建虚拟机（含 ciPackages）
当请求包含 `ciPackages` 时，系统 SHALL 生成包含 `hostname`、`fqdn` 和 `preserve_hostname: false` 字段的 cloud-init user-data YAML，确保虚拟机首次开机后主机名与虚拟机 `name` 一致。

#### Scenario: ciPackages 非空时 user-data 包含 hostname
- **WHEN** 客户端发送 `POST /api/pve/nodes/:node/vms`，且 `ciPackages` 非空
- **THEN** 上传的 cloud-init user-data YAML SHALL 包含 `hostname: <name>`、`fqdn: <name>` 和 `preserve_hostname: false`

#### Scenario: ciPackages 为空时行为不变
- **WHEN** 客户端发送 `POST /api/pve/nodes/:node/vms`，且 `ciPackages` 为空
- **THEN** 系统 SHALL 使用 PVE 原生 CloudInit 参数，不生成自定义 user-data，行为与修改前一致

### Requirement: 查询节点上的虚拟机列表
系统 SHALL 提供接口返回指定节点上所有 QEMU/KVM 虚拟机的列表，包含 VMID、名称、状态、CPU、内存信息。

#### Scenario: 获取虚拟机列表成功
- **WHEN** 客户端发送 `GET /api/pve/nodes/:node/vms`
- **THEN** 系统 SHALL 返回 200 状态码及虚拟机数组（直接返回，不包装 `data` 字段），每条记录包含 `vmid`、`name`、`status`、`cpus`、`mem`、`maxmem` 字段

### Requirement: 查询虚拟机详情
系统 SHALL 提供接口返回指定虚拟机的当前状态详情。

#### Scenario: 获取存在的虚拟机详情
- **WHEN** 客户端发送 `GET /api/pve/nodes/:node/vms/:vmid`，且虚拟机存在
- **THEN** 系统 SHALL 返回 200 状态码及该虚拟机详细状态（直接返回，不包装 `data` 字段）

#### Scenario: 虚拟机不存在时返回 404
- **WHEN** 客户端发送 `GET /api/pve/nodes/:node/vms/:vmid`，且 VMID 不存在
- **THEN** 系统 SHALL 返回 404 状态码及标准错误响应

### Requirement: 启动虚拟机
系统 SHALL 提供接口触发指定虚拟机的异步启动操作，返回标准 LRO Operation-Location header。

#### Scenario: 启动虚拟机成功
- **WHEN** 客户端发送 `POST /api/pve/nodes/:node/vms/:vmid/start`
- **THEN** 系统 SHALL 返回 202 状态码，无响应体，响应头包含 `Operation-Location: /api/pve/operations/{id}`

#### Scenario: 虚拟机已在运行中
- **WHEN** 客户端对已运行的虚拟机发送启动请求
- **THEN** 系统 SHALL 返回 409 状态码及标准错误响应

### Requirement: 停止虚拟机
系统 SHALL 提供接口触发指定虚拟机的异步停止操作，返回标准 LRO Operation-Location header。

#### Scenario: 停止虚拟机成功
- **WHEN** 客户端发送 `POST /api/pve/nodes/:node/vms/:vmid/stop`
- **THEN** 系统 SHALL 返回 202 状态码，无响应体，响应头包含 `Operation-Location: /api/pve/operations/{id}`

### Requirement: 删除虚拟机
系统 SHALL 提供接口异步删除指定虚拟机，虚拟机 SHALL 处于停止状态方可删除，返回标准 LRO Operation-Location header。

#### Scenario: 删除已停止的虚拟机
- **WHEN** 客户端发送 `DELETE /api/pve/nodes/:node/vms/:vmid`，且虚拟机已停止
- **THEN** 系统 SHALL 返回 202 状态码，无响应体，响应头包含 `Operation-Location: /api/pve/operations/{id}`

#### Scenario: 删除运行中的虚拟机被拒绝
- **WHEN** 客户端对运行中的虚拟机发送删除请求
- **THEN** 系统 SHALL 返回 409 状态码及标准错误响应

### Requirement: 新建虚拟机并导入磁盘
系统 SHALL 提供接口异步创建虚拟机，返回标准 LRO Operation-Location header 及 Location header。

#### Scenario: 新建虚拟机成功
- **WHEN** 客户端发送 `POST /api/pve/nodes/:node/vms`，且参数合法
- **THEN** 系统 SHALL 返回 202 状态码，无响应体，响应头包含 `Operation-Location: /api/pve/operations/{id}` 及 `Location: /api/pve/nodes/{node}/vms/{vmid}`
