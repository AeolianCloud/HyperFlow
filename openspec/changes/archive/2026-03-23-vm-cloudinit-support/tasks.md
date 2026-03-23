## 1. 扩展 CreateVmRequest 结构体

- [x] 1.1 在 `internal/pve/vms.go` 的 `CreateVmRequest` 中新增 CloudInit 可选字段：`CIUser`、`CIPassword`、`SSHKeys`、`IPConfig0`、`Nameserver`、`SearchDomain`，并补全 Swagger example 注释

## 2. 更新 CreateVm 业务逻辑

- [x] 2.1 在 `CreateVm` 方法中，检测是否存在任意 CloudInit 字段，若存在则向 PVE 请求体追加 `ide2: <storage>:cloudinit`
- [x] 2.2 按需追加 `ciuser`、`cipassword`、`nameserver`、`searchdomain`、`ipconfig0` 参数
- [x] 2.3 对 `SSHKeys` 字段使用 `url.QueryEscape` 编码后作为 `sshkeys` 参数传递

## 3. 更新 Swagger 文档

- [x] 3.1 在 `cmd/handlers.go` 的 `createVm` godoc 注释中更新请求体描述，说明 CloudInit 字段为可选
- [x] 3.2 运行 `swag init` 重新生成 `docs/` 下的 Swagger 文档
