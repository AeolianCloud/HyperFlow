## MODIFIED Requirements

### Requirement: 新建虚拟机（含 ciPackages）
当请求包含 `ciPackages` 时，系统 SHALL 生成包含 `hostname`、`fqdn` 和 `preserve_hostname: false` 字段的 cloud-init user-data YAML，确保虚拟机首次开机后主机名与虚拟机 `name` 一致。

#### Scenario: ciPackages 非空时 user-data 包含 hostname
- **WHEN** 客户端发送 `POST /api/pve/nodes/:node/vms`，且 `ciPackages` 非空
- **THEN** 上传的 cloud-init user-data YAML SHALL 包含 `hostname: <name>`、`fqdn: <name>` 和 `preserve_hostname: false`

#### Scenario: ciPackages 为空时行为不变
- **WHEN** 客户端发送 `POST /api/pve/nodes/:node/vms`，且 `ciPackages` 为空
- **THEN** 系统 SHALL 使用 PVE 原生 CloudInit 参数，不生成自定义 user-data，行为与修改前一致
