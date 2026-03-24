## Context

Hyperflow 通过 PVE REST API 管理虚拟机。现有 `Client` 仅支持无请求体的 POST（用于 start/stop），`VmsService` 只覆盖查询、启停、删除操作。PVE 的模板克隆接口 `POST /nodes/{node}/qemu/{vmid}/clone` 需要传递 JSON 请求体（newid、name、storage 等参数）。

## Goals / Non-Goals

**Goals:**
- 扩展 `Client` 支持带请求体的 POST 调用
- 在 `VmsService` 新增 `CloneVm` 方法，封装 PVE clone API
- 新增 HTTP handler `cloneVm`，注册为 `POST /api/pve/nodes/{node}/vms`
- 遵循现有 202 Accepted 的异步响应模式

**Non-Goals:**
- 轮询或等待 PVE 任务完成
- 跨节点克隆
- 完整克隆参数以外的高级配置

## Decisions

### 1. 新增 `PostWithBody` 而非修改现有 `Post`

现有 `Post` 无请求体，用于 start/stop，语义清晰。引入 `PostWithBody(path string, body io.Reader)` 保持向后兼容，不破坏现有调用。

**备选方案**：将 `Post` 改为可选 body，会污染现有调用处，增加 nil 判断噪音。

### 2. 克隆接口路由为 `POST /nodes/{node}/vms`（不带 vmid）

创建是对 vms 集合的操作，符合 RESTful 设计规范。源模板 vmid 放在请求体中，因为路径中的 vmid 语义是操作目标对象，而非克隆来源。

### 3. 请求体字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `templateid` | int | 是 | 源模板 VMID |
| `newid` | int | 是 | 新虚拟机 VMID |
| `name` | string | 是 | 新虚拟机名称 |
| `storage` | string | 否 | 目标存储，不填使用模板默认存储 |
| `cores` | int | 否 | CPU 核数，不填保持模板默认值 |
| `memory` | int | 否 | 内存大小（MB），不填保持模板默认值 |

### 4. 克隆后配置更新

PVE 的 clone API 不支持在克隆时直接指定 CPU/内存。若请求体中包含 `cores` 或 `memory`，需在 clone 请求返回后，额外调用 `PUT /nodes/{node}/qemu/{newid}/config` 更新配置。

**注意**：clone 操作为异步，PVE 返回任务 ID 时虚拟机配置文件已创建，此时可立即调用 config 接口（无需等待任务完成）。

### 5. 新增 `PutWithBody` 方法

与 `PostWithBody` 对称，用于发送带请求体的 PUT 请求（config 更新）。

## Risks / Trade-offs

- [风险] PVE 克隆为异步操作，API 返回任务 ID 后虚拟机尚未就绪 → 调用方需自行轮询任务状态（当前不在范围内）
- [风险] 新 VMID 若已存在，PVE 返回错误 → 由现有 `handlePveError` 统一处理，返回对应 HTTP 状态码
- [风险] clone 返回后立即调用 config 接口，极少数情况下 PVE 配置文件可能尚未落盘 → 当前不做重试，由调用方感知错误后重试
