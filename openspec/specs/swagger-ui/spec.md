### Requirement: Swagger UI 可访问
服务 SHALL 在 `/swagger/index.html` 提供可交互的 Swagger UI 页面，展示所有已注册 API 的文档。

#### Scenario: 访问 Swagger UI
- **WHEN** 用户在浏览器中访问 `http://<host>:8080/swagger/index.html`
- **THEN** 返回 HTTP 200，页面显示 Swagger UI 界面并列出所有接口

### Requirement: 接口文档完整性
服务 SHALL 为以下所有路由生成文档：`GET /api/pve/nodes`、`GET /api/pve/nodes/:node`、`GET /api/pve/nodes/:node/vms`、`GET /api/pve/nodes/:node/vms/:vmid`、`POST /api/pve/nodes/:node/vms/:vmid/start`、`POST /api/pve/nodes/:node/vms/:vmid/stop`、`DELETE /api/pve/nodes/:node/vms/:vmid`、`GET /api/pve/storage`。

#### Scenario: 节点列表接口出现在文档中
- **WHEN** 用户查看 Swagger UI
- **THEN** 能看到 `GET /api/pve/nodes` 接口及其响应结构说明

#### Scenario: VM 操作接口出现在文档中
- **WHEN** 用户查看 Swagger UI
- **THEN** 能看到 start/stop/delete VM 的接口，并显示路径参数 `node` 和 `vmid` 的说明

### Requirement: 错误响应有文档
每个接口 SHALL 在文档中描述可能的错误响应，至少包含 404 和 500。

#### Scenario: 404 响应有描述
- **WHEN** 用户在 Swagger UI 查看任意资源获取接口
- **THEN** 文档中显示 404 响应及说明

### Requirement: 全局 API 元信息
文档 SHALL 包含 API 标题、版本号和 base URL。

#### Scenario: 文档显示标题和版本
- **WHEN** 用户打开 Swagger UI
- **THEN** 页面顶部显示项目名称（Hyperflow API）和版本（v1.0）
