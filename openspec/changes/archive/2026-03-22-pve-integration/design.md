## Context

Hyperflow 是一个基于 OpenSpec 规范驱动的开发平台。当前平台尚无虚拟化基础设施管理能力。PVE（Proxmox Virtual Environment）是业界广泛使用的开源虚拟化平台，提供完整的 REST API，支持对 QEMU/KVM 虚拟机、LXC 容器、节点、存储进行管理。

本次集成通过封装 PVE REST API，以统一的 HTTP API 层对外暴露虚拟化资源管理能力，遵循 Hyperflow 既有的 RESTful API 设计规范。

## Goals / Non-Goals

**Goals:**
- 封装 PVE REST API 客户端，支持 API Token 认证
- 提供节点、虚拟机（QEMU/KVM）、存储池的查询与生命周期管理 API
- 遵循 Hyperflow RESTful API 设计规范（统一响应格式、错误处理、分页）
- 通过环境变量配置 PVE 连接信息，不硬编码敏感信息

**Non-Goals:**
- 不实现 PVE 集群高可用、备份、快照等高级功能（后续迭代）
- 不实现虚拟机模板管理
- 不实现 WebSocket 控制台代理
- 不支持多 PVE 集群同时管理

## Decisions

### 1. 分层架构：客户端 → Service → Controller

采用三层结构：

```
Controller (HTTP路由/请求解析/响应格式化)
    └── Service (业务逻辑/参数校验)
            └── PveClient (HTTP封装/认证/重试)
```

**理由**：与项目现有 API 架构保持一致，便于测试和维护。Controller 只处理 HTTP 层，Service 处理业务逻辑，PveClient 专注与 PVE 通信。

### 2. 认证方式：API Token（而非用户名/密码 Ticket）

PVE 提供两种认证：用户名/密码换取 Ticket（有时效，需刷新），以及 API Token（长期有效，权限可细粒度控制）。

**选择 API Token**：无需处理 Token 刷新逻辑，适合服务端集成场景，安全性更可控。

格式：`Authorization: PVEAPIToken=<tokenid>=<secret>`

### 3. HTTP 客户端：axios

项目若已有 axios 依赖则复用；若无则引入。PVE 默认使用自签名 SSL 证书，开发环境允许通过配置跳过证书校验（`rejectUnauthorized: false`），生产环境应配置正式证书。

### 4. API 路由前缀：`/api/pve`

所有 PVE 相关接口统一以 `/api/pve` 为前缀，子资源路径：

```
GET  /api/pve/nodes                          # 节点列表
GET  /api/pve/nodes/:node                    # 节点详情
GET  /api/pve/nodes/:node/vms                # 虚拟机列表
GET  /api/pve/nodes/:node/vms/:vmid          # 虚拟机详情
POST /api/pve/nodes/:node/vms/:vmid/start    # 启动虚拟机
POST /api/pve/nodes/:node/vms/:vmid/stop     # 停止虚拟机
DELETE /api/pve/nodes/:node/vms/:vmid        # 删除虚拟机
GET  /api/pve/storage                        # 存储池列表
```

### 5. 错误处理：统一错误格式

PVE API 返回的错误（如 4xx/5xx）统一转换为 Hyperflow 标准错误响应格式，避免将 PVE 内部错误直接透传给客户端。

## Risks / Trade-offs

- **PVE 自签名证书** → 开发环境通过配置 `PVE_INSECURE=true` 跳过验证；生产环境要求配置受信任证书
- **PVE API 版本兼容性** → 当前针对 PVE 7.x/8.x；如遇 API 变更需更新客户端封装
- **网络可达性** → PVE 服务器需与 Hyperflow 服务网络互通，防火墙需开放 8006 端口
- **异步任务** → PVE 部分操作（如删除、创建）返回 task UPID 而非立即完成；当前设计仅返回 UPID，不实现任务轮询（后续迭代）

## Migration Plan

1. 添加环境变量配置：`PVE_HOST`、`PVE_TOKEN_ID`、`PVE_TOKEN_SECRET`、`PVE_INSECURE`（可选）
2. 实现 PveClient 模块
3. 实现各 Service 模块
4. 注册路由
5. 本地联调验证（需可访问的 PVE 测试环境）

**回滚**：路由注册为独立模块，回滚只需移除路由注册，不影响现有功能。

## Open Questions

- 项目当前使用的 Web 框架是什么（Express / Fastify / Koa）？需确认后选择对应路由注册方式
- 是否需要对 PVE 操作做权限控制（与 Hyperflow 用户系统集成）？
- PVE 异步任务（UPID）是否需要在本期实现状态查询接口？
