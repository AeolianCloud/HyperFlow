## Why

Hyperflow 目前仅支持对已有虚拟机的查询、启动、停止和删除操作，缺少从模板克隆创建新虚拟机的能力，无法满足快速批量部署虚拟机的需求。

## What Changes

- 新增 `POST /api/pve/nodes/{node}/vms` 接口，通过克隆 PVE 模板创建虚拟机
- 请求体包含：源模板 VMID、新虚拟机 VMID、名称、目标存储（可选）、CPU 核数（可选）、内存大小（可选，单位 MB）
- 克隆完成后若指定了 CPU/内存，自动调用 PVE config 接口更新配置
- 接口异步执行，返回 202 Accepted 及 PVE 任务 ID
- 在 `VmsService` 中新增 `CloneVm` 方法，调用 PVE API `POST /nodes/{node}/qemu/{vmid}/clone`
- `Client` 增加支持带请求体的 `PostWithBody` 和 `PutWithBody` 方法

## Capabilities

### New Capabilities

- `vm-clone`: 通过克隆 PVE 模板创建新虚拟机的 API 能力，包含请求参数定义、PVE clone 调用、克隆后 CPU/内存配置更新及响应格式

### Modified Capabilities

## Impact

- `internal/pve/client.go`：新增 `PostWithBody`、`PutWithBody` 方法
- `internal/pve/vms.go`：新增 `CloneVm` 方法及请求结构体（含 CPU/内存字段）
- `cmd/handlers.go`：新增 `cloneVm` handler 及路由注册
- `cmd/main.go`：路由注册无需修改（路由已在 `registerVmsRoutes` 中管理）
- 依赖 PVE API：`POST /nodes/{node}/qemu/{vmid}/clone`、`PUT /nodes/{node}/qemu/{newid}/config`
