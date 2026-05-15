## 1. PVE 客户端层

- [x] 1.1 实现 `GetVmConfig(ctx, node, vmid string) (json.RawMessage, error)` — 调用 `GET /nodes/{node}/qemu/{vmid}/config`
- [x] 1.2 实现 `UpdateVmConfig(ctx, node, vmid string, config map[string]any) (json.RawMessage, error)` — 调用 `POST /nodes/{node}/qemu/{vmid}/config`
- [x] 1.3 实现 `UnlinkDisk(ctx, node, vmid, diskId string, purge bool) (json.RawMessage, error)` — 调用 `POST /nodes/{node}/qemu/{vmid}/unlink`

## 2. 数据结构

- [x] 2.1 在 `CreateVmRequest` 中新增 `DataDisks []DataDiskSpec` 字段，其中 `DataDiskSpec` 包含 `Size int` 和 `Storage string`（均必填）
- [x] 2.2 创建 `VmDisk` 响应结构体：`DiskId`, `Size`, `Storage`, `Format`, `Interface`
- [x] 2.3 创建 `AttachDiskRequest` 结构体：`Size int` + `Storage string`（均必填）

## 3. 创建 VM 扩展数据盘支持

- [x] 3.1 在 `CreateVm` 方法中解析 `req.DataDisks`，根据系统盘接口名自动计算数据盘 scsi 接口索引
- [x] 3.2 在 PVE 请求 body 中添加 `scsi{N}: {storage}:{size}` 参数
- [x] 3.3 更新创建 VM 的请求校验：`dataDisks[]` 中每项 `size` 和 `storage` 均必填

## 4. 已有 VM 的磁盘操作 Handler

- [x] 4.1 实现 `listVmDisks` handler — `GET /api/pve/nodes/{node}/vms/{vmid}/disks`，读取 VM config 解析并返回磁盘列表
- [x] 4.2 实现 `attachVmDisk` handler — `POST /api/pve/nodes/{node}/vms/{vmid}/disks`：
  - MySQL `GET_LOCK("disk_ops:{node}/{vmid}", 5)` 获取锁
  - 调用 `GetVmConfig` 读取当前磁盘列表
  - 自动计算下一个可用 scsi 接口索引
  - 调用 `UpdateVmConfig` 添加磁盘
  - `RELEASE_LOCK`
  - 创建 Operation 记录（LRO），返回 202
- [x] 4.3 实现 `detachVmDisk` handler — `DELETE /api/pve/nodes/{node}/vms/{vmid}/disks/{diskId}[?purge=true]`：
  - MySQL `GET_LOCK("disk_ops:{node}/{vmid}", 5)` 获取锁
  - 调用 `UnlinkDisk`
  - `RELEASE_LOCK`
  - 创建 Operation 记录，返回 202
- [x] 4.4 新增 `POST /disks` 的 `storage` 字段必填校验
- [x] 4.5 新增 `DELETE /disks` 的 `purge=true` 查询参数处理

## 5. 路由注册

- [x] 5.1 在 `registerVmsRoutes` 中注册 3 个新路由：
  - `GET /:vmid/disks` → `listVmDisks`
  - `POST /:vmid/disks` → `attachVmDisk`
  - `DELETE /:vmid/disks/:diskId` → `detachVmDisk`

## 6. 并发安全

- [x] 6.1 实现 MySQL 命名锁辅助函数（`acquireDiskLock` / `releaseDiskLock`），基于 `GET_LOCK` / `RELEASE_LOCK`
- [x] 6.2 锁超时时返回 409 Conflict 标准错误响应
- [x] 6.3 确保锁在 handler 返回前或 panic 时释放（defer）

## 7. 规范同步

- [x] 7.1 创建正式规范 `openspec/specs/vm-data-disks/spec.md`（从 change specs 同步）
- [x] 7.2 更新 `openspec/specs/vm-create-with-disk-import/spec.md`，新增 dataDisks 需求

## 8. 测试

- [x] 8.1 `TestNextSCSIIndex` / `TestParseSCSIDisks` — PVE 工具函数单元测试（8 subtests）
- [x] 8.2 `TestAttachDiskRequestValidation` / `TestCreateVmDataDisksValidation` — handler 层测试（6 subtests）
- [x] 8.3 创建 VM 附带数据盘的请求校验测试（已覆盖）
- [x] 8.4 `TestConcurrentLockSameName` / `TestReleaseDiskLockNoPanic` 等 — 并发安全场景测试（4 subtests）

## 9. Swagger 注释

- [x] 9.1 为新增端点添加 Swagger 注解（请求/响应结构、状态码）
- [x] 9.2 更新 `CreateVmRequest` 的 Swagger 示例，包含 `dataDisks`
- [x] 9.3 运行 `swag init -g cmd/main.go` 重新生成文档
