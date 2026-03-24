## Why

当前 API 响应格式不符合 Microsoft REST API Guidelines：成功响应统一用 `{"data": ...}` 包装，异步操作直接暴露 PVE 内部 UPID，客户端必须感知 PVE 底层细节。本次改动完整对齐 Guidelines，屏蔽 PVE 实现细节，引入标准 LRO（Long-Running Operation）模式。

## What Changes

- **BREAKING** 去掉所有成功响应的 `{"data": ...}` 包装层，直接返回资源对象或数组
- **BREAKING** 异步操作（create/start/stop/delete VM）不再返回 body，改为返回 `Operation-Location` header 指向操作状态端点
- **新增** `GET /api/pve/operations/{id}` 端点，返回标准 LRO 状态（Running / Succeeded / Failed）
- **新增** MySQL 持久化 operations 表，重启后操作记录不丢失
- **新增** `.env` 配置项 `MYSQL_DSN`
- **移除** 未使用的 MongoDB driver 依赖

## Capabilities

### New Capabilities
- `lro-operations`: 长时间运行操作（LRO）管理，含 MySQL 持久化、PVE 任务状态懒查询、标准 LRO 响应格式

### Modified Capabilities
- `pve-nodes`: GET 响应去掉 `data` 包装，直接返回资源
- `pve-vms`: GET 响应去掉 `data` 包装；异步操作响应改为 `Operation-Location` header
- `pve-storage`: GET 响应去掉 `data` 包装

## Impact

- 影响文件：`cmd/handlers.go`、`cmd/main.go`、`go.mod`、`.env.example`
- 新增文件：`internal/operations/store.go`、`internal/operations/service.go`
- 所有 API 客户端需同步更新（breaking change）
- 需要外部 MySQL 实例
