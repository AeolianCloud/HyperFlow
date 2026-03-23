## Context

现有 `CreateVm` 接口通过 PVE `POST /nodes/{node}/qemu` 创建虚拟机，支持磁盘导入（`import-from`）。PVE 原生支持 CloudInit：只需在创建时附加一块 CloudInit 驱动盘（`ide2: <storage>:cloudinit`）并传入 `ciuser`、`cipassword`、`sshkeys`、`ipconfig0` 等参数，PVE 即会在首次启动时将配置写入虚拟机。

变更范围仅限于 `internal/pve/vms.go` 中的 `CreateVmRequest` 结构体与 `CreateVm` 方法，以及相应的 Swagger 注释。

## Goals / Non-Goals

**Goals:**
- 在 `CreateVmRequest` 中新增可选 CloudInit 字段
- `CreateVm` 方法在有 CloudInit 配置时，自动向 PVE 请求体追加 CloudInit 驱动盘及配置参数
- 保持与已有磁盘导入行为完全向后兼容（不传 CloudInit 字段时行为不变）
- 更新 Swagger 注释

**Non-Goals:**
- 不支持在虚拟机创建后单独修改 CloudInit 配置（属于独立接口）
- 不验证 SSH 公钥格式（交由 PVE 处理）
- 不支持多网卡 CloudInit 配置（`ipconfig1+`）

## Decisions

### 决策 1：CloudInit 字段全部可选，通过 `omitempty` 控制

CloudInit 配置对使用非云镜像的场景无意义，设为可选字段。只有当请求中包含至少一个 CloudInit 字段时，才向 PVE 附加 CloudInit 驱动盘（`ide2`）。

**备选方案**：新增独立端点 `/vms/:vmid/cloudinit`。
**拒绝原因**：PVE 要求 CloudInit 盘在创建时一并配置，事后添加需要额外步骤，拆分端点会增加调用方复杂度。

### 决策 2：CloudInit 驱动盘固定使用 `ide2`

PVE 官方文档与社区实践均推荐使用 `ide2` 作为 CloudInit 驱动盘接口，与 `virtio0`/`scsi0` 数据盘不冲突。

**备选方案**：允许调用方指定 CloudInit 盘接口。
**拒绝原因**：增加复杂度，无实际需求。

### 决策 3：`ipconfig0` 以字符串形式直传 PVE 格式

PVE 的 `ipconfig0` 语法为 `ip=<CIDR>,gw=<GW>` 或 `ip=dhcp`，直接透传可避免在服务层重复解析，且与 PVE API 文档保持一致。

## Risks / Trade-offs

- [Risk] SSH 公钥含特殊字符（换行、空格）时 URL 编码可能出错 → PVE 要求对 `sshkeys` 进行 URL 编码，需在 `CreateVm` 中使用 `url.QueryEscape`
- [Risk] 若目标存储不支持 CloudInit 盘格式 → PVE 返回错误，由现有 `handlePveError` 透传给调用方
- [Trade-off] `ipconfig0` 字符串格式不做校验，调用方需自行保证格式正确

## Migration Plan

1. 修改 `CreateVmRequest` 结构体，新增 CloudInit 字段
2. 修改 `CreateVm` 方法，检测 CloudInit 字段并构造相应 PVE 参数
3. 运行 `swag init` 重新生成 Swagger 文档
4. 无数据迁移，无破坏性变更，直接部署即可
