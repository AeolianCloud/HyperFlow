## Purpose

定义通过磁盘导入创建虚拟机时的 CloudInit 扩展行为。

## Requirements

### Requirement: 创建虚拟机时支持 CloudInit 配置
系统 SHALL 在创建虚拟机请求体中接受可选的 CloudInit 配置参数；当请求体包含至少一个 CloudInit 字段时，系统 SHALL 向 PVE 请求附加 CloudInit 驱动盘及对应配置。新增 `ciUpgrade`、`ciPackages`、`snippetsStorage` 可选字段：当 `ciUpgrade=true` 且 `ciPackages` 为空时，传递 `ciupgrade=1`；当 `ciPackages` 非空时，生成 cloud-init user-data Snippet 文件并通过 `cicustom` 引用。

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
- **WHEN** 请求体不包含任何 CloudInit 相关字段（ciUser、ciPassword、sshKeys、ipConfig0、nameserver、searchDomain、ciUpgrade、ciPackages）
- **THEN** 系统不向 PVE 传递 `ide2` 或任何 CloudInit 参数，行为与原有磁盘导入创建一致

#### Scenario: 携带 ciUpgrade=true 且无 ciPackages
- **WHEN** 请求体包含 `ciUpgrade: true`，不包含 `ciPackages`
- **THEN** 系统向 PVE 传递 `ciupgrade=1` 及 CloudInit 驱动盘，不生成 Snippets 文件

#### Scenario: 携带 ciPackages 安装软件包
- **WHEN** 请求体包含 `ciPackages: ["qemu-guest-agent"]` 及 `snippetsStorage: "local"`
- **THEN** 系统生成 user-data Snippet 文件并通过 `cicustom` 参数引用，附加 CloudInit 驱动盘，返回 202 Accepted
