## 1. 依赖与配置

- [x] 1.1 在 `go.mod` 中添加 `github.com/go-sql-driver/mysql`，移除未使用的 `go.mongodb.org/mongo-driver/v2`
- [x] 1.2 在 `.env.example` 中新增 `MYSQL_DSN` 配置项

## 2. Operations 持久化层

- [x] 2.1 创建 `internal/operations/store.go`：定义 `Operation` 结构体、`Store` 接口及 MySQL 实现（`Insert`、`GetByID`、`UpdateStatus`、`CreateTable`）
- [x] 2.2 创建 `internal/operations/service.go`：定义 `Service`，实现 `CreateOperation`（生成随机 ID、写 DB）和 `GetOperation`（懒查询 PVE 任务状态、更新 DB）

## 3. PVE 任务状态查询

- [x] 3.1 在 `internal/pve/vms.go` 中新增 `GetTaskStatus(node, upid string)` 方法，查询 PVE `/nodes/{node}/tasks/{upid}/status`

## 4. 响应格式规范化

- [x] 4.1 修改 `cmd/handlers.go` 中 `respondOK`，去掉 `{"data": ...}` 包装，直接输出资源；移除 `respondAccepted`（异步操作改为无 body 202）
- [x] 4.2 更新所有 `@Success` Swagger 注释，去掉 `map[string]any` 包装

## 5. 异步操作改为 LRO

- [x] 5.1 修改 `cmd/main.go`，初始化 MySQL 连接和 `operations.Service`，传入 handler
- [x] 5.2 重构 `startVm` handler：调用 PVE → 创建 operation → 返回 202 + `Operation-Location` header，无 body
- [x] 5.3 重构 `stopVm` handler：同上
- [x] 5.4 重构 `deleteVm` handler：同上
- [x] 5.5 重构 `createVm` handler：同上，额外保留 `Location` header

## 6. 新增 Operations 端点

- [x] 6.1 在 `cmd/handlers.go` 中新增 `getOperation` handler 及 Swagger 注释
- [x] 6.2 在 `cmd/main.go` 中注册 `GET /operations/:id` 路由
