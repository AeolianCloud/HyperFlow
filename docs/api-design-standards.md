# API设计标准

本文档定义RESTful API的设计标准，确保API的一致性、可维护性和易用性。

## 目录

1. [资源命名规范](#资源命名规范)
2. [HTTP方法使用规范](#http方法使用规范)
3. [HTTP状态码规范](#http状态码规范)
4. [API版本控制](#api版本控制)

---

## 资源命名规范

### 基本原则

API资源路径必须遵循以下命名规范：

1. **使用复数名词**：资源名称使用复数形式，表示资源集合
2. **kebab-case命名**：多个单词使用连字符分隔，全部小写
3. **清晰表达资源类型**：路径应直观反映资源的业务含义

### 规范详情

#### 1. 复数名词

资源路径使用复数名词，即使操作单个资源也使用复数形式。

**正确示例：**
```
GET /api/users          # 获取用户列表
GET /api/users/123      # 获取单个用户
POST /api/orders        # 创建订单
```

**错误示例：**
```
GET /api/user           # ✗ 应使用复数
GET /api/User           # ✗ 不应使用大写
```

#### 2. kebab-case命名风格

多词资源名称使用kebab-case（连字符分隔），避免使用camelCase或snake_case。

**正确示例：**
```
GET /api/user-profiles
GET /api/order-items
GET /api/payment-methods
```

**错误示例：**
```
GET /api/userProfiles   # ✗ 不使用camelCase
GET /api/user_profiles  # ✗ 不使用snake_case
GET /api/UserProfiles   # ✗ 不使用PascalCase
```

#### 3. 嵌套资源路径

当资源之间存在明确的层级关系时，使用嵌套路径表达。嵌套层级不应超过3层。

**正确示例：**
```
GET /api/users/123/orders              # 获取用户123的订单列表
GET /api/users/123/orders/456          # 获取用户123的订单456
POST /api/users/123/addresses          # 为用户123创建地址
GET /api/organizations/1/teams/5/members  # 获取组织1的团队5的成员
```

**何时使用嵌套：**
- 子资源强依赖于父资源（如用户的订单）
- 需要在父资源上下文中操作子资源

**何时避免嵌套：**
- 资源可以独立存在和访问
- 嵌套层级超过3层
- 需要跨多个父资源查询子资源

**替代方案（使用查询参数）：**
```
GET /api/orders?userId=123             # 通过查询参数过滤
GET /api/members?organizationId=1&teamId=5
```

#### 4. 避免动词

资源路径应使用名词，通过HTTP方法表达操作，避免在路径中使用动词。

**正确示例：**
```
POST /api/users                        # 创建用户
DELETE /api/users/123                  # 删除用户
PUT /api/users/123/password            # 更新密码（password作为子资源）
```

**错误示例：**
```
POST /api/createUser                   # ✗ 路径中包含动词
GET /api/getUsers                      # ✗ 路径中包含动词
POST /api/users/123/activate           # ✗ 应使用PATCH更新状态
```

**例外情况（非资源操作）：**
某些操作不适合映射为资源，可以使用动词：
```
POST /api/users/123/send-notification  # 发送通知（动作）
POST /api/orders/123/cancel            # 取消订单（状态转换）
POST /api/search                       # 复杂搜索
```

#### 5. 版本前缀

所有API路径应包含版本号前缀（详见版本控制章节）。

**标准格式：**
```
/api/v1/users
/api/v2/orders
```

## HTTP方法使用规范

### 基本原则

API必须根据操作类型正确使用HTTP方法，遵循RESTful语义和幂等性要求。

### HTTP方法详解

#### GET - 获取资源

**用途：** 获取资源或资源集合，不应产生副作用。

**特性：**
- 安全操作（Safe）：不修改服务器状态
- 幂等操作（Idempotent）：多次调用结果相同
- 可缓存（Cacheable）

**使用场景：**

1. **获取资源列表**
```
GET /api/v1/users
响应：200 OK
{
  "data": [...],
  "total": 100,
  "offset": 0,
  "limit": 20
}
```

2. **获取单个资源**
```
GET /api/v1/users/123
响应：200 OK
{
  "id": 123,
  "name": "张三",
  "email": "zhangsan@example.com"
}
```

3. **获取嵌套资源**
```
GET /api/v1/users/123/orders
响应：200 OK
```

**注意事项：**
- GET请求不应包含请求体
- 使用查询参数进行过滤、排序、分页
- 资源不存在时返回404

#### POST - 创建资源

**用途：** 创建新资源或执行非幂等操作。

**特性：**
- 非安全操作：修改服务器状态
- 非幂等操作：多次调用产生多个资源
- 不可缓存

**使用场景：**

1. **创建新资源**
```
POST /api/v1/users
请求体：
{
  "name": "李四",
  "email": "lisi@example.com"
}

响应：201 Created
Location: /api/v1/users/124
{
  "id": 124,
  "name": "李四",
  "email": "lisi@example.com",
  "createdAt": "2024-03-19T10:00:00Z"
}
```

2. **执行操作（非资源创建）**
```
POST /api/v1/orders/123/cancel
响应：200 OK
```

**注意事项：**
- 成功创建返回201，并在Location头中包含新资源URI
- 请求体包含资源数据
- 验证失败返回400

#### PUT - 完整更新资源

**用途：** 完整替换现有资源。

**特性：**
- 非安全操作：修改服务器状态
- 幂等操作：多次调用结果相同
- 不可缓存

**使用场景：**

```
PUT /api/v1/users/123
请求体：
{
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "address": "北京市朝阳区"
}

响应：200 OK
{
  "id": 123,
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "address": "北京市朝阳区",
  "updatedAt": "2024-03-19T10:05:00Z"
}
```

**注意事项：**
- 请求体必须包含资源的完整表示
- 未提供的字段将被清空或设为默认值
- 资源不存在时可选择返回404或创建新资源（返回201）
- 多次执行相同PUT请求，资源状态保持一致

#### PATCH - 部分更新资源

**用途：** 部分更新资源的某些字段。

**特性：**
- 非安全操作：修改服务器状态
- 幂等操作（推荐）：多次调用结果相同
- 不可缓存

**使用场景：**

```
PATCH /api/v1/users/123
请求体：
{
  "phone": "13900139000"
}

响应：200 OK
{
  "id": 123,
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13900139000",
  "address": "北京市朝阳区",
  "updatedAt": "2024-03-19T10:10:00Z"
}
```

**注意事项：**
- 请求体只包含需要更新的字段
- 未提供的字段保持不变
- 适合频繁的小范围更新
- 返回更新后的完整资源

#### DELETE - 删除资源

**用途：** 删除指定资源。

**特性：**
- 非安全操作：修改服务器状态
- 幂等操作：多次删除同一资源结果相同
- 不可缓存

**使用场景：**

```
DELETE /api/v1/users/123

响应：204 No Content
（无响应体）
```

**注意事项：**
- 成功删除返回204 No Content
- 资源不存在时返回404（首次删除后的重复删除也返回404）
- 如需返回被删除的资源信息，可返回200 OK并包含响应体
- 软删除（标记为删除）应使用PATCH更新状态

#### HEAD - 获取资源元数据

**用途：** 获取资源的元数据，不返回响应体。

**使用场景：**
```
HEAD /api/v1/users/123
响应：200 OK
Content-Length: 256
Last-Modified: Wed, 19 Mar 2024 10:00:00 GMT
```

**注意事项：**
- 响应头与GET相同，但无响应体
- 用于检查资源是否存在或获取元信息

#### OPTIONS - 获取支持的方法

**用途：** 查询资源支持的HTTP方法。

**使用场景：**
```
OPTIONS /api/v1/users/123
响应：200 OK
Allow: GET, PUT, PATCH, DELETE
```

### 方法选择决策树

```
需要获取数据？
  └─ 是 → GET

需要创建新资源？
  └─ 是 → POST

需要更新资源？
  ├─ 完整替换 → PUT
  └─ 部分更新 → PATCH

需要删除资源？
  └─ 是 → DELETE

需要执行操作（非CRUD）？
  └─ POST（如发送邮件、取消订单）
```

### 幂等性总结

| 方法 | 幂等性 | 说明 |
|------|--------|------|
| GET | ✓ | 多次调用返回相同结果 |
| POST | ✗ | 多次调用创建多个资源 |
| PUT | ✓ | 多次调用资源状态相同 |
| PATCH | ✓ | 推荐设计为幂等 |
| DELETE | ✓ | 多次删除结果相同（资源不存在） |

## HTTP状态码规范

### 基本原则

API必须使用标准HTTP状态码准确表达操作结果，帮助客户端理解请求处理状态。

### 状态码分类

#### 2xx 成功

表示请求已成功被服务器接收、理解并处理。

##### 200 OK
**用途：** 请求成功，返回请求的数据。

**使用场景：**
- GET请求成功获取资源
- PUT/PATCH请求成功更新资源
- DELETE请求成功删除资源（如需返回被删除资源信息）

```
GET /api/v1/users/123
响应：200 OK
{
  "id": 123,
  "name": "张三"
}
```

##### 201 Created
**用途：** 资源创建成功。

**使用场景：**
- POST请求成功创建新资源

**要求：**
- 必须在Location响应头中包含新资源的URI
- 响应体应包含新创建的资源

```
POST /api/v1/users
响应：201 Created
Location: /api/v1/users/124
{
  "id": 124,
  "name": "李四",
  "createdAt": "2024-03-19T10:00:00Z"
}
```

##### 202 Accepted
**用途：** 请求已接受，但处理尚未完成（异步处理）。

**使用场景：**
- 长时间运行的操作
- 批量处理任务
- 异步任务提交

```
POST /api/v1/reports/generate
响应：202 Accepted
{
  "taskId": "task-123",
  "status": "processing",
  "statusUrl": "/api/v1/tasks/task-123"
}
```

##### 204 No Content
**用途：** 请求成功，但无内容返回。

**使用场景：**
- DELETE请求成功删除资源
- PUT/PATCH请求成功但不需要返回更新后的资源

```
DELETE /api/v1/users/123
响应：204 No Content
（无响应体）
```

#### 3xx 重定向

表示需要客户端采取进一步操作才能完成请求。

##### 301 Moved Permanently
**用途：** 资源已永久移动到新位置。

**使用场景：**
- API端点永久迁移

```
GET /api/v1/old-endpoint
响应：301 Moved Permanently
Location: /api/v2/new-endpoint
```

##### 304 Not Modified
**用途：** 资源未修改，可使用缓存。

**使用场景：**
- 条件GET请求，资源未变化

```
GET /api/v1/users/123
If-None-Match: "etag-value"
响应：304 Not Modified
```

#### 4xx 客户端错误

表示请求包含错误或无法完成。

##### 400 Bad Request
**用途：** 请求参数错误或格式不正确。

**使用场景：**
- 请求体JSON格式错误
- 缺少必需参数
- 参数类型错误
- 参数值不符合验证规则

```
POST /api/v1/users
请求体：{ "email": "invalid-email" }

响应：400 Bad Request
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": [
      {
        "field": "email",
        "message": "邮箱格式不正确"
      }
    ]
  }
}
```

##### 401 Unauthorized
**用途：** 未提供认证凭证或认证凭证无效。

**使用场景：**
- 未提供Authorization头
- Token无效或已过期
- Token签名验证失败

```
GET /api/v1/users/me
响应：401 Unauthorized
WWW-Authenticate: Bearer realm="API"
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "认证失败，请提供有效的访问令牌"
  }
}
```

##### 403 Forbidden
**用途：** 已认证但无权限访问资源。

**使用场景：**
- 用户角色权限不足
- 访问其他用户的私有资源
- 操作被策略禁止

```
DELETE /api/v1/users/123
响应：403 Forbidden
{
  "error": {
    "code": "FORBIDDEN",
    "message": "您没有权限删除此用户"
  }
}
```

##### 404 Not Found
**用途：** 请求的资源不存在。

**使用场景：**
- 资源ID不存在
- 端点路径错误
- 资源已被删除

```
GET /api/v1/users/999
响应：404 Not Found
{
  "error": {
    "code": "NOT_FOUND",
    "message": "用户不存在"
  }
}
```

##### 405 Method Not Allowed
**用途：** HTTP方法不被允许。

**使用场景：**
- 对只读资源使用POST/PUT/DELETE
- 端点不支持该HTTP方法

```
POST /api/v1/users/123
响应：405 Method Not Allowed
Allow: GET, PUT, PATCH, DELETE
{
  "error": {
    "code": "METHOD_NOT_ALLOWED",
    "message": "此端点不支持POST方法"
  }
}
```

##### 409 Conflict
**用途：** 请求与当前资源状态冲突。

**使用场景：**
- 并发更新冲突（乐观锁）
- 唯一性约束冲突（如邮箱已存在）
- 业务状态冲突（如订单已取消无法支付）

```
POST /api/v1/users
请求体：{ "email": "existing@example.com" }

响应：409 Conflict
{
  "error": {
    "code": "CONFLICT",
    "message": "该邮箱已被注册"
  }
}
```

##### 422 Unprocessable Entity
**用途：** 请求格式正确但语义错误。

**使用场景：**
- 业务逻辑验证失败
- 数据关系不满足约束

```
POST /api/v1/orders
请求体：{ "quantity": -5 }

响应：422 Unprocessable Entity
{
  "error": {
    "code": "INVALID_DATA",
    "message": "订单数量必须大于0"
  }
}
```

##### 429 Too Many Requests
**用途：** 请求频率超过限制。

**使用场景：**
- 触发限流策略

**要求：**
- 必须包含Retry-After响应头

```
GET /api/v1/users
响应：429 Too Many Requests
Retry-After: 60
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1710842460
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "请求过于频繁，请稍后再试"
  }
}
```

#### 5xx 服务器错误

表示服务器在处理请求时发生错误。

##### 500 Internal Server Error
**用途：** 服务器内部错误。

**使用场景：**
- 未捕获的异常
- 数据库连接失败
- 第三方服务调用失败

```
GET /api/v1/users
响应：500 Internal Server Error
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "服务器内部错误，请稍后重试"
  }
}
```

**注意：** 不应在响应中暴露敏感的错误堆栈信息。

##### 502 Bad Gateway
**用途：** 网关或代理服务器从上游服务器收到无效响应。

**使用场景：**
- 上游服务不可用
- 上游服务返回无效响应

##### 503 Service Unavailable
**用途：** 服务暂时不可用。

**使用场景：**
- 服务维护中
- 服务过载
- 依赖服务不可用

**要求：**
- 应包含Retry-After响应头

```
GET /api/v1/users
响应：503 Service Unavailable
Retry-After: 300
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "服务维护中，预计5分钟后恢复"
  }
}
```

##### 504 Gateway Timeout
**用途：** 网关或代理服务器等待上游服务器响应超时。

**使用场景：**
- 上游服务响应缓慢
- 请求处理超时

### 状态码选择指南

| 场景 | 状态码 |
|------|--------|
| 成功获取资源 | 200 |
| 成功创建资源 | 201 |
| 成功删除资源（无返回） | 204 |
| 异步任务已接受 | 202 |
| 请求参数错误 | 400 |
| 未认证 | 401 |
| 已认证但无权限 | 403 |
| 资源不存在 | 404 |
| HTTP方法不支持 | 405 |
| 资源冲突 | 409 |
| 业务逻辑验证失败 | 422 |
| 请求频率超限 | 429 |
| 服务器内部错误 | 500 |
| 服务不可用 | 503 |

### 最佳实践

1. **一致性**：相同场景使用相同状态码
2. **准确性**：选择最能描述情况的状态码
3. **错误详情**：4xx和5xx响应应包含错误详情
4. **避免滥用200**：不要所有情况都返回200，在响应体中表示错误

## API版本控制

### 基本原则

API必须支持版本控制，确保向后兼容性和平滑升级，避免破坏现有客户端。

### 版本控制策略

#### URI路径版本控制（推荐）

在URI路径中包含版本号，格式为 `/api/v{major}/`。

**格式：**
```
/api/v1/users
/api/v2/orders
/api/v3/products
```

**优点：**
- 版本号清晰可见，易于理解
- 便于缓存和路由配置
- 客户端可明确选择使用的版本
- 符合RESTful最佳实践

**缺点：**
- URL会随版本变化
- 需要维护多个版本的端点

#### 版本号规则

##### 主版本号（Major Version）

使用整数表示主版本号（v1, v2, v3...），仅在发生破坏性变更时递增。

**破坏性变更（Breaking Changes）包括：**
- 删除端点或资源
- 删除请求/响应字段
- 修改字段类型
- 修改字段语义
- 修改认证机制
- 修改错误响应格式

**示例：**
```
v1: /api/v1/users
     响应：{ "id": 1, "name": "张三", "email": "..." }

v2: /api/v2/users  # 删除了email字段（破坏性变更）
     响应：{ "id": 1, "name": "张三" }
```

##### 非破坏性变更

以下变更不需要增加主版本号：
- 添加新端点
- 添加可选的请求参数
- 添加响应字段
- 添加新的HTTP方法支持
- 修复bug

**示例：**
```
v1: /api/v1/users
     响应：{ "id": 1, "name": "张三" }

v1（更新后）: /api/v1/users  # 添加了phone字段（非破坏性）
     响应：{ "id": 1, "name": "张三", "phone": "138..." }
```

### 版本生命周期管理

#### 版本并存

新版本发布后，旧版本应继续维护一段时间，确保客户端有足够时间迁移。

**推荐策略：**
- 同时维护最多3个主版本
- 每个版本至少维护6个月
- 提前3个月通知版本废弃

**示例：**
```
当前支持的版本：
- v3（最新版本，推荐使用）
- v2（维护中，计划2024年12月废弃）
- v1（维护中，计划2024年9月废弃）
```

#### 版本废弃通知

当计划废弃某个版本时，必须通过响应头提前通知客户端。

**使用Deprecation和Sunset响应头：**

```
GET /api/v1/users
响应：200 OK
Deprecation: true
Sunset: Sat, 31 Dec 2024 23:59:59 GMT
Link: </api/v2/users>; rel="successor-version"
{
  "data": [...]
}
```

**响应头说明：**
- `Deprecation: true` - 表示此版本已废弃
- `Sunset` - 版本停止服务的日期
- `Link` - 指向新版本的链接

#### 版本停止服务

版本到达Sunset日期后，应返回410 Gone状态码。

```
GET /api/v1/users
响应：410 Gone
{
  "error": {
    "code": "VERSION_RETIRED",
    "message": "API v1已停止服务，请使用v2",
    "migrationGuide": "https://docs.example.com/migration/v1-to-v2"
  }
}
```

### 版本迁移指南

#### 为客户端提供迁移支持

每次发布新版本时，应提供：

1. **变更日志（Changelog）**
   - 列出所有破坏性变更
   - 列出新增功能
   - 列出废弃的功能

2. **迁移指南（Migration Guide）**
   - 详细说明如何从旧版本迁移到新版本
   - 提供代码示例
   - 说明不兼容的地方及解决方案

3. **并行运行期**
   - 允许客户端在迁移期间同时使用新旧版本
   - 提供工具帮助客户端测试新版本

#### 版本迁移示例

**场景：** 从v1迁移到v2，用户资源的email字段移到contacts子资源

**v1（旧版本）：**
```
GET /api/v1/users/123
响应：
{
  "id": 123,
  "name": "张三",
  "email": "zhangsan@example.com"
}
```

**v2（新版本）：**
```
GET /api/v2/users/123
响应：
{
  "id": 123,
  "name": "张三",
  "contacts": {
    "email": "zhangsan@example.com"
  }
}
```

**迁移指南：**
```
# v1到v2迁移指南

## 破坏性变更

### 用户资源结构变更
- email字段移至contacts对象
- 旧代码：user.email
- 新代码：user.contacts.email

## 迁移步骤

1. 更新API基础URL：/api/v1 → /api/v2
2. 更新用户对象访问方式：
   ```javascript
   // v1
   const email = user.email;

   // v2
   const email = user.contacts.email;
   ```
```

### 版本检测

客户端应在请求中标识自己使用的版本，便于服务端统计和监控。

**推荐方式：**
```
GET /api/v2/users
User-Agent: MyApp/2.0 (API-v2)
```

### 最佳实践

1. **谨慎引入破坏性变更**：尽量通过添加而非修改来扩展API
2. **提前通知**：至少提前3个月通知版本废弃
3. **文档完善**：每个版本都应有完整的文档
4. **监控使用情况**：跟踪各版本的使用量，了解迁移进度
5. **自动化测试**：确保多版本并存时的正确性
6. **版本号从v1开始**：避免使用v0，给人不稳定的印象

### 不推荐的版本控制方式

#### 请求头版本控制
```
GET /api/users
Accept: application/vnd.example.v2+json
```
**缺点：** 对客户端不够直观，调试困难

#### 查询参数版本控制
```
GET /api/users?version=2
```
**缺点：** 不符合RESTful语义，容易被忽略或遗漏

#### 子域名版本控制
```
https://v2.api.example.com/users
```
**缺点：** 需要额外的DNS和证书配置，增加运维复杂度

---

## 设计标准示例

### 完整示例：用户管理API

以下是一个符合所有设计标准的完整API示例。

#### 1. 获取用户列表

**请求：**
```http
GET /api/v1/users?offset=0&limit=20&status=active&sort=-createdAt HTTP/1.1
Host: api.example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应：**
```http
HTTP/1.1 200 OK
Content-Type: application/json
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1710842460

{
  "data": [
    {
      "id": 123,
      "name": "张三",
      "email": "zhangsan@example.com",
      "status": "active",
      "createdAt": "2024-03-19T10:00:00Z"
    },
    {
      "id": 124,
      "name": "李四",
      "email": "lisi@example.com",
      "status": "active",
      "createdAt": "2024-03-18T15:30:00Z"
    }
  ],
  "total": 156,
  "offset": 0,
  "limit": 20
}
```

**符合的规范：**
- ✓ 使用复数名词（users）
- ✓ 使用GET方法获取资源列表
- ✓ 返回200状态码
- ✓ 包含版本号（v1）
- ✓ 支持分页（offset/limit）
- ✓ 支持过滤（status）
- ✓ 支持排序（sort）
- ✓ 包含限流响应头

#### 2. 创建用户

**请求：**
```http
POST /api/v1/users HTTP/1.1
Host: api.example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "name": "王五",
  "email": "wangwu@example.com",
  "password": "SecurePass123!"
}
```

**响应：**
```http
HTTP/1.1 201 Created
Content-Type: application/json
Location: /api/v1/users/125

{
  "id": 125,
  "name": "王五",
  "email": "wangwu@example.com",
  "status": "active",
  "createdAt": "2024-03-19T11:00:00Z"
}
```

**符合的规范：**
- ✓ 使用POST方法创建资源
- ✓ 返回201状态码
- ✓ 包含Location响应头
- ✓ 响应体包含新创建的资源
- ✓ 不返回敏感字段（password）

#### 3. 更新用户（部分更新）

**请求：**
```http
PATCH /api/v1/users/123 HTTP/1.1
Host: api.example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "name": "张三（已更新）"
}
```

**响应：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": 123,
  "name": "张三（已更新）",
  "email": "zhangsan@example.com",
  "status": "active",
  "updatedAt": "2024-03-19T11:05:00Z"
}
```

**符合的规范：**
- ✓ 使用PATCH方法部分更新
- ✓ 返回200状态码
- ✓ 返回完整的更新后资源
- ✓ 未修改的字段保持不变

#### 4. 删除用户

**请求：**
```http
DELETE /api/v1/users/123 HTTP/1.1
Host: api.example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应：**
```http
HTTP/1.1 204 No Content
```

**符合的规范：**
- ✓ 使用DELETE方法删除资源
- ✓ 返回204状态码
- ✓ 无响应体

#### 5. 获取用户的订单（嵌套资源）

**请求：**
```http
GET /api/v1/users/123/orders HTTP/1.1
Host: api.example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**响应：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "data": [
    {
      "id": 456,
      "userId": 123,
      "total": 299.99,
      "status": "completed",
      "createdAt": "2024-03-15T10:00:00Z"
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 20
}
```

**符合的规范：**
- ✓ 使用嵌套路径表达资源关系
- ✓ 返回200状态码
- ✓ 包含分页元数据

### 常见错误示例与改进

#### 错误示例1：路径中使用动词

**❌ 错误：**
```http
POST /api/v1/createUser
GET /api/v1/getUsers
DELETE /api/v1/deleteUser/123
```

**✓ 正确：**
```http
POST /api/v1/users
GET /api/v1/users
DELETE /api/v1/users/123
```

**说明：** RESTful API使用名词表示资源，通过HTTP方法表达操作。

#### 错误示例2：使用单数名词

**❌ 错误：**
```http
GET /api/v1/user
GET /api/v1/user/123
```

**✓ 正确：**
```http
GET /api/v1/users
GET /api/v1/users/123
```

**说明：** 资源路径统一使用复数形式，保持一致性。

#### 错误示例3：不正确的状态码使用

**❌ 错误：**
```http
POST /api/v1/users
HTTP/1.1 200 OK
{
  "success": true,
  "data": { "id": 123 }
}
```

**✓ 正确：**
```http
POST /api/v1/users
HTTP/1.1 201 Created
Location: /api/v1/users/123
{
  "id": 123,
  "name": "张三"
}
```

**说明：** 创建资源应返回201，并包含Location头。

#### 错误示例4：错误响应格式不统一

**❌ 错误：**
```http
HTTP/1.1 400 Bad Request
{
  "msg": "Invalid email"
}

HTTP/1.1 404 Not Found
{
  "error": "User not found"
}
```

**✓ 正确：**
```http
HTTP/1.1 400 Bad Request
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": [
      {
        "field": "email",
        "message": "邮箱格式不正确"
      }
    ]
  }
}

