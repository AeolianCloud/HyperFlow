## Why

当前 HyperFlow 创建 VM 时只支持一块系统盘（通过 import-from 导入镜像），不支持额外数据盘。创建完成后也无法对已有 VM 挂载、卸载或删除数据盘。用户在需要额外存储时只能手动登录 PVE Web 界面操作，无法通过 API 完成。

## What Changes

- 创建 VM 时支持可选挂载多块数据盘，每块独立指定大小和存储位置
- 新增 POST /disks 端点，给已有 VM 挂载数据盘，指定大小和存储位置
- 新增 DELETE /disks/{diskId} 端点，卸载/删除 VM 的数据盘
- 新增 GET /disks 端点，列出 VM 的所有磁盘
- 所有涉及 PVE 异步任务的操作用 LRO Operation 模式（202 Accepted）
- 数据盘统一使用 scsi 总线接口，支持热插拔

### 不包含

- 磁盘扩容（resize）
- 磁盘迁移（move_disk）
- CloudInit 自动格式化/挂载数据盘

## Capabilities

### New Capabilities
- `vm-data-disks`: 已有 VM 的数据盘生命周期管理（查询、挂载、卸载、删除）

### Modified Capabilities
- `vm-create-with-disk-import`: 在创建 VM 的请求体中新增 `dataDisks` 可选字段，支持创建时附带数据盘

## Impact

- **cmd/handlers.go**: 新增 4 个路由和 handler（create VM 扩展、POST/DELETE/GET disks）
- **internal/pve/vms.go**: 新增 CreateVmRequest.DataDisks 字段、GetVmConfig/UpdateVmConfig/UnlinkDisk 方法
- **internal/operations/store.go**: 可能新增操作类型或可复用现有结构
- **openspec/specs/**: 新增 `vm-data-disks/spec.md`，修改 `vm-create-with-disk-import/spec.md`
