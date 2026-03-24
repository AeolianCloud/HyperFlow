## ADDED Requirements

### Requirement: API 错误响应格式
所有 API 端点在返回错误时，响应体 SHALL 遵循微软 REST API Guidelines 的标准错误结构：顶层包含 `error` 对象，`error` 对象包含 `code`（PascalCase 机器可读错误码）和 `message`（人类可读说明）字段。

#### Scenario: 客户端请求资源不存在
- **WHEN** 客户端请求一个不存在的资源
- **THEN** 系统 SHALL 返回 404 状态码，响应体格式为 `{"error": {"code": "NotFound", "message": "..."}}`

#### Scenario: 客户端请求参数无效
- **WHEN** 客户端发送缺少必填字段或格式错误的请求体
- **THEN** 系统 SHALL 返回 400 状态码，响应体格式为 `{"error": {"code": "BadRequest", "message": "..."}}`

#### Scenario: 资源状态冲突
- **WHEN** 客户端操作与当前资源状态冲突（如启动已运行的虚拟机）
- **THEN** 系统 SHALL 返回 409 状态码，响应体格式为 `{"error": {"code": "Conflict", "message": "..."}}`

#### Scenario: 服务器内部错误
- **WHEN** 系统发生未预期的内部错误
- **THEN** 系统 SHALL 返回 500 状态码，响应体格式为 `{"error": {"code": "InternalServerError", "message": "..."}}`

#### Scenario: 上游 PVE 服务不可达
- **WHEN** PVE 服务器无法连接
- **THEN** 系统 SHALL 返回 502 状态码，响应体格式为 `{"error": {"code": "BadGateway", "message": "..."}}`
