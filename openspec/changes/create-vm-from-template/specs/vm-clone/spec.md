## ADDED Requirements

### Requirement: 通过模板克隆创建虚拟机
系统 SHALL 提供 `POST /api/pve/nodes/{node}/vms` 接口，接受 JSON 请求体，调用 PVE clone API 从指定模板克隆创建新虚拟机，克隆后若指定了 CPU 核数或内存大小则调用 PVE config 接口更新，并返回 202 Accepted 及 PVE 任务 ID。

请求体字段：
- `templateid`（int，必填）：源模板 VMID
- `newid`（int，必填）：新虚拟机 VMID
- `name`（string，必填）：新虚拟机名称
- `storage`（string，选填）：目标存储名称
- `cores`（int，选填）：CPU 核数，不指定则保持模板默认值
- `memory`（int，选填）：内存大小（MB），不指定则保持模板默认值

#### Scenario: 成功克隆虚拟机（不指定 CPU/内存）
- **WHEN** 客户端发送合法请求体到 `POST /api/pve/nodes/{node}/vms`，未包含 `cores` 和 `memory`
- **THEN** 系统调用 PVE `POST /nodes/{node}/qemu/{templateid}/clone`，不调用 config 接口，返回 HTTP 202 及 PVE 任务 ID

#### Scenario: 成功克隆虚拟机并更新 CPU 和内存
- **WHEN** 客户端请求体中包含 `cores` 和/或 `memory`
- **THEN** 系统先调用 PVE clone 接口，再调用 `PUT /nodes/{node}/qemu/{newid}/config` 更新 CPU/内存，返回 HTTP 202 及 clone 任务 ID

#### Scenario: 请求体缺少必填字段
- **WHEN** 客户端发送缺少 `templateid`、`newid` 或 `name` 的请求体
- **THEN** 系统返回 HTTP 400 Bad Request 及错误描述

#### Scenario: 目标 VMID 已存在
- **WHEN** 客户端指定的 `newid` 在 PVE 中已被占用
- **THEN** 系统返回 PVE 返回的错误状态码（如 500/409）及错误信息

#### Scenario: 源模板不存在
- **WHEN** 客户端指定的 `templateid` 在指定节点上不存在
- **THEN** 系统返回 HTTP 404 及错误信息

### Requirement: Client 支持带请求体的 POST 和 PUT 调用
`pve.Client` SHALL 提供 `PostWithBody(path string, body io.Reader)` 和 `PutWithBody(path string, body io.Reader)` 方法，发送带 JSON 请求体的请求到 PVE API。

#### Scenario: 发送带请求体的 POST 请求
- **WHEN** 调用 `PostWithBody` 并传入有效路径和 JSON body
- **THEN** 请求携带 `Content-Type: application/json` 头及正确的认证信息发送到 PVE

#### Scenario: 发送带请求体的 PUT 请求
- **WHEN** 调用 `PutWithBody` 并传入有效路径和 JSON body
- **THEN** 请求携带 `Content-Type: application/json` 头及正确的认证信息发送到 PVE
