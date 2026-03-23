## 1. 修改 VmsService

- [x] 1.1 删除 `internal/pve/vms.go` 中的 `CloneVmRequest` 结构体和 `CloneVm` 方法
- [x] 1.2 新增 `CreateVmRequest` 结构体（字段：VMID、Name、Cores、Memory、DiskSource、DiskInterface、Storage）
- [x] 1.3 新增 `CreateVm(node string, req CreateVmRequest) (json.RawMessage, error)` 方法，构造磁盘参数（`<interface>: <storage>:0,import-from=<diskSource>`）并调用 PVE `POST /nodes/{node}/qemu`

## 2. 修改 HTTP Handler

- [x] 2.1 将 `cmd/handlers.go` 中的 `cloneVm` handler 替换为 `createVm` handler，解析新请求体并校验必填字段（vmid、name、cores、memory、diskSource、storage）
- [x] 2.2 在 `registerVmsRoutes` 中将 `POST ""` 路由指向 `createVm`

## 3. 更新 API 文档注释

- [x] 3.1 为 `createVm` handler 添加 swaggo 注释（Summary、Tags、Param、Body、Success、Failure、Router），移除旧的 `cloneVm` 注释
- [x] 3.2 运行 `swag init -g cmd/main.go -o docs` 重新生成 swagger 文档
