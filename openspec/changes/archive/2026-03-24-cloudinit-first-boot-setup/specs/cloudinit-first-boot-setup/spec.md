## ADDED Requirements

### Requirement: 创建虚拟机时支持首次开机软件包更新
系统 SHALL 在 `CreateVmRequest` 中接受可选的 `ciUpgrade` 布尔字段；当 `ciUpgrade` 为 `true` 且未指定 `ciPackages` 时，系统 SHALL 向 PVE 传递 `ciupgrade=1` 参数，触发首次开机执行完整软件包更新。

#### Scenario: 仅设置 ciUpgrade=true 创建虚拟机
- **WHEN** 客户端发送 `POST /api/pve/nodes/:node/vms`，请求体包含必填字段及 `ciUpgrade: true`，不包含 `ciPackages`
- **THEN** 系统调用 PVE `POST /nodes/{node}/qemu` 时，参数中包含 `ciupgrade=1` 及 CloudInit 驱动盘 `ide2`，不生成 Snippets 文件，返回 202 Accepted

#### Scenario: ciUpgrade=false 时不传 ciupgrade 参数
- **WHEN** 请求体中 `ciUpgrade` 为 `false` 或未设置
- **THEN** 系统不向 PVE 传递 `ciupgrade` 参数

### Requirement: 创建虚拟机时支持首次开机安装指定软件包
系统 SHALL 在 `CreateVmRequest` 中接受可选的 `ciPackages` 字符串列表字段；当 `ciPackages` 非空时，系统 SHALL 在指定的 Snippets 存储中生成 cloud-init user-data YAML 文件，并通过 `cicustom` 参数引用该文件，文件中包含 `packages` 列表及（若 `ciUpgrade=true`）`package_upgrade: true`。

#### Scenario: 指定 ciPackages 安装 qemu-guest-agent
- **WHEN** 客户端请求体包含 `ciPackages: ["qemu-guest-agent"]` 及 `snippetsStorage: "local"`
- **THEN** 系统在 `local` Snippets 存储中生成文件 `cloudinit-<vmid>-userdata.yaml`，内容包含 `packages: [qemu-guest-agent]`，并向 PVE 传递 `cicustom: user=local:snippets/cloudinit-<vmid>-userdata.yaml`，返回 202 Accepted

#### Scenario: ciPackages 同时设置 ciUpgrade=true
- **WHEN** 请求体包含 `ciPackages: ["qemu-guest-agent"]` 且 `ciUpgrade: true`
- **THEN** 生成的 user-data 文件中包含 `package_upgrade: true` 和 `packages: [qemu-guest-agent]`，系统不额外传递 `ciupgrade` 参数（已在文件中处理）

#### Scenario: ciPackages 多个软件包
- **WHEN** 请求体包含 `ciPackages: ["qemu-guest-agent", "curl"]`
- **THEN** 生成的 user-data 文件中 `packages` 列表包含全部指定软件包

#### Scenario: 未指定 snippetsStorage 但包含 ciPackages
- **WHEN** 请求体包含 `ciPackages` 但未提供 `snippetsStorage`
- **THEN** 系统返回 400 Bad Request，说明 `snippetsStorage` 为必填项

#### Scenario: PVE Snippets 存储写入失败
- **WHEN** 系统向 PVE 上传 user-data 文件时 PVE 返回错误
- **THEN** 系统返回对应 HTTP 错误码及错误信息，不创建虚拟机

#### Scenario: 未设置 ciPackages 时不生成 Snippets 文件
- **WHEN** 请求体不包含 `ciPackages` 字段
- **THEN** 系统不向 PVE 上传任何 user-data 文件，不传递 `cicustom` 参数
