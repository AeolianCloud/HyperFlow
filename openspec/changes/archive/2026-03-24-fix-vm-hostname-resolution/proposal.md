## Why

新建虚拟机时，若使用 `ciPackages` 触发了自定义 `cicustom` user-data，生成的 cloud-init YAML 中未包含 `hostname` 指令，导致虚拟机首次开机后无法正确识别主机名（hostname 仍为镜像默认值）。

## What Changes

- 在 `buildCloudInitUserData` 函数生成的 YAML 中加入 `hostname` 和 `fqdn` 字段，确保 cloud-init 首次启动时写入正确的主机名
- `CreateVmRequest` 无需新增字段，直接复用已有的 `Name` 字段作为 hostname 来源
- 将 VM `Name` 作为参数传入 `buildCloudInitUserData`

## Capabilities

### New Capabilities
<!-- 无新能力，仅修复现有行为 -->

### Modified Capabilities
- `pve-vms`: `buildCloudInitUserData` 函数输出的 cloud-init user-data 需包含 `hostname` 字段，保证主机名与虚拟机名称一致

## Impact

- 影响文件：`internal/pve/vms.go`
- 影响函数：`buildCloudInitUserData`（新增 `name` 参数）、`CreateVm`（更新调用处）
- 无 API 接口变更，无 breaking change
