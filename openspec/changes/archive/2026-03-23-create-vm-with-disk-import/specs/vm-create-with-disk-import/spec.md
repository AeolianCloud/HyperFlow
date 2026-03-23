## ADDED Requirements

### Requirement: 新建虚拟机并导入磁盘
系统 SHALL 通过调用 PVE `POST /nodes/{node}/qemu` 创建虚拟机，并在创建时通过磁盘参数的 `import-from` 字段导入指定磁盘卷，而非使用 clone API。

#### Scenario: 成功创建虚拟机并导入磁盘
- **WHEN** 客户端发送 `POST /nodes/{node}/vms`，请求体包含 vmid、name、cores、memory、diskSource、storage
- **THEN** 系统调用 PVE `POST /nodes/{node}/qemu`，参数中包含 `virtio0: <storage>:0,import-from=<diskSource>`，返回 202 Accepted 及 PVE 任务 ID

#### Scenario: 指定磁盘接口类型
- **WHEN** 请求体包含 `diskInterface` 字段（如 `scsi0`）
- **THEN** 系统使用指定的接口类型构造磁盘参数（如 `scsi0: <storage>:0,import-from=<diskSource>`）

#### Scenario: diskInterface 默认值
- **WHEN** 请求体未包含 `diskInterface` 字段
- **THEN** 系统使用 `virtio0` 作为默认磁盘接口

#### Scenario: 缺少必填字段
- **WHEN** 请求体缺少 vmid、name、cores、memory、diskSource 或 storage 中的任意一个
- **THEN** 系统返回 400 Bad Request，并说明缺少哪个字段

#### Scenario: PVE 返回错误
- **WHEN** PVE API 返回错误（如 VMID 已存在）
- **THEN** 系统将 PVE 错误映射为对应 HTTP 状态码返回给客户端
