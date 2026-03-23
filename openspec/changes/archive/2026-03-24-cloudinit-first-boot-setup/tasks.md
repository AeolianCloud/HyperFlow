## 1. 扩展请求结构体

- [x] 1.1 在 `internal/pve/vms.go` 的 `CreateVmRequest` 中新增 `CIUpgrade bool`、`CIPackages []string`、`SnippetsStorage string` 字段，补全 JSON tag 和 Swagger 注释

## 2. 实现 Snippets 文件上传

- [x] 2.1 在 `internal/pve/` 中实现向 PVE Snippets 存储上传文本文件的方法（调用 PVE `POST /nodes/{node}/storage/{storage}/upload` 或直接写入 `/var/lib/vz/snippets/`），命名为 `UploadSnippet(node, storage, filename, content string) error`
- [x] 2.2 实现生成 cloud-init user-data YAML 内容的函数，接收 `ciUpgrade bool` 和 `packages []string`，返回格式正确的 `#cloud-config` YAML 字符串

## 3. 更新 CreateVm 业务逻辑

- [x] 3.1 在 `CreateVm` 方法中处理 `CIUpgrade`：当 `CIUpgrade=true` 且 `CIPackages` 为空时，向 PVE body 追加 `ciupgrade=1`
- [x] 3.2 在 `CreateVm` 方法中处理 `CIPackages`：当 `CIPackages` 非空时，校验 `SnippetsStorage` 非空（否则返回错误），调用 `UploadSnippet` 上传 user-data 文件，向 PVE body 追加 `cicustom` 参数，不额外传 `ciupgrade`
- [x] 3.3 确保以上逻辑触发 `hasCloudInit` 标志，附加 `ide2: <storage>:cloudinit`

## 4. 更新 Swagger 文档

- [x] 4.1 在 `cmd/handlers.go` 对应路由的 Swagger 注释中同步更新请求体描述，说明 `ciUpgrade`、`ciPackages`、`snippetsStorage` 三个新字段
- [x] 4.2 运行 `swag init` 重新生成 `docs/docs.go`，确认文档正确反映新字段

## 5. 验证

- [x] 5.1 构建项目（`go build ./...`），确认无编译错误
- [ ] 5.2 手动测试：创建虚拟机时携带 `ciPackages: ["qemu-guest-agent"]`，确认 Snippets 文件生成并虚拟机首次开机后 qemu-guest-agent 正常运行
- [ ] 5.3 手动测试：仅设置 `ciUpgrade: true`，确认 PVE 收到 `ciupgrade=1` 且无 Snippets 文件生成
