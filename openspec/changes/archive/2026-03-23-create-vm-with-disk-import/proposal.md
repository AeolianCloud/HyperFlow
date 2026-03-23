## Why

当前通过 PVE clone API 创建虚拟机，会完整复制模板的所有配置和磁盘，灵活性受限且依赖 PVE 模板克隆机制。改为「新建 VM + 导入磁盘」的方式，可以在创建时完全自定义 VM 配置，并从任意已有磁盘卷导入磁盘，适应更多使用场景。

## What Changes

- 移除 `CloneVm` 方法及 `CloneVmRequest` 结构体
- 新增 `CreateVmRequest` 结构体，包含 VM 配置和磁盘导入来源
- 新增 `CreateVm` 方法，调用 PVE `POST /nodes/{node}/qemu`，通过 `import-from` 参数在创建时导入磁盘
- 将 `cmd/handlers.go` 中的 `cloneVm` handler 替换为 `createVm` handler
- API 端点路径不变（`POST /nodes/{node}/vms`），请求体字段变更

## Capabilities

### New Capabilities
- `vm-create-with-disk-import`: 新建虚拟机并在创建时通过 `import-from` 导入指定磁盘卷

### Modified Capabilities

## Impact

- `internal/pve/vms.go`：移除 `CloneVm`/`CloneVmRequest`，新增 `CreateVm`/`CreateVmRequest`
- `cmd/handlers.go`：`cloneVm` handler 替换为 `createVm`，请求体校验字段变更
- API 请求体字段变更（breaking change for existing clients）：去掉 `templateid`，新增 `diskSource`、`diskInterface`、`storage`（改为必填）
- `docs/`：需重新生成 swagger 文档
