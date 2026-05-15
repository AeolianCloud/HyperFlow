## 1. 数据层

- [x] 1.1 在 `internal/operations/store.go` 的 `CreateTable` 中加入 `ip_pools`、`ip_pool_nodes`、`ip_pool_addresses` 三张表的建表语句
- [x] 1.2 在 `Operation` 结构体中增加 `VMID` 和 `AllocationID` 字段，对应 MySQL 新增列
- [x] 1.3 `Store` 接口增加 IP 池相关方法：`CreatePool`, `GetPool`, `ListPools`, `UpdatePool`, `DeletePool`
- [x] 1.4 `Store` 接口增加地址相关方法：`InsertAddresses`, `DeleteAddresses`, `ListAddresses`, `AllocateAddress`, `ReleaseAddress`
- [x] 1.5 实现 IP 段解析工具函数：解析 `"10.0.0.1-10.0.0.50"` 格式，展开为 IP 列表，校验最大 254 个

## 2. IP 池管理 API

- [x] 2.1 创建 `internal/ippool/service.go`：IP 池业务逻辑（CRUD 校验、节点绑定校验、地址冲突校验）
- [x] 2.2 在 `cmd/handlers.go` 中添加 IP 池 CRUD 路由和处理器：`POST/GET /api/pve/ip-pools`、`GET/PUT/DELETE /api/pve/ip-pools/:id`
- [x] 2.3 在 `cmd/handlers.go` 中添加地址管理路由：`POST/DELETE /api/pve/ip-pools/:id/addresses`、`GET /api/pve/ip-pools/:id/addresses`
- [x] 2.4 注册路由到 `cmd/main.go` 的 `routerGroup` 中

## 3. IP 分配与 VM 创建集成

- [x] 3.1 在 `internal/pve/vms.go` 的 `CreateVmRequest` 中增加 `IPPoolID`, `IPAddress`, `AutoAssign` 字段
- [x] 3.2 在 VM 创建流程中增加 IP 池逻辑：校验节点绑定 → 分配 IP（SELECT FOR UPDATE）→ 构造 ipConfig0/nameserver
- [x] 3.3 创建 VM 成功后，将 `allocation_id` 写入 Operation
- [x] 3.4 创建 VM 失败时，立即释放 IP 回 available
- [x] 3.5 修改 `cmd/handlers.go` 的 createVm 处理器，传递 IP 池参数

## 4. Reconciler IP 状态推进

- [x] 4.1 修改 `CompleteOperation` 事务逻辑：根据 `AllocationID` 将 reserved→used 或 reserved→available
- [x] 4.2 删除 VM 的 reconcile 完成后，根据 `vm_id` 释放 IP 回 available
- [x] 4.3 应用启动时扫描孤儿 reserved IP 并释放回 available

## 5. 测试

- [x] 5.1 为 `internal/ippool/` 包编写单元测试（fake store）
- [x] 5.2 为修改后的 Reconciler 逻辑补充测试（含 allocation_id 的场景）
- [x] 5.3 为 handler 层编写测试（IP 池 CRUD、VM 创建含 IP 池参数）
- [x] 5.4 IP 段解析工具函数的测试
- [x] 5.5 `go test -v ./...` 全量通过

## 6. 文档

- [x] 6.1 运行 `swag init -g cmd/main.go` 重新生成 Swagger 文档
