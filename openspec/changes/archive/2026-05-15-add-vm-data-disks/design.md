## Context

HyperFlow 通过 `POST /api/pve/nodes/{node}/qemu` 创建 VM，只支持单块系统盘（import-from 导入镜像），不支持额外数据盘。PVE 本身支持在创建时一次性设置多个磁盘参数，也支持通过 `POST /nodes/{node}/qemu/{vmid}/config` 对运行中的 VM 热插拔磁盘。

项目已有完整的 LRO Operation 模式（202 + Operation-Location + Reconciler），可复用。

## Goals / Non-Goals

**Goals:**
- 创建 VM 时支持可选数据盘，每块独立指定大小和存储位置
- 给已有 VM 挂载数据盘（热插拔），指定大小和存储位置
- 从 VM 卸载数据盘（仅从配置移除，保留存储卷）
- 从 VM 卸载并销毁数据盘（移除配置 + 删除存储卷）
- 查询 VM 当前所有磁盘列表
- 所有 PVE 异步操作走 LRO 模式（202 Accepted）
- 并发安全：同一 VM 的加盘/拆盘操作互斥

**Non-Goals:**
- 磁盘扩容（resize）
- 磁盘迁移（move_disk）
- CloudInit 自动格式化/挂载数据盘
- 磁盘性能指标查询

## Decisions

### 1. 数据盘接口总线：固定 scsi

数据盘统一使用 `scsi` 总线，接口名格式为 `scsi0`、`scsi1`、`scsi2`……

**理由**：
- 项目已硬编码 `scsihw: "virtio-scsi-single"`，scsi 控制器已就绪
- scsi 支持对运行中 VM 热插拔 ✅
- 系统盘可能用 `virtio0`，数据盘用 `scsiN` 在 PVE 中是正常混搭，无实际约束

### 2. 接口索引自动分配：填充空洞

创建时将数据盘依次分配 `scsi1`、`scsi2`……（假设系统盘为 `scsi0` 或 `virtio0`）。
对已有 VM 加盘时：先读当前 VM config，扫描所有 `scsiN` 键，找出最小可用的索引。

**分配算法**：
```
收集所有匹配 /^scsi(\d+)$/ 的 key → 提取索引号 → 从 0 开始找最小缺失索引
例: scsi0, scsi2, scsi3 → 返回 scsi1
例: scsi0, scsi1       → 返回 scsi2
```
填充空洞而非 max+1，避免因用户手动删盘导致接口号无限增长。

### 3. 并发安全：MySQL 命名锁

加盘/拆盘时，用 MySQL `GET_LOCK() / RELEASE_LOCK()` 以 `disk_ops:{node}/{vmid}` 为锁名进行互斥。

```
锁名称: "disk_ops:pve-node-01/100"
获取超时: 5 秒（超时则返回 409 Conflict）
持有期: 不超过 10 秒
```

**理由**：
- 项目已有 MySQL 依赖，零新基础设施
- 分布式友好（多实例部署时也安全）
- 死锁自动释放（连接断开时 MySQL 自动释放锁）

### 4. PVE 异步任务统一走 LRO

三种场景涉及 PVE 异步任务：
- `POST /nodes/{node}/qemu/{vmid}/config`（加盘至运行中 VM → 返回 UPID）
- `POST /nodes/{node}/qemu/{vmid}/unlink`（拆盘 → 返回 UPID）
- VM 已停止时 PVE 同步完成，无 UPID

统一策略：尝试获取 UPID，若有则创建 Operation（Running）；若无则直接标记 Operation 为 Succeeded。Handler 统一返回 202 + Operation-Location header。

### 5. diskId 格式：PVE 接口名

磁盘标识直接使用 PVE 原生接口名（如 `scsi1`），作为 URL path 参数：
```
DELETE /api/pve/nodes/{node}/vms/{vmid}/disks/scsi1
```

GET /disks 返回的 diskId 直接可用于 DELETE。零映射层。

### 6. storage 必填

每个数据盘必须显式指定 `storage`，无默认值。

- `POST /vms` 的 `dataDisks[].storage` → 必填
- `POST /vms/{vmid}/disks` 的 `storage` → 必填

创建时不继承顶层的 `storage` 字段，避免隐式行为导致磁盘落在非预期的存储池上。

### 7. 大小单位：GB

API 层接收整数 GB，直接透传 PVE（PVE 以 GB 为单位）。

## Risks / Trade-offs

| 风险 | 缓解措施 |
|---|---|
| MySQL 命名锁增加数据库连接耗时 | 锁持有时间极短（<10ms 实际 PVE 调用），超时 5 秒足够 |
| PVE config 端点在高并发下可能覆盖而非报错 | MySQL 命名锁确保同一 VM 的磁盘操作串行执行 |
| 自动分配的接口索引在跨节点场景下不一致 | 分配算法基于实时读取的 VM config，每次加盘都重新计算 |
| VM 停止时 PVE 同步返回（无 UPID），LRO 模式可能短暂延迟 | Operation 读时推进机制（GetOperation 内也调用 reconcile）确保用户及时获知终态 |
