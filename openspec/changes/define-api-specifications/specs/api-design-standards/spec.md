## ADDED Requirements

### Requirement: 资源命名规范
API资源路径SHALL使用复数名词，采用kebab-case命名风格，清晰表达资源类型。

#### Scenario: 标准资源路径
- **WHEN** 定义用户资源的API端点
- **THEN** 路径应为 `/api/users` 而不是 `/api/user` 或 `/api/User`

#### Scenario: 嵌套资源路径
- **WHEN** 定义用户的订单资源
- **THEN** 路径应为 `/api/users/{userId}/orders` 表达资源层级关系

#### Scenario: 多词资源命名
- **WHEN** 资源名称包含多个单词
- **THEN** 使用kebab-case格式，如 `/api/user-profiles` 而不是 `/api/userProfiles`

### Requirement: HTTP方法使用规范
API SHALL根据操作类型正确使用HTTP方法，遵循RESTful语义。

#### Scenario: 获取资源列表
- **WHEN** 客户端需要获取资源集合
- **THEN** 使用GET方法，如 `GET /api/users`

#### Scenario: 获取单个资源
- **WHEN** 客户端需要获取特定资源
- **THEN** 使用GET方法，如 `GET /api/users/{id}`

#### Scenario: 创建新资源
- **WHEN** 客户端需要创建新资源
- **THEN** 使用POST方法，如 `POST /api/users`，请求体包含资源数据

#### Scenario: 完整更新资源
- **WHEN** 客户端需要完整替换资源
- **THEN** 使用PUT方法，如 `PUT /api/users/{id}`，请求体包含完整资源数据

#### Scenario: 部分更新资源
- **WHEN** 客户端只需更新资源的部分字段
- **THEN** 使用PATCH方法，如 `PATCH /api/users/{id}`，请求体只包含需要更新的字段

#### Scenario: 删除资源
- **WHEN** 客户端需要删除资源
- **THEN** 使用DELETE方法，如 `DELETE /api/users/{id}`

### Requirement: HTTP状态码规范
API响应SHALL使用标准HTTP状态码，准确表达操作结果。

#### Scenario: 成功获取资源
- **WHEN** GET请求成功返回资源
- **THEN** 返回200 OK状态码

#### Scenario: 成功创建资源
- **WHEN** POST请求成功创建资源
- **THEN** 返回201 Created状态码，并在Location头中包含新资源的URI

#### Scenario: 成功更新资源
- **WHEN** PUT或PATCH请求成功更新资源
- **THEN** 返回200 OK状态码，响应体包含更新后的资源

#### Scenario: 成功删除资源
- **WHEN** DELETE请求成功删除资源
- **THEN** 返回204 No Content状态码

#### Scenario: 请求参数错误
- **WHEN** 客户端请求包含无效参数或缺少必需参数
- **THEN** 返回400 Bad Request状态码，响应体包含错误详情

#### Scenario: 未授权访问
- **WHEN** 客户端未提供有效的认证凭证
- **THEN** 返回401 Unauthorized状态码

#### Scenario: 权限不足
- **WHEN** 客户端已认证但无权限访问资源
- **THEN** 返回403 Forbidden状态码

#### Scenario: 资源不存在
- **WHEN** 请求的资源不存在
- **THEN** 返回404 Not Found状态码

#### Scenario: 请求冲突
- **WHEN** 请求与当前资源状态冲突（如并发更新）
- **THEN** 返回409 Conflict状态码

#### Scenario: 服务器内部错误
- **WHEN** 服务器处理请求时发生未预期的错误
- **THEN** 返回500 Internal Server Error状态码

### Requirement: API版本控制
API SHALL支持版本控制，确保向后兼容性和平滑升级。

#### Scenario: URI路径版本控制
- **WHEN** 定义API版本
- **THEN** 在URI路径中包含版本号，如 `/api/v1/users`、`/api/v2/users`

#### Scenario: 主版本号变更
- **WHEN** API发生破坏性变更（breaking changes）
- **THEN** 增加主版本号，如从v1升级到v2

#### Scenario: 版本废弃通知
- **WHEN** 旧版本API计划废弃
- **THEN** 在响应头中添加Deprecation和Sunset头，提前通知客户端

#### Scenario: 多版本并存
- **WHEN** 新版本API发布
- **THEN** 旧版本API应继续维护一段时间，确保客户端有足够时间迁移
