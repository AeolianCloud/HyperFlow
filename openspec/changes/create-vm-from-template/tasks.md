## 1. 扩展 PVE Client

- [ ] 1.1 在 `internal/pve/client.go` 中新增 `PostWithBody(path string, body io.Reader) (json.RawMessage, error)` 方法
- [ ] 1.2 在 `internal/pve/client.go` 中新增 `PutWithBody(path string, body io.Reader) (json.RawMessage, error)` 方法

## 2. 扩展 VmsService

- [ ] 2.1 在 `internal/pve/vms.go` 中新增 `CloneVmRequest` 结构体（字段：TemplateID、NewID、Name、Storage、Cores、Memory）
- [ ] 2.2 在 `internal/pve/vms.go` 中新增 `CloneVm(node string, req CloneVmRequest) (json.RawMessage, error)` 方法：调用 PVE `POST /nodes/{node}/qemu/{templateid}/clone`，若 Cores 或 Memory 不为零则再调用 `PUT /nodes/{node}/qemu/{newid}/config` 更新配置

## 3. 新增 HTTP Handler

- [ ] 3.1 在 `cmd/handlers.go` 中新增 `cloneVm` handler，解析请求体、校验必填字段（templateid、newid、name），调用 `vmsSvcGlobal.CloneVm`，返回 202 Accepted
- [ ] 3.2 在 `registerVmsRoutes` 中注册 `POST ""` 路由指向 `cloneVm`

## 4. 更新 API 文档注释

- [ ] 4.1 为 `cloneVm` handler 添加 swaggo 注释（Summary、Tags、Param、Body、Success、Failure、Router）
- [ ] 4.2 运行 `swag init -g cmd/main.go -o docs` 重新生成 swagger 文档
