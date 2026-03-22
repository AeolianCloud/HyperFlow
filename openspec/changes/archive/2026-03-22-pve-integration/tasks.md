## 1. 环境配置与依赖

- [x] 1.1 确认项目 Web 框架（Express/Fastify/Koa）及现有目录结构
- [x] 1.2 添加 `axios` 依赖（如未安装）
- [x] 1.3 在环境变量配置文件中增加 `PVE_HOST`、`PVE_TOKEN_ID`、`PVE_TOKEN_SECRET`、`PVE_INSECURE` 说明

## 2. PVE 客户端模块（pve-client）

- [x] 2.1 创建 PveClient 类，读取环境变量并在缺失时抛出配置错误
- [x] 2.2 实现带 `Authorization: PVEAPIToken` 认证头的 axios 实例，支持 `PVE_INSECURE` 跳过 SSL 校验
- [x] 2.3 实现统一错误处理：将 PVE 非 2xx 响应转换为结构化错误对象（含状态码与错误描述）
- [x] 2.4 封装通用 `get`、`post`、`delete` 方法

## 3. 节点管理（pve-nodes）

- [x] 3.1 实现 NodesService：`listNodes()` 调用 `GET /nodes`
- [x] 3.2 实现 NodesService：`getNode(node)` 调用 `GET /nodes/:node/status`
- [x] 3.3 实现 NodesController：注册路由 `GET /api/pve/nodes` 与 `GET /api/pve/nodes/:node`

## 4. 虚拟机管理（pve-vms）

- [x] 4.1 实现 VmsService：`listVms(node)` 调用 `GET /nodes/:node/qemu`
- [x] 4.2 实现 VmsService：`getVm(node, vmid)` 调用 `GET /nodes/:node/qemu/:vmid/status/current`
- [x] 4.3 实现 VmsService：`startVm(node, vmid)` 调用 `POST /nodes/:node/qemu/:vmid/status/start`
- [x] 4.4 实现 VmsService：`stopVm(node, vmid)` 调用 `POST /nodes/:node/qemu/:vmid/status/stop`
- [x] 4.5 实现 VmsService：`deleteVm(node, vmid)` 调用 `DELETE /nodes/:node/qemu/:vmid`
- [x] 4.6 实现 VmsController：注册所有虚拟机路由，运行中虚拟机执行删除/冲突操作时返回 409

## 5. 存储管理（pve-storage）

- [x] 5.1 实现 StorageService：`listStorage()` 调用 `GET /storage`
- [x] 5.2 实现 StorageController：注册路由 `GET /api/pve/storage`

## 6. 路由注册

- [x] 6.1 在应用主路由入口注册所有 `/api/pve/*` 路由
- [x] 6.2 确认 PveClient 在应用启动时完成初始化校验（缺失配置时快速失败）
