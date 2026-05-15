## Context

PVE 提供 VM 创建能力，CloudInit 支持通过 `ipconfig0` + `nameserver` 配置静态 IP，但 HyperFlow 缺乏 IP 地址统一管理机制。当前 `CreateVmRequest.ipConfig0` 为原始字符串，用户需自行管理 IP 分配，易产生冲突。

本设计在 HyperFlow 内部新增 IP 池子系统，接管 VM 的 IP 生命周期。

## Goals / Non-Goals

**Goals:**
- 管理员可创建 IP 池，指定地址段、网关、掩码、DNS，绑定 PVE 节点
- VM 创建时从 IP 池分配地址（指定或随机），自动构造 CloudInit 参数
- 已分配 IP 不可重复使用
- VM 删除时自动释放 IP
- 两阶段分配：先标记 reserved，PVE 确认成功后标记 used，失败则回退 available

**Non-Goals:**
- 不管理 VM 系统内部的 IP 配置（CloudInit 配置后，OS 内部行为不干预）
- 不支持 IPv6
- 不支持 DHCP 模式
- 不提供 IP 冲突检测
- 不提供 IP 预留功能
- 不涉及 PVE 现有 IP 地址的导入

## Decisions

### 1. 两阶段 IP 分配 (reserved → used)

**选择 B: 同步收拢的事务标记。**

流程：
```
available → (开始创建 VM) → reserved → (PVE 成功) → used
                                      → (PVE 失败) → available
```

分配时使用 `SELECT ... FOR UPDATE` 行级锁防并发。创建 VM 的 PVE 调用失败时立即回滚 IP 状态。

**风险：** 在 "标记 reserved" 和 "创建 Operation 记录" 之间若进程崩溃，会产生孤儿 reserved IP。
**兜底：** 应用启动时扫描 status=reserved 但无对应 Operation 的记录，释放回 available。

### 2. Reconciler 整合

在 `Operation` 模型增加字段，不抽象回调或独立 Reconciler：

```go
type Operation struct {
    // ... 现有字段 ...
    VMID         *int    `json:"vmid,omitempty"`
    AllocationID *string `json:"allocationId,omitempty"` // ip_pool_addresses.id
}
```

Reconciler 在 `CompleteOperation` 事务中增加逻辑：
- 若 `AllocationID != nil` 且状态变为 Succeeded → `UPDATE ip_pool_addresses SET status='used'`
- 若 `AllocationID != nil` 且状态变为 Failed → `UPDATE ip_pool_addresses SET status='available', vm_id=NULL`
- VM 删除操作完成时 → `UPDATE ip_pool_addresses SET status='available', vm_id=NULL WHERE vm_id=?`

### 3. 数据模型

```sql
CREATE TABLE IF NOT EXISTS ip_pools (
    id          VARCHAR(32)  NOT NULL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    gateway     VARCHAR(45)  NOT NULL,
    netmask     INT          NOT NULL,
    dns1        VARCHAR(45)  NULL,
    dns2        VARCHAR(45)  NULL,
    description TEXT         NULL,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_name (name)
);

CREATE TABLE IF NOT EXISTS ip_pool_nodes (
    pool_id VARCHAR(32)  NOT NULL,
    node    VARCHAR(128) NOT NULL,
    PRIMARY KEY (pool_id, node)
);

CREATE TABLE IF NOT EXISTS ip_pool_addresses (
    id         VARCHAR(32)  NOT NULL PRIMARY KEY,
    pool_id    VARCHAR(32)  NOT NULL,
    address    VARCHAR(45)  NOT NULL,
    status     VARCHAR(16)  NOT NULL DEFAULT 'available',
    vm_id      INT          NULL,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_address (address),
    INDEX idx_pool_status (pool_id, status)
);
```

### 4. API 设计

```
POST   /api/pve/ip-pools                      → 创建 IP 池
  Body: { name, gateway, netmask, dns1?, dns2?, description?, nodes: [...], addresses: ["10.0.0.1-10.0.0.50"] }

GET    /api/pve/ip-pools                      → 列表
  Response: [{ id, name, gateway, netmask, dns1, dns2, total, available, used, nodes: [...] }]

GET    /api/pve/ip-pools/:id                  → 详情

PUT    /api/pve/ip-pools/:id                  → 更新（仅 name, dns1, dns2, description, nodes）

DELETE /api/pve/ip-pools/:id                  → 删除（有 used/reserved IP 时返回 409 Conflict）

POST   /api/pve/ip-pools/:id/addresses        → 追加地址（最大 254 个）
  Body: { addresses: ["10.0.1.1-10.0.1.50"] }

DELETE /api/pve/ip-pools/:id/addresses        → 删除地址（仅 available 状态可删除）
  Body: { addresses: ["10.0.1.10", "10.0.1.20"] }

GET    /api/pve/ip-pools/:id/addresses        → 地址列表
  Query: ?status=available&page=1&size=50
```

### 5. CreateVmRequest 扩展

```go
type CreateVmRequest struct {
    // ... 现有字段 ...

    IPPoolID    string `json:"ipPoolId"`               // 选池
    IPAddress   string `json:"ipAddress"`              // 指定 IP（可选）
    AutoAssign  *bool  `json:"autoAssignIp"`           // 随机分配（默认 true 如果给了 ipPoolId）
}
```

校验逻辑：
- 若 `ipPoolId` 非空 → 校验节点是否绑定该池
- 若 `ipPoolId` + `ipAddress` 有值 → 校验该地址是否属于此池且 available
- 若 `ipPoolId` + 无 `ipAddress` + `autoAssignIp` 为 true（或空）→ 随机分配
- 若 `ipPoolId` + 无 `ipAddress` + `autoAssignIp` 为 false → 报错

## Risks / Trade-offs

| 风险 | 缓解 |
|---|---|
| 分配 IP 后 PVE 创建失败，IP 卡在 reserved | 启动时孤儿清理 + 同步路径中的立即回滚 |
| 两处同时分配同一 IP | `SELECT ... FOR UPDATE` 行级锁 |
| 单次导入 254 个地址，插入性能 | 单条 INSERT 带多行 values，MySQL 级毫秒完成 |
| IP 地址全局唯一，不同池不能有相同 IP | 设计如此，不视为风险 |
| Operation 表新增字段影响现有查询 | 字段带默认值，向后兼容 |
