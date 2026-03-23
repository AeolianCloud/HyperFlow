## Context

Hyperflow 通过 PVE API 创建虚拟机时已支持 CloudInit 基础配置（用户名、密码、SSH 公钥、网络等）。但 qemu-guest-agent 是 PVE 管理虚拟机的核心组件（用于在线快照、IP 上报、QEMU 命令通道），云镜像默认不预装，需首次开机自动安装。此外，首次开机更新软件包也是云镜像最佳实践之一。

PVE 提供两种机制满足此需求：
1. `ciupgrade=1`：PVE 8.1+ 原生支持，通知 cloud-init 在首次启动时执行 `package_upgrade`
2. `cicustom=user:snippets/xxx.yaml`：通过 Snippets 存储提供完整的 cloud-init user-data 文件，支持 `packages` 列表（可指定安装 qemu-guest-agent 等）

## Goals / Non-Goals

**Goals:**
- 支持 `ciUpgrade` 字段，控制首次开机是否更新全部软件包
- 支持 `ciPackages` 字段，指定首次开机安装的软件包列表（如 `["qemu-guest-agent"]`）
- 当指定 `ciPackages` 时，自动生成 cloud-init user-data Snippet 并通过 `cicustom` 引用
- 保持向后兼容，未传新字段时行为不变

**Non-Goals:**
- 不支持完整 cloud-init user-data 自定义（仅 packages/upgrade）
- 不管理 Snippets 存储的生命周期（不自动删除已生成文件）
- 不支持 PVE 8.1 以下版本的 ciupgrade（该版本无此参数，使用 cicustom 代替）

## Decisions

### 决策 1：`ciUpgrade` 使用 PVE 原生 `ciupgrade` 参数

**方案 A（选用）**：直接向 PVE 传递 `ciupgrade=1`（PVE 8.1+ 支持）。
**方案 B**：在 user-data Snippet 中写 `package_upgrade: true`。
**理由**：方案 A 更简洁，无需生成文件；若同时有 `ciPackages`，则两者可共用同一 Snippet，在 Snippet 中用 `package_upgrade: true` + `packages:` 列表，不再额外传 `ciupgrade`。

### 决策 2：`ciPackages` 通过 Snippets 存储的 user-data 文件实现

PVE 没有原生的「安装指定包」参数，必须通过 `cicustom` 指向一个 cloud-init 格式的 YAML 文件（存放在 Snippets 类型存储中）。
文件内容格式：
```yaml
#cloud-config
package_update: true
package_upgrade: <ciUpgrade>
packages:
  - qemu-guest-agent
```

**文件命名**：`cloudinit-<vmid>-userdata.yaml`，存放在请求方指定的 Snippets 存储中（新增 `snippetsStorage` 字段）。

### 决策 3：仅当 `ciPackages` 非空时生成 Snippet

若只设置 `ciUpgrade=true` 而无 `ciPackages`，直接用 `ciupgrade=1` 参数，无需生成文件，减少外部存储依赖。
若 `ciPackages` 非空，则生成 Snippet 并在其中同时处理 `package_upgrade`（忽略 `ciupgrade` 参数），两者合并为一个文件。

## Risks / Trade-offs

- **Snippets 存储依赖** → 若未配置 Snippets 类型存储，`ciPackages` 功能不可用。缓解：在请求参数中要求用户明确指定 `snippetsStorage`，并在 PVE 返回错误时透传给调用方。
- **文件残留** → 虚拟机删除后 Snippets 文件不会自动清理，占用少量存储。缓解：文件体积极小（< 1KB），可记录在文档中由运维手动清理，或在后续版本中实现删除钩子。
- **PVE 版本兼容** → `ciupgrade` 参数仅 PVE 8.1+ 支持。缓解：若同时有 `ciPackages`，在 Snippet 中用 `package_upgrade` 字段替代，无需 `ciupgrade` 参数。
