## MODIFIED Requirements

### Requirement: 新建虚拟机并导入磁盘
系统 SHALL 通过调用 PVE `POST /nodes/{node}/qemu` 创建虚拟机，并在创建时通过磁盘参数的 `import-from` 字段导入指定磁盘卷，而非使用 clone API。当请求体包含任意 CloudInit 字段时，系统 SHALL 同时附加 CloudInit 驱动盘（`ide2: <storage>:cloudinit`）及对应 CloudInit 配置参数。

#### Scenario: 成功创建虚拟机并导入磁盘
- **WHEN** 客户端发送 `POST /nodes/{node}/vms`，请求体包含 vmid、name、cores、memory、diskSource、storage，不包含任何 CloudInit 字段
- **THEN** 系统调用 PVE `POST /nodes/{node}/qemu`，参数中包含 `virtio0: <storage>:0,import-from=<diskSource>`，不包含 `ide2` 及 CloudInit 参数，返回 202 Accepted 及 PVE 任务 ID

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

## ADDED Requirements

### Requirement: 创建虚拟机时支持 CloudInit 配置
系统 SHALL 在创建虚拟机请求体中接受可选的 CloudInit 配置参数；当请求体包含至少一个 CloudInit 字段时，系统 SHALL 向 PVE 请求附加 CloudInit 驱动盘及对应配置。

#### Scenario: 携带 CloudInit 用户名和密码创建虚拟机
- **WHEN** 客户端发送 `POST /nodes/{node}/vms`，请求体包含必填字段及 `ciUser`、`ciPassword`
- **THEN** 系统调用 PVE `POST /nodes/{node}/qemu` 时，参数中包含 `ide2: <storage>:cloudinit`、`ciuser: <ciUser>`、`cipassword: <ciPassword>`，返回 202 Accepted

#### Scenario: 携带 SSH 公钥创建虚拟机
- **WHEN** 客户端发送请求体包含 `sshKeys` 字段（SSH 公钥字符串）
- **THEN** 系统对 `sshKeys` 进行 URL 编码后作为 `sshkeys` 参数传递给 PVE，并附加 CloudInit 驱动盘

#### Scenario: 携带静态 IP 配置创建虚拟机
- **WHEN** 客户端请求体包含 `ipConfig0` 字段（如 `ip=192.168.1.100/24,gw=192.168.1.1`）
- **THEN** 系统将 `ipConfig0` 的值作为 `ipconfig0` 参数传递给 PVE，并附加 CloudInit 驱动盘

#### Scenario: 携带 DHCP 配置创建虚拟机
- **WHEN** 客户端请求体包含 `ipConfig0: "ip=dhcp"`
- **THEN** 系统将 `ipconfig0=ip=dhcp` 传递给 PVE，并附加 CloudInit 驱动盘

#### Scenario: 携带 DNS 配置创建虚拟机
- **WHEN** 客户端请求体包含 `nameserver` 或 `searchDomain` 字段
- **THEN** 系统将对应值作为 `nameserver`、`searchdomain` 参数传递给 PVE，并附加 CloudInit 驱动盘

#### Scenario: 未携带任何 CloudInit 字段
- **WHEN** 请求体不包含任何 CloudInit 相关字段（ciUser、ciPassword、sshKeys、ipConfig0、nameserver、searchDomain）
- **THEN** 系统不向 PVE 传递 `ide2` 或任何 CloudInit 参数，行为与原有磁盘导入创建一致
