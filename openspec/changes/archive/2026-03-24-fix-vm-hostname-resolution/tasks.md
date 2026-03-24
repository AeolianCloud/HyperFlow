## 1. 修改 buildCloudInitUserData 函数

- [x] 1.1 为 `buildCloudInitUserData` 新增 `name string` 参数
- [x] 1.2 在生成的 YAML 中 `#cloud-config` 之后注入 `hostname`、`fqdn`、`preserve_hostname: false` 字段

## 2. 更新调用处

- [x] 2.1 更新 `CreateVm` 中对 `buildCloudInitUserData` 的调用，传入 `req.Name`
