## Why

当前 CloudInit 配置已支持用户名、密码、SSH 公钥、网络等基础参数，但缺少首次开机自动更新软件包和安装 qemu-guest-agent 的能力。qemu-guest-agent 是 PVE 管理虚拟机的重要组件（支持在线快照、QEMU 命令通道等），而云镜像默认不预装，需要在首次开机时通过 CloudInit 自动完成安装。

## What Changes

- 在 `CreateVmRequest` 中新增 `ciUpgrade` 字段（bool），控制首次开机是否执行软件包更新（对应 PVE 的 `ciupgrade` 参数，PVE 8.1+）
- 在 `CreateVmRequest` 中新增 `ciPackages` 字段（字符串列表），指定首次开机需安装的软件包（如 `qemu-guest-agent`）
- 当 `ciPackages` 非空时，系统通过 PVE Snippets 存储生成 cloud-init `user-data` 文件，并使用 `cicustom` 参数引用该文件
- 更新 Swagger 注释，保证接口文档完整

## Capabilities

### New Capabilities

- `cloudinit-first-boot-setup`: 支持在创建虚拟机时通过 CloudInit 配置首次开机自动更新软件并安装指定软件包（包括 qemu-guest-agent）

### Modified Capabilities

- `vm-create-with-disk-import`: 在现有创建虚拟机请求中扩展 CloudInit 可选参数，新增 `ciUpgrade` 和 `ciPackages` 字段

## Impact

- `internal/pve/vms.go`：`CreateVmRequest` 新增字段，`CreateVm` 方法处理新字段
- `cmd/handlers.go`：无需修改接口路由，仅 Swagger 注释需同步更新
- `docs/docs.go`：需重新生成 swag 文档
- 依赖 PVE 节点上存在可用的 Snippets 类型存储（用于存放 user-data 文件）；`ciupgrade` 需要 PVE 8.1+