HTTP/1.1 404 Not Found
{
  "error": {
    "code": "NOT_FOUND",
    "message": "用户不存在"
  }
}
```

**说明：** 所有错误响应应使用统一的格式。

#### 错误示例5：过度嵌套

**❌ 错误：**
```http
GET /api/v1/organizations/1/departments/2/teams/3/members/4/tasks/5
```

**✓ 正确：**
```http
GET /api/v1/tasks/5
或
GET /api/v1/tasks?memberId=4
```

**说明：** 避免超过3层嵌套，使用查询参数或直接访问资源。

#### 错误示例6：混用命名风格

**❌ 错误：**
```http
GET /api/v1/userProfiles
GET /api/v1/order_items
GET /api/v1/PaymentMethods
```

**✓ 正确：**
```http
GET /api/v1/user-profiles
GET /api/v1/order-items
GET /api/v1/payment-methods
```

**说明：** 统一使用kebab-case命名风格。

#### 错误示例7：滥用200状态码

**❌ 错误：**
```http
GET /api/v1/users/999
HTTP/1.1 200 OK
{
  "success": false,
  "error": "User not found"
}
```

**✓ 正确：**
```http
GET /api/v1/users/999
HTTP/1.1 404 Not Found
{
  "error": {
    "code": "NOT_FOUND",
    "message": "用户不存在"
  }
}
```

**说明：** 使用正确的HTTP状态码，不要在200响应中表示错误。

### 设计检查清单

在设计API时，使用以下清单确保符合规范：

- [ ] 资源路径使用复数名词
- [ ] 资源路径使用kebab-case命名
- [ ] 路径中不包含动词
- [ ] 包含版本号（/api/v1/）
- [ ] 使用正确的HTTP方法（GET/POST/PUT/PATCH/DELETE）
- [ ] 返回正确的HTTP状态码
- [ ] 创建资源返回201和Location头
- [ ] 错误响应使用统一格式
- [ ] 支持分页（offset/limit）
- [ ] 支持过滤和排序
- [ ] 包含认证机制（Authorization头）
- [ ] 包含限流响应头
- [ ] 嵌套层级不超过3层
- [ ] 文档完整（所有端点都有文档）
