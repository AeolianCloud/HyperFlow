## Why

当前 Hyperflow 平台缺乏对虚拟化基础设施的管理能力。通过接入 PVE（Proxmox Virtual Environment）虚拟化底座，平台可以对虚拟机、容器等计算资源进行统一的生命周期管理，满足云基础设施自动化运维需求。

## What Changes

- 新增 PVE API 客户端模块，封装与 Proxmox VE REST API 的通信
- 新增节点管理 API：查询集群节点状态、资源使用率
- 新增虚拟机管理 API：列出、创建、启动、停止、删除虚拟机（QEMU/KVM）
- 新增存储管理 API：查询存储池状态与容量
- 支持 PVE API Token 认证方式

## Capabilities

### New Capabilities

- `pve-client`: PVE REST API 客户端，负责认证、请求封装与错误处理
- `pve-nodes`: 集群节点查询与状态监控能力
- `pve-vms`: 虚拟机（QEMU/KVM）生命周期管理能力
- `pve-storage`: 存储池查询与管理能力

### Modified Capabilities

（无现有 capability 变更）

## Impact

- **新增依赖**：需要 HTTP 客户端库（如 axios）与 PVE 服务器网络可达
- **API 层**：新增 `/api/pve/nodes`、`/api/pve/vms`、`/api/pve/storage` 路由
- **配置**：需在环境变量或配置文件中提供 PVE 主机地址、API Token ID 及 Secret
- **安全**：API Token 不得明文写入代码，通过环境变量注入
