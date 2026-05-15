## Why

PVE 本身不提供 IP 地址管理能力，用户手动管理 IP 分配容易产生冲突和重复。HyperFlow 需要内置 IP 池功能，让管理员可以创建和管理 IP 地址池，在创建 VM 时自动或手动分配 IP，确保地址不重复使用。

## What Changes

- 新增 `ip_pools`、`ip_pool_addresses`、`ip_pool_nodes` 三张 MySQL 表
- 新增 IP 池 CRUD API 端点（创建/列表/详情/更新/删除/追加地址/删除地址/地址列表）
- 修改 VM 创建流程：支持从 IP 池分配地址，自动填充 ipConfig0、nameserver
- 修改 VM 删除流程：Reconciler 完成时自动释放已分配的 IP
- 修改 Operation 模型：增加 `vmid` 和 `allocation_id` 字段
- 新增长度限制：单次最大导入 254 个地址
- 新增强约束：IP 全局唯一、有已用 IP 时不允许删除池、不允许修改 gateway/netmask

## Capabilities

### New Capabilities

- `ip-pool-management`: IP 池的生命周期管理——创建、更新、删除、绑定节点、导入导出地址
- `ip-pool-allocation`: VM 创建时从 IP 池分配地址，支持指定 IP 和随机分配

### Modified Capabilities

- *暂无现有关联的能力规约*

## Impact

- `internal/pve/vms.go`: `CreateVmRequest` 新增 `ipPoolId`/`ipAddress`/`autoAssignIp` 字段；创建流程中增加 IP 分配逻辑
- `internal/operations/store.go`: `Operation` 新增 `vmid` 和 `allocation_id` 字段；新增 `ip_pools` 系列表
- `internal/operations/service.go`: Reconciler 完成操作时根据 `allocation_id` 释放或确认 IP
- `cmd/handlers.go`: 新增 IP 池 CRUD 路由和处理函数；修改 VM 创建/删除的处理逻辑
- 新增 `internal/ippool/` 包：IP 池业务逻辑、数据访问层
