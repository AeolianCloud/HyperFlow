## Context

Hyperflow 通过 PVE REST API 管理虚拟机。当前使用 PVE clone API（`POST /nodes/{node}/qemu/{templateid}/clone`）从模板创建虚拟机，该接口会完整复制模板的配置和磁盘，无法在创建时自定义 VM 配置参数。

PVE 提供另一种方式：直接调用 `POST /nodes/{node}/qemu`（创建 VM API），在请求体的磁盘参数中指定 `import-from=<volume>` 来导入已有磁盘。这样可以在创建时完全自定义 CPU、内存、磁盘接口类型等所有配置。

## Goals / Non-Goals

**Goals:**
- 用 `CreateVm`（调用 `POST /nodes/{node}/qemu`）替换 `CloneVm`
- 支持在创建时通过 `import-from` 导入指定磁盘卷
- 在创建时直接指定 CPU 核数、内存、磁盘接口类型
- API 端点路径不变（`POST /nodes/{node}/vms`）

**Non-Goals:**
- 多磁盘导入
- 轮询或等待 PVE 任务完成
- 跨节点操作

## Decisions

### 1. 使用 `POST /nodes/{node}/qemu` + `import-from`

PVE 创建 VM 的接口支持在磁盘参数中内联指定 `import-from=<storage>:<volume>`，一次请求完成 VM 创建和磁盘导入，无需两步操作。

**备选方案**：先创建空 VM，再调用 `POST /nodes/{node}/qemu/{vmid}/move_disk` 导入磁盘 —— 多一次 API 调用，且需处理中间状态错误。

### 2. 请求体字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `vmid` | int | 是 | 新虚拟机 VMID |
| `name` | string | 是 | 新虚拟机名称 |
| `cores` | int | 是 | CPU 核数 |
| `memory` | int | 是 | 内存大小（MB）|
| `diskSource` | string | 是 | 导入磁盘来源（格式：`storage:volname`）|
| `diskInterface` | string | 否 | 磁盘接口类型（`virtio0`/`scsi0`/`ide0`），默认 `virtio0` |
| `storage` | string | 是 | 目标存储（导入后存放的存储池）|

### 3. 移除 CloneVm，不保留兼容层

`CloneVm` 与 `CreateVm` 语义不同，不存在兼容映射关系。直接移除，调用方需迁移到新接口。这是有意的 breaking change，在 proposal 中已标注。

### 4. PVE API 请求体构造

PVE 创建 VM 接口使用 form-encoded 或 JSON 请求体。磁盘参数格式为 `<interface>: <storage>:0,import-from=<source>`，例如：
```
virtio0: local-lvm:0,import-from=local:vm-100-disk-0
```

## Risks / Trade-offs

- [风险] `import-from` 操作为异步，VM 创建后磁盘导入在后台进行，API 立即返回任务 ID → 调用方需通过任务 ID 轮询状态（当前不在范围内）
- [风险] `diskSource` 格式错误时 PVE 返回错误 → 由现有 `handlePveError` 统一处理
- [Breaking] 现有调用 `POST /nodes/{node}/vms` 的客户端需更新请求体字段
