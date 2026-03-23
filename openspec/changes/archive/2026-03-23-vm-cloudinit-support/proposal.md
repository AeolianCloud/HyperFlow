## Why

当前新建虚拟机接口仅支持磁盘导入，无法在创建时配置 CloudInit 参数（如用户名、密码、SSH 公钥、网络配置等）。使用云镜像（cloud image）创建虚拟机时，必须通过 CloudInit 完成首次启动配置，否则用户需在虚拟机启动后手动进入控制台操作，体验差且无法自动化。

## What Changes

- 在创建虚拟机请求体中新增可选的 CloudInit 配置字段（用户名、密码、SSH 公钥、IP 配置、DNS 等）
- `CreateVm` 业务逻辑在调用 PVE `POST /nodes/{node}/qemu` 时，自动添加 CloudInit 驱动盘（`ide2: <storage>:cloudinit`）及对应配置参数
- 更新 Swagger 文档及 `CreateVmRequest` 结构体注释

## Capabilities

### New Capabilities

（无新增独立 capability，属于对现有 VM 创建能力的扩展）

### Modified Capabilities

- `vm-create-with-disk-import`: 在现有磁盘导入创建虚拟机的基础上，增加 CloudInit 配置参数支持，变更创建时传递给 PVE 的参数集合

## Impact

- `internal/pve/vms.go`：`CreateVmRequest` 新增 CloudInit 相关字段；`CreateVm` 方法构造 PVE 请求参数时加入 CloudInit 驱动盘与配置
- `cmd/handlers.go`：`createVm` handler 无需改动（逻辑封装在 service 层）
- `docs/`：需重新生成 Swagger 文档（`swag init`）
- 依赖：PVE API 原生支持 CloudInit，无需额外依赖
