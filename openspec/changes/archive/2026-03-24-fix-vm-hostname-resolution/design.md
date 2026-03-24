## Context

新建虚拟机时，若请求包含 `ciPackages`，系统会调用 `buildCloudInitUserData` 生成自定义 cloud-init user-data 并通过 `cicustom` 参数传递给 PVE。当前该函数的输出中没有 `hostname` 字段，导致虚拟机首次开机后 cloud-init 不会设置主机名，hostname 保持为镜像默认值（通常是镜像名称或随机字符串）。

无 `ciPackages` 时，PVE 使用原生 CloudInit 参数，PVE 会自动将虚拟机 `name` 字段写为 hostname，不受此问题影响。

## Goals / Non-Goals

**Goals:**
- 使用 `ciPackages` 创建的虚拟机，首次开机后 hostname 与虚拟机 `name` 一致
- 不引入新的 API 字段或接口变更

**Non-Goals:**
- 修改无 `ciPackages` 时的原生 CloudInit 路径（该路径已正常工作）
- 支持运行时动态修改 hostname

## Decisions

### 决策：在 user-data YAML 中注入 `hostname` 字段

`buildCloudInitUserData` 新增 `name string` 参数，在生成的 YAML 顶部（`#cloud-config` 之后）写入：

```yaml
hostname: <name>
fqdn: <name>
preserve_hostname: false
```

- `hostname`：cloud-init 设置的短主机名
- `fqdn`：防止 cloud-init 用 DHCP 获取的 FQDN 覆盖 hostname
- `preserve_hostname: false`：确保 cloud-init 有权限修改 hostname（部分镜像默认为 true）

调用处 `CreateVm` 中将 `req.Name` 作为新参数传入。

**备选方案：新增 `CIHostname` 字段** — 增加 API 复杂度，而 VM name 即是预期 hostname，无需额外字段。

## Risks / Trade-offs

- [Risk] VM name 含有不合法的 hostname 字符（如下划线） → hostname 写入后可能不符合 RFC 1123。缓解：PVE 本身对 VM name 有校验，实际创建的 name 通常已合法；本次不额外校验，可在后续迭代中加入。
- [Risk] 镜像 `preserve_hostname` 默认为 true → 通过显式设置 `preserve_hostname: false` 覆盖。
