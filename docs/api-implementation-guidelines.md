# API实现指南

本文档定义API实现的技术规范，确保API的健壮性、安全性和一致性。

## 目录

1. [错误处理](#错误处理)
2. [分页机制](#分页机制)
3. [过滤和搜索](#过滤和搜索)
4. [排序](#排序)
5. [字段选择](#字段选择)
6. [认证机制](#认证机制)
7. [授权机制](#授权机制)
8. [请求限流](#请求限流)
9. [幂等性保证](#幂等性保证)

---

## 错误处理

### 基本原则

API必须返回统一的错误响应格式，包含错误代码、消息和详细信息，帮助客户端准确理解和处理错误。

### 统一错误响应格式

所有错误响应（4xx和5xx）必须使用以下JSON格式：

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "人类可读的错误消息",
    "details": [
      {
        "field": "字段名",
        "message": "字段错误描述"
      }
    ]
  }
}
```

### 字段说明

#### error.code（必需）

错误代码，使用UPPER_SNAKE_CASE格式，用于程序化处理错误。

**命名规范：**
- 使用大写字母和下划线
- 清晰表达错误类型
- 保持简洁

**常用错误代码：**

| 错误代码 | HTTP状态码 | 说明 |
|---------|-----------|------|
| VALIDATION_ERROR | 400 | 请求参数验证失败 |
| INVALID_REQUEST | 400 | 请求格式错误 |
| UNAUTHORIZED | 401 | 未认证或认证失败 |
| TOKEN_EXPIRED | 401 | 访问令牌已过期 |
| FORBIDDEN | 403 | 无权限访问 |
| NOT_FOUND | 404 | 资源不存在 |
| METHOD_NOT_ALLOWED | 405 | HTTP方法不支持 |
| CONFLICT | 409 | 资源冲突 |
| DUPLICATE_RESOURCE | 409 | 资源已存在 |
| INVALID_DATA | 422 | 数据不符合业务规则 |
| RATE_LIMIT_EXCEEDED | 429 | 请求频率超限 |
| INTERNAL_ERROR | 500 | 服务器内部错误 |
| SERVICE_UNAVAILABLE | 503 | 服务不可用 |

#### error.message（必需）

人类可读的错误消息，使用用户的语言（中文或英文）。

**编写规范：**
- 清晰描述错误原因
- 避免技术术语（面向最终用户）
- 提供解决建议（如适用）
- 不暴露敏感信息（如数据库错误、堆栈跟踪）

**示例：**
```
✓ "邮箱格式不正确"
✓ "用户名已被注册"
✓ "订单数量必须大于0"
✗ "NullPointerException at line 42"
✗ "Database connection failed"
```

#### error.details（可选）

错误详细信息数组，用于字段级验证错误或多个错误。

**使用场景：**
- 表单验证错误（多个字段）
- 批量操作的部分失败
- 需要提供额外上下文信息

**字段：**
- `field`：字段名（使用camelCase）
- `message`：字段错误描述
- 其他自定义字段（如`code`、`value`等）

### 错误响应示例

#### 1. 字段验证错误（400）

```http
POST /api/v1/users
Content-Type: application/json

{
  "name": "",
  "email": "invalid-email",
  "age": -5
}
```

**响应：**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": [
      {
        "field": "name",
        "message": "姓名不能为空"
      },
      {
        "field": "email",
        "message": "邮箱格式不正确"
      },
      {
        "field": "age",
        "message": "年龄必须大于0"
      }
    ]
  }
}
```

#### 2. 认证失败（401）

```http
GET /api/v1/users/me
Authorization: Bearer invalid_token
```

**响应：**
```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer realm="API"
Content-Type: application/json

{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "认证失败，请提供有效的访问令牌"
  }
}
```

#### 3. Token过期（401）

```http
GET /api/v1/users/me
Authorization: Bearer expired_token
```

**响应：**
```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": {
    "code": "TOKEN_EXPIRED",
    "message": "访问令牌已过期，请刷新令牌"
  }
}
```

#### 4. 权限不足（403）

```http
DELETE /api/v1/users/123
Authorization: Bearer valid_token
```

**响应：**
```http
HTTP/1.1 403 Forbidden
Content-Type: application/json

{
  "error": {
    "code": "FORBIDDEN",
    "message": "您没有权限删除此用户"
  }
}
```

#### 5. 资源不存在（404）

```http
GET /api/v1/users/999
```

**响应：**
```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{
  "error": {
    "code": "NOT_FOUND",
    "message": "用户不存在"
  }
}
```

#### 6. 资源冲突（409）

```http
POST /api/v1/users
Content-Type: application/json

{
  "email": "existing@example.com"
}
```

**响应：**
```http
HTTP/1.1 409 Conflict
Content-Type: application/json

{
  "error": {
    "code": "DUPLICATE_RESOURCE",
    "message": "该邮箱已被注册"
  }
}
```

#### 7. 业务逻辑错误（422）

```http
POST /api/v1/orders
Content-Type: application/json

{
  "productId": 123,
  "quantity": 100
}
```

**响应：**
```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json

{
  "error": {
    "code": "INVALID_DATA",
    "message": "库存不足，当前库存仅剩50件"
  }
}
```

#### 8. 请求频率超限（429）

```http
GET /api/v1/users
```

**响应：**
```http
HTTP/1.1 429 Too Many Requests
Retry-After: 60
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1710842460
Content-Type: application/json

{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "请求过于频繁，请60秒后再试"
  }
}
```

#### 9. 服务器内部错误（500）

```http
GET /api/v1/users
```

**响应：**
```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json

{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "服务器内部错误，请稍后重试"
  }
}
```

**注意：** 不要在生产环境的错误响应中暴露堆栈跟踪、数据库错误等敏感信息。

#### 10. 服务不可用（503）

```http
GET /api/v1/users
```

**响应：**
```http
HTTP/1.1 503 Service Unavailable
Retry-After: 300
Content-Type: application/json

{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "服务维护中，预计5分钟后恢复"
  }
}
```

### 错误处理最佳实践

#### 1. 一致性

所有API端点使用相同的错误响应格式，不要在不同端点使用不同的格式。

#### 2. 信息充分但不过度

提供足够的信息帮助客户端处理错误，但不要暴露敏感的系统信息。

**✓ 好的错误消息：**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "邮箱格式不正确"
  }
}
```

**✗ 不好的错误消息：**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "SQLException: Duplicate entry 'user@example.com' for key 'users.email'"
  }
}
```

#### 3. 使用正确的HTTP状态码

不要所有错误都返回200，在响应体中表示错误。

**✗ 错误做法：**
```http
HTTP/1.1 200 OK
{
  "success": false,
  "error": "User not found"
}
```

**✓ 正确做法：**
```http
HTTP/1.1 404 Not Found
{
  "error": {
    "code": "NOT_FOUND",
    "message": "用户不存在"
  }
}
```

#### 4. 国际化支持

如果API支持多语言，根据Accept-Language请求头返回相应语言的错误消息。

```http
GET /api/v1/users/999
Accept-Language: en-US
```

**响应：**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found"
  }
}
```

#### 5. 日志记录

服务端应记录详细的错误日志（包括堆栈跟踪），但不要将这些信息返回给客户端。

**服务端日志：**
```
[ERROR] 2024-03-19 10:00:00 - GET /api/v1/users/123
java.lang.NullPointerException: User object is null
    at com.example.UserService.getUser(UserService.java:42)
    ...
```

**客户端响应：**
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "服务器内部错误，请稍后重试"
  }
}
```

#### 6. 错误追踪ID

对于5xx错误，可以在响应中包含追踪ID，便于客户端报告问题和服务端排查。

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "服务器内部错误，请稍后重试",
    "traceId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}
```

### 错误代码设计指南

#### 分类命名

使用前缀对错误代码分类：

- `VALIDATION_*` - 验证错误
- `AUTH_*` - 认证相关错误
- `PERMISSION_*` - 权限相关错误
- `RESOURCE_*` - 资源相关错误
- `RATE_LIMIT_*` - 限流相关错误
- `INTERNAL_*` - 内部错误

**示例：**
```
VALIDATION_REQUIRED_FIELD
VALIDATION_INVALID_FORMAT
AUTH_TOKEN_EXPIRED
AUTH_INVALID_CREDENTIALS
PERMISSION_DENIED
RESOURCE_NOT_FOUND
RESOURCE_CONFLICT
RATE_LIMIT_EXCEEDED
INTERNAL_DATABASE_ERROR
```

#### 避免过度细分

不要为每个可能的错误创建唯一的错误代码，保持错误代码数量可管理。

**✗ 过度细分：**
```
USER_EMAIL_INVALID
USER_NAME_INVALID
USER_PHONE_INVALID
ORDER_QUANTITY_INVALID
ORDER_PRICE_INVALID
```

**✓ 合理分类：**
```
VALIDATION_ERROR (配合details字段说明具体字段)
```

## 分页机制

### 基本原则

API必须为返回集合的端点提供分页支持，避免一次性返回大量数据，提高性能和用户体验。

### 分页方式：基于偏移量（Offset-based）

使用`offset`和`limit`查询参数实现分页。

#### 查询参数

| 参数 | 类型 | 必需 | 默认值 | 说明 |
|------|------|------|--------|------|
| offset | integer | 否 | 0 | 偏移量，从0开始 |
| limit | integer | 否 | 20 | 每页数量，最大100 |

#### 请求示例

```http
GET /api/v1/users?offset=0&limit=20 HTTP/1.1
Host: api.example.com
Authorization: Bearer token
```

#### 响应格式

响应必须包含以下字段：

```json
{
  "data": [...],
  "total": 156,
  "offset": 0,
  "limit": 20
}
```

**字段说明：**
- `data`：当前页的数据数组
- `total`：总记录数
- `offset`：当前偏移量
- `limit`：当前每页数量

#### 完整示例

**请求：**
```http
GET /api/v1/users?offset=20&limit=10 HTTP/1.1
```

**响应：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "data": [
    {
      "id": 21,
      "name": "用户21",
      "email": "user21@example.com"
    },
    {
      "id": 22,
      "name": "用户22",
      "email": "user22@example.com"
    }
    // ... 共10条记录
  ],
  "total": 156,
  "offset": 20,
  "limit": 10
}
```

**计算页码：**
- 当前页：`Math.floor(offset / limit) + 1` = `Math.floor(20 / 10) + 1` = 3
- 总页数：`Math.ceil(total / limit)` = `Math.ceil(156 / 10)` = 16

### 参数验证

#### offset验证

- 必须是非负整数
- 如果offset >= total，返回空数组（不报错）

**示例：**
```http
GET /api/v1/users?offset=200&limit=20
```

**响应（total=156）：**
```json
{
  "data": [],
  "total": 156,
  "offset": 200,
  "limit": 20
}
```

#### limit验证

- 必须是正整数
- 最小值：1
- 最大值：100（可根据业务调整）
- 超过最大值时返回400错误

**示例（超过最大值）：**
```http
GET /api/v1/users?limit=500
```

**响应：**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "limit参数必须在1到100之间",
    "details": [
      {
        "field": "limit",
        "message": "最大值为100"
      }
    ]
  }
}
```

### 默认值

当客户端未提供分页参数时，使用默认值：

- `offset=0`
- `limit=20`

**示例：**
```http
GET /api/v1/users
```

等同于：
```http
GET /api/v1/users?offset=0&limit=20
```

### 空结果处理

当查询结果为空时，返回空数组，不返回404。

**响应：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "data": [],
  "total": 0,
  "offset": 0,
  "limit": 20
}
```

### 与过滤和排序结合

分页参数可以与过滤、排序参数组合使用。

**示例：**
```http
GET /api/v1/users?status=active&sort=-createdAt&offset=0&limit=20
```

**处理顺序：**
1. 应用过滤条件（status=active）
2. 应用排序（按createdAt降序）
3. 应用分页（offset=0, limit=20）
4. 计算total（过滤后的总数）

### 性能优化建议

#### 1. 索引优化

为常用的排序和过滤字段创建数据库索引。

```sql
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_status ON users(status);
```

#### 2. 避免COUNT(*)

对于大数据集，COUNT(*)可能很慢。考虑：
- 缓存总数
- 使用近似值
- 只在第一页计算总数

#### 3. 限制最大offset

对于非常大的offset，查询性能会下降。考虑：
- 限制最大offset（如10000）
- 引导用户使用过滤而非深度分页
- 对于深度分页场景，使用游标分页

### 游标分页（可选）

对于实时数据流或需要深度分页的场景，可以提供游标分页作为补充。

#### 请求参数

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| cursor | string | 否 | 游标，指向下一页的起始位置 |
| limit | integer | 否 | 每页数量 |

#### 响应格式

```json
{
  "data": [...],
  "nextCursor": "eyJpZCI6MTIzfQ==",
  "hasMore": true
}
```

**字段说明：**
- `data`：当前页数据
- `nextCursor`：下一页的游标（base64编码）
- `hasMore`：是否还有更多数据

#### 示例

**第一页：**
```http
GET /api/v1/posts?limit=10
```

**响应：**
```json
{
  "data": [
    {"id": 100, "title": "Post 100"},
    {"id": 99, "title": "Post 99"}
    // ... 共10条
  ],
  "nextCursor": "eyJpZCI6OTB9",
  "hasMore": true
}
```

**第二页：**
```http
GET /api/v1/posts?cursor=eyJpZCI6OTB9&limit=10
```

**优点：**
- 性能稳定，不受数据量影响
- 适合实时数据流
- 避免重复或遗漏数据

**缺点：**
- 不支持跳页
- 不能显示总页数
- 实现相对复杂

### 分页最佳实践

1. **默认启用分页**：所有返回集合的端点都应支持分页
2. **合理的默认值**：limit默认值通常为20-50
3. **限制最大值**：防止客户端请求过多数据
4. **返回总数**：帮助客户端计算总页数
5. **一致的响应格式**：所有分页端点使用相同的响应结构
6. **文档说明**：在API文档中明确说明分页参数和响应格式
7. **性能监控**：监控大offset查询的性能，及时优化

### 客户端使用示例

#### JavaScript

```javascript
async function fetchUsers(page = 1, pageSize = 20) {
  const offset = (page - 1) * pageSize;
  const response = await fetch(
    `/api/v1/users?offset=${offset}&limit=${pageSize}`,
    {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }
  );

  const result = await response.json();

  return {
    users: result.data,
    totalPages: Math.ceil(result.total / result.limit),
    currentPage: Math.floor(result.offset / result.limit) + 1
  };
}

// 使用
const { users, totalPages, currentPage } = await fetchUsers(1, 20);
```

#### Python

```python
import requests

def fetch_users(page=1, page_size=20):
    offset = (page - 1) * page_size
    response = requests.get(
        f'/api/v1/users',
        params={'offset': offset, 'limit': page_size},
        headers={'Authorization': f'Bearer {token}'}
    )

    result = response.json()

    return {
        'users': result['data'],
        'total_pages': (result['total'] + result['limit'] - 1) // result['limit'],
        'current_page': result['offset'] // result['limit'] + 1
    }

# 使用
data = fetch_users(page=1, page_size=20)
```

## 过滤和搜索

### 基本原则

API必须支持资源过滤和搜索，允许客户端精确查询所需数据，减少数据传输和客户端处理负担。

### 字段相等过滤

使用查询参数进行字段相等匹配，参数名为字段名。

#### 格式

```
?fieldName=value
```

#### 示例

**单个过滤条件：**
```http
GET /api/v1/users?status=active
```

返回所有status为active的用户。

**多个过滤条件（AND逻辑）：**
```http
GET /api/v1/users?status=active&role=admin
```

返回status为active且role为admin的用户。

#### 响应

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "data": [
    {
      "id": 1,
      "name": "张三",
      "status": "active",
      "role": "admin"
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": 20
}
```

### 多值过滤（OR逻辑）

使用逗号分隔多个值，表示OR逻辑。

#### 格式

```
?fieldName=value1,value2,value3
```

#### 示例

```http
GET /api/v1/users?status=active,pending
```

返回status为active或pending的用户。

**等价SQL：**
```sql
WHERE status IN ('active', 'pending')
```

### 范围过滤

使用方括号语法进行范围查询。

#### 支持的操作符

| 操作符 | 说明 | 示例 |
|--------|------|------|
| [gte] | 大于等于 | ?age[gte]=18 |
| [gt] | 大于 | ?age[gt]=18 |
| [lte] | 小于等于 | ?age[lte]=65 |
| [lt] | 小于 | ?age[lt]=65 |
| [ne] | 不等于 | ?status[ne]=deleted |

#### 示例

**单个范围条件：**
```http
GET /api/v1/users?age[gte]=18
```

返回年龄大于等于18的用户。

**组合范围条件：**
```http
GET /api/v1/users?age[gte]=18&age[lte]=65
```

返回年龄在18到65之间的用户。

**日期范围：**
```http
GET /api/v1/orders?createdAt[gte]=2024-01-01&createdAt[lte]=2024-12-31
```

返回2024年创建的订单。

### 模糊搜索

使用`search`或`q`参数进行全文搜索。

#### 格式

```
?search=keyword
或
?q=keyword
```

#### 示例

```http
GET /api/v1/users?search=张三
```

在用户的name、email等字段中搜索包含"张三"的记录。

**响应：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "data": [
    {
      "id": 1,
      "name": "张三",
      "email": "zhangsan@example.com"
    },
    {
      "id": 2,
      "name": "张三丰",
      "email": "zhangsanfeng@example.com"
    }
  ],
  "total": 2,
  "offset": 0,
  "limit": 20
}
```

#### 搜索字段

API文档应明确说明搜索哪些字段。

**示例：**
```
GET /api/v1/users?search=keyword

搜索字段：name, email, phone
```

### 字段特定搜索

对特定字段进行模糊搜索。

#### 格式

```
?fieldName[like]=pattern
```

#### 示例

```http
GET /api/v1/users?name[like]=张
```

返回name包含"张"的用户。

```http
GET /api/v1/users?email[like]=@gmail.com
```

返回email包含"@gmail.com"的用户。

### 布尔值过滤

布尔字段使用`true`或`false`字符串。

#### 示例

```http
GET /api/v1/users?isVerified=true
```

返回已验证的用户。

```http
GET /api/v1/products?inStock=false
```

返回缺货的产品。

### 空值过滤

使用特殊值表示空值查询。

#### 格式

```
?fieldName=null    # 查询字段为null的记录
?fieldName[ne]=null # 查询字段不为null的记录
```

#### 示例

```http
GET /api/v1/users?deletedAt=null
```

返回未删除的用户（deletedAt为null）。

```http
GET /api/v1/users?phone[ne]=null
```

返回有电话号码的用户。

### 嵌套字段过滤

使用点号访问嵌套字段。

#### 格式

```
?parent.child=value
```

#### 示例

```http
GET /api/v1/users?address.city=北京
```

返回地址在北京的用户。

**数据结构：**
```json
{
  "id": 1,
  "name": "张三",
  "address": {
    "city": "北京",
    "district": "朝阳区"
  }
}
```

### 过滤与分页、排序组合

过滤可以与分页、排序参数组合使用。

#### 示例

```http
GET /api/v1/users?status=active&role=admin&sort=-createdAt&offset=0&limit=20
```

**处理顺序：**
1. 应用过滤（status=active AND role=admin）
2. 应用排序（按createdAt降序）
3. 应用分页（offset=0, limit=20）
4. 计算total（过滤后的总数）

### 参数验证

#### 无效字段名

当客户端使用不存在的字段进行过滤时，返回400错误。

**请求：**
```http
GET /api/v1/users?invalidField=value
```

**响应：**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "无效的过滤字段",
    "details": [
      {
        "field": "invalidField",
        "message": "该字段不支持过滤"
      }
    ]
  }
}
```

#### 无效操作符

当使用不支持的操作符时，返回400错误。

**请求：**
```http
GET /api/v1/users?age[invalid]=18
```

**响应：**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "无效的操作符",
    "details": [
      {
        "field": "age",
        "message": "不支持的操作符: invalid"
      }
    ]
  }
}
```

#### 类型不匹配

当过滤值类型不匹配时，返回400错误。

**请求：**
```http
GET /api/v1/users?age=abc
```

**响应：**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "参数类型错误",
    "details": [
      {
        "field": "age",
        "message": "必须是整数"
      }
    ]
  }
}
```

### 高级过滤（可选）

对于复杂查询需求，可以提供高级过滤语法。

#### JSON过滤（可选）

使用JSON格式表达复杂查询条件。

**格式：**
```
?filter={"and":[{"status":"active"},{"or":[{"role":"admin"},{"role":"moderator"}]}]}
```

**等价SQL：**
```sql
WHERE status = 'active' AND (role = 'admin' OR role = 'moderator')
```

**注意：** 这种方式较复杂，仅在简单过滤无法满足需求时使用。

### 过滤最佳实践

#### 1. 明确支持的过滤字段

在API文档中列出所有支持过滤的字段。

**示例文档：**
```
GET /api/v1/users

支持的过滤字段：
- status (string): 用户状态 (active, inactive, suspended)
- role (string): 用户角色 (admin, user, guest)
- isVerified (boolean): 是否已验证
- createdAt (datetime): 创建时间，支持范围查询
- age (integer): 年龄，支持范围查询
```

#### 2. 性能优化

为常用的过滤字段创建数据库索引。

```sql
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at);
```

#### 3. 限制过滤复杂度

避免过于复杂的过滤条件导致性能问题：
- 限制同时使用的过滤字段数量
- 限制OR条件的数量
- 对复杂查询使用专门的搜索端点

#### 4. 一致的命名

过滤参数名与响应字段名保持一致（使用camelCase）。

**✓ 正确：**
```
GET /api/v1/users?createdAt[gte]=2024-01-01

响应字段：createdAt
```

**✗ 错误：**
```
GET /api/v1/users?created_at[gte]=2024-01-01

响应字段：createdAt
```

#### 5. 默认过滤

某些端点可以应用默认过滤（如不返回已删除的记录）。

**示例：**
```
GET /api/v1/users

默认过滤：deletedAt=null（不返回已删除用户）

如需包含已删除用户：
GET /api/v1/users?includeDeleted=true
```

### 客户端使用示例

#### JavaScript

```javascript
function buildQueryString(filters) {
  const params = new URLSearchParams();

  for (const [key, value] of Object.entries(filters)) {
    if (value !== null && value !== undefined) {
      params.append(key, value);
    }
  }

  return params.toString();
}

// 使用
const filters = {
  status: 'active',
  role: 'admin',
  'age[gte]': 18,
  'createdAt[gte]': '2024-01-01'
};

const queryString = buildQueryString(filters);
// status=active&role=admin&age[gte]=18&createdAt[gte]=2024-01-01

const response = await fetch(`/api/v1/users?${queryString}`);
```

#### Python

```python
def fetch_users(filters):
    params = {k: v for k, v in filters.items() if v is not None}

    response = requests.get(
        '/api/v1/users',
        params=params,
        headers={'Authorization': f'Bearer {token}'}
    )

    return response.json()

# 使用
filters = {
    'status': 'active',
    'role': 'admin',
    'age[gte]': 18,
    'createdAt[gte]': '2024-01-01'
}

result = fetch_users(filters)
```

## 排序

### 基本原则

API必须支持结果排序，允许客户端指定排序字段和顺序，提供灵活的数据展示方式。

### 排序参数

使用`sort`查询参数指定排序规则。

#### 格式

```
?sort=fieldName        # 升序
?sort=-fieldName       # 降序（使用减号前缀）
```

### 单字段排序

#### 升序排序

```http
GET /api/v1/users?sort=name
```

按name字段升序排序（A-Z）。

**响应：**
```json
{
  "data": [
    {"id": 1, "name": "Alice"},
    {"id": 2, "name": "Bob"},
    {"id": 3, "name": "Charlie"}
  ],
  "total": 3,
  "offset": 0,
  "limit": 20
}
```

#### 降序排序

```http
GET /api/v1/users?sort=-createdAt
```

按createdAt字段降序排序（最新的在前）。

**响应：**
```json
{
  "data": [
    {"id": 3, "name": "Charlie", "createdAt": "2024-03-19T10:00:00Z"},
    {"id": 2, "name": "Bob", "createdAt": "2024-03-18T15:00:00Z"},
    {"id": 1, "name": "Alice", "createdAt": "2024-03-17T09:00:00Z"}
  ],
  "total": 3,
  "offset": 0,
  "limit": 20
}
```

### 多字段排序

使用逗号分隔多个排序字段，按顺序应用。

#### 格式

```
?sort=field1,-field2,field3
```

#### 示例

```http
GET /api/v1/users?sort=status,-createdAt
```

**排序规则：**
1. 首先按status升序排序
2. status相同的记录按createdAt降序排序

**响应：**
```json
{
  "data": [
    {"id": 1, "status": "active", "createdAt": "2024-03-19T10:00:00Z"},
    {"id": 2, "status": "active", "createdAt": "2024-03-18T15:00:00Z"},
    {"id": 3, "status": "inactive", "createdAt": "2024-03-17T09:00:00Z"}
  ],
  "total": 3,
  "offset": 0,
  "limit": 20
}
```

**等价SQL：**
```sql
ORDER BY status ASC, created_at DESC
```

### 默认排序

当客户端未指定排序参数时，API应使用合理的默认排序。

#### 推荐默认排序

- **时间序列数据**：按创建时间降序（`-createdAt`）
- **列表数据**：按主键升序（`id`）
- **用户数据**：按名称升序（`name`）

#### 示例

```http
GET /api/v1/orders
```

默认排序：`-createdAt`（最新订单在前）

**响应：**
```json
{
  "data": [
    {"id": 100, "createdAt": "2024-03-19T10:00:00Z"},
    {"id": 99, "createdAt": "2024-03-18T15:00:00Z"}
  ],
  "total": 100,
  "offset": 0,
  "limit": 20
}
```

### 排序与过滤、分页组合

排序可以与过滤、分页参数组合使用。

#### 示例

```http
GET /api/v1/users?status=active&sort=-createdAt&offset=0&limit=20
```

**处理顺序：**
1. 应用过滤（status=active）
2. 应用排序（按createdAt降序）
3. 应用分页（offset=0, limit=20）

### 嵌套字段排序

使用点号访问嵌套字段进行排序。

#### 格式

```
?sort=parent.child
```

#### 示例

```http
GET /api/v1/users?sort=address.city
```

按用户地址的城市字段排序。

**数据结构：**
```json
{
  "id": 1,
  "name": "张三",
  "address": {
    "city": "北京",
    "district": "朝阳区"
  }
}
```

### 参数验证

#### 无效字段名

当客户端使用不存在的字段进行排序时，返回400错误。

**请求：**
```http
GET /api/v1/users?sort=invalidField
```

**响应：**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "无效的排序字段",
    "details": [
      {
        "field": "sort",
        "message": "字段 'invalidField' 不支持排序"
      }
    ]
  }
}
```

#### 格式错误

当排序参数格式错误时，返回400错误。

**请求：**
```http
GET /api/v1/users?sort=name:asc
```

**响应：**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "排序参数格式错误",
    "details": [
      {
        "field": "sort",
        "message": "正确格式: fieldName 或 -fieldName"
      }
    ]
  }
}
```

### 排序字段限制

#### 可排序字段

并非所有字段都应支持排序。在API文档中明确列出可排序字段。

**推荐可排序字段：**
- 数值字段（id, age, price, quantity）
- 日期时间字段（createdAt, updatedAt, publishedAt）
- 状态字段（status, priority）
- 名称字段（name, title）

**不推荐排序的字段：**
- 大文本字段（description, content）
- 二进制字段（avatar, file）
- 复杂对象字段

#### 示例文档

```
GET /api/v1/users

支持的排序字段：
- id: 用户ID
- name: 用户名
- email: 邮箱
- status: 状态
- createdAt: 创建时间
- updatedAt: 更新时间

默认排序: -createdAt
```

### 性能优化

#### 1. 索引优化

为常用的排序字段创建数据库索引。

```sql
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_name ON users(name);
CREATE INDEX idx_orders_status_created_at ON orders(status, created_at);
```

#### 2. 复合索引

对于常见的多字段排序组合，创建复合索引。

```sql
-- 支持 ?status=active&sort=-createdAt
CREATE INDEX idx_users_status_created_at ON users(status, created_at DESC);
```

#### 3. 限制排序字段数量

限制同时使用的排序字段数量（建议不超过3个）。

**请求：**
```http
GET /api/v1/users?sort=field1,field2,field3,field4
```

**响应：**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "排序字段过多，最多支持3个字段"
  }
}
```

### 特殊排序需求

#### 自定义排序顺序

对于枚举类型字段，可能需要自定义排序顺序。

**示例：优先级排序**
```
priority: high > medium > low
```

**实现方式：**
```sql
ORDER BY
  CASE priority
    WHEN 'high' THEN 1
    WHEN 'medium' THEN 2
    WHEN 'low' THEN 3
  END
```

#### 空值处理

明确空值在排序中的位置。

**推荐规则：**
- 升序：空值在最后（NULLS LAST）
- 降序：空值在最后（NULLS LAST）

**示例：**
```sql
ORDER BY name ASC NULLS LAST
ORDER BY created_at DESC NULLS LAST
```

### 排序最佳实践

#### 1. 提供合理的默认排序

不要返回无序的结果，始终应用默认排序。

**✓ 正确：**
```
GET /api/v1/users
默认排序: -createdAt
```

**✗ 错误：**
```
GET /api/v1/users
无排序（数据库返回顺序不确定）
```

#### 2. 一致的命名

排序参数名与响应字段名保持一致。

**✓ 正确：**
```
GET /api/v1/users?sort=-createdAt

响应字段：createdAt
```

**✗ 错误：**
```
GET /api/v1/users?sort=-created_at

响应字段：createdAt
```

#### 3. 文档说明

在API文档中明确说明：
- 支持的排序字段
- 默认排序规则
- 排序参数格式
- 多字段排序的优先级

#### 4. 性能监控

监控排序查询的性能，识别慢查询并优化索引。

#### 5. 分页一致性

使用排序时，确保分页结果的一致性。建议在排序字段中包含唯一字段（如id）。

**✓ 推荐：**
```
?sort=-createdAt,id
```

这样即使createdAt相同，也能保证排序稳定。

### 客户端使用示例

#### JavaScript

```javascript
function buildSortParam(sortFields) {
  // sortFields: [{ field: 'createdAt', order: 'desc' }, { field: 'name', order: 'asc' }]

  const sortParts = sortFields.map(({ field, order }) => {
    return order === 'desc' ? `-${field}` : field;
  });

  return sortParts.join(',');
}

// 使用
const sortFields = [
  { field: 'status', order: 'asc' },
  { field: 'createdAt', order: 'desc' }
];

const sortParam = buildSortParam(sortFields);
// "status,-createdAt"

const response = await fetch(`/api/v1/users?sort=${sortParam}`);
```

#### Python

```python
def build_sort_param(sort_fields):
    """
    sort_fields: [{'field': 'createdAt', 'order': 'desc'}, {'field': 'name', 'order': 'asc'}]
    """
    sort_parts = []
    for item in sort_fields:
        field = item['field']
        order = item.get('order', 'asc')
        sort_parts.append(f"-{field}" if order == 'desc' else field)

    return ','.join(sort_parts)

# 使用
sort_fields = [
    {'field': 'status', 'order': 'asc'},
    {'field': 'createdAt', 'order': 'desc'}
]

sort_param = build_sort_param(sort_fields)
# "status,-createdAt"

response = requests.get(
    '/api/v1/users',
    params={'sort': sort_param},
    headers={'Authorization': f'Bearer {token}'}
)
```

### 排序与UI组件集成

#### 表格排序

```javascript
// 表格列配置
const columns = [
  {
    key: 'name',
    title: '姓名',
    sortable: true
  },
  {
    key: 'createdAt',
    title: '创建时间',
    sortable: true,
    defaultSort: 'desc'
  }
];

// 处理排序点击
function handleSort(field, order) {
  const sortParam = order === 'desc' ? `-${field}` : field;
  fetchUsers({ sort: sortParam });
}
```

## 字段选择

### 基本原则

API应支持字段选择（Sparse Fieldsets），允许客户端只获取需要的字段，减少数据传输量，提高性能。

### 字段选择参数

使用`fields`查询参数指定需要返回的字段。

#### 格式

```
?fields=field1,field2,field3
```

### 基本用法

#### 选择部分字段

```http
GET /api/v1/users/123?fields=id,name,email
```

**响应：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": 123,
  "name": "张三",
  "email": "zhangsan@example.com"
}
```

**完整资源（不使用fields）：**
```json
{
  "id": 123,
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "address": "北京市朝阳区",
  "status": "active",
  "role": "user",
  "createdAt": "2024-03-19T10:00:00Z",
  "updatedAt": "2024-03-19T10:00:00Z"
}
```

#### 列表资源的字段选择

```http
GET /api/v1/users?fields=id,name&limit=10
```

**响应：**
```json
{
  "data": [
    {"id": 1, "name": "张三"},
    {"id": 2, "name": "李四"},
    {"id": 3, "name": "王五"}
  ],
  "total": 100,
  "offset": 0,
  "limit": 10
}
```

### 嵌套字段选择

使用点号选择嵌套对象的字段。

#### 格式

```
?fields=id,name,address.city,address.district
```

#### 示例

```http
GET /api/v1/users/123?fields=id,name,address.city
```

**响应：**
```json
{
  "id": 123,
  "name": "张三",
  "address": {
    "city": "北京"
  }
}
```

**完整address对象：**
```json
{
  "address": {
    "city": "北京",
    "district": "朝阳区",
    "street": "建国路1号",
    "zipCode": "100000"
  }
}
```

### 默认字段

当客户端未指定`fields`参数时，返回所有字段（除敏感字段外）。

#### 示例

```http
GET /api/v1/users/123
```

返回所有非敏感字段。

### 敏感字段保护

某些敏感字段永远不应返回，即使客户端明确请求。

#### 敏感字段示例

- `password`：密码哈希
- `passwordHash`：密码哈希
- `salt`：密码盐值
- `privateKey`：私钥
- `secretKey`：密钥
- `internalNotes`：内部备注

#### 示例

```http
GET /api/v1/users/123?fields=id,name,password
```

**响应：**
```json
{
  "id": 123,
  "name": "张三"
}
```

**说明：** password字段被自动排除，不返回给客户端。

### 字段别名（可选）

对于复杂的嵌套结构，可以支持字段别名简化请求。

#### 示例

```http
GET /api/v1/users?fields=basic
```

**预定义别名：**
- `basic`: `id,name,email`
- `full`: 所有非敏感字段
- `minimal`: `id,name`

**响应：**
```json
{
  "data": [
    {
      "id": 1,
      "name": "张三",
      "email": "zhangsan@example.com"
    }
  ]
}
```

### 参数验证

#### 无效字段名

当客户端请求不存在的字段时，返回400错误。

**请求：**
```http
GET /api/v1/users/123?fields=id,invalidField
```

**响应：**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "无效的字段名",
    "details": [
      {
        "field": "fields",
        "message": "字段 'invalidField' 不存在"
      }
    ]
  }
}
```

#### 格式错误

当字段参数格式错误时，返回400错误。

**请求：**
```http
GET /api/v1/users/123?fields=id;name
```

**响应：**
```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "字段参数格式错误",
    "details": [
      {
        "field": "fields",
        "message": "使用逗号分隔字段名"
      }
    ]
  }
}
```

### 字段选择与其他参数组合

字段选择可以与过滤、排序、分页参数组合使用。

#### 示例

```http
GET /api/v1/users?status=active&sort=-createdAt&fields=id,name,email&offset=0&limit=20
```

**响应：**
```json
{
  "data": [
    {
      "id": 1,
      "name": "张三",
      "email": "zhangsan@example.com"
    },
    {
      "id": 2,
      "name": "李四",
      "email": "lisi@example.com"
    }
  ],
  "total": 50,
  "offset": 0,
  "limit": 20
}
```

### 性能优化

#### 1. 数据库查询优化

只查询客户端请求的字段，减少数据库负载。

**SQL示例：**
```sql
-- 不使用字段选择
SELECT * FROM users WHERE id = 123;

-- 使用字段选择
SELECT id, name, email FROM users WHERE id = 123;
```

#### 2. 关联查询优化

避免加载不需要的关联数据。

**示例：**
```http
GET /api/v1/users/123?fields=id,name
```

不需要加载用户的订单、地址等关联数据。

#### 3. 序列化优化

在序列化阶段排除不需要的字段，而不是查询后再过滤。

### 字段选择最佳实践

#### 1. 明确可选字段

在API文档中列出所有可选字段。

**示例文档：**
```
GET /api/v1/users/{id}

可选字段：
- id: 用户ID
- name: 用户名
- email: 邮箱
- phone: 电话
- address: 地址对象
  - address.city: 城市
  - address.district: 区县
  - address.street: 街道
- status: 状态
- role: 角色
- createdAt: 创建时间
- updatedAt: 更新时间

敏感字段（不可选）：
- password: 密码哈希
- internalNotes: 内部备注

默认返回：所有非敏感字段
```

#### 2. 保持向后兼容

添加新字段时，不应影响现有客户端。

**✓ 正确：**
```
新增字段默认包含在响应中（不使用fields时）
使用fields的客户端不受影响
```

**✗ 错误：**
```
新增字段导致响应结构变化
破坏现有客户端的解析逻辑
```

#### 3. 合理的默认字段集

默认返回最常用的字段，避免返回过多不必要的数据。

**示例：**
```
默认字段：id, name, email, status, createdAt
完整字段：需要使用 ?fields=* 或不传fields参数
```

#### 4. 一致的命名

字段参数名与响应字段名保持一致。

**✓ 正确：**
```
GET /api/v1/users?fields=createdAt

响应字段：createdAt
```

**✗ 错误：**
```
GET /api/v1/users?fields=created_at

响应字段：createdAt
```

#### 5. 文档说明

在API文档中明确说明：
- 支持的字段列表
- 敏感字段（不可选）
- 默认字段集
- 嵌套字段的访问方式

### 客户端使用示例

#### JavaScript

```javascript
function buildFieldsParam(fields) {
  // fields: ['id', 'name', 'email', 'address.city']
  return fields.join(',');
}

// 使用
const fields = ['id', 'name', 'email'];
const fieldsParam = buildFieldsParam(fields);
// "id,name,email"

const response = await fetch(
  `/api/v1/users/123?fields=${fieldsParam}`,
  {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  }
);

const user = await response.json();
// { id: 123, name: "张三", email: "zhangsan@example.com" }
```

#### Python

```python
def fetch_user(user_id, fields=None):
    params = {}
    if fields:
        params['fields'] = ','.join(fields)

    response = requests.get(
        f'/api/v1/users/{user_id}',
        params=params,
        headers={'Authorization': f'Bearer {token}'}
    )

    return response.json()

# 使用
user = fetch_user(123, fields=['id', 'name', 'email'])
# {'id': 123, 'name': '张三', 'email': 'zhangsan@example.com'}
```

### 高级用法（可选）

#### 排除字段

使用减号前缀排除特定字段。

**格式：**
```
?fields=-field1,-field2
```

**示例：**
```http
GET /api/v1/users/123?fields=-createdAt,-updatedAt
```

返回除createdAt和updatedAt外的所有字段。

**注意：** 这种方式增加了复杂度，仅在必要时使用。

#### 字段展开

对于关联资源，支持展开（embed）。

**格式：**
```
?fields=id,name&expand=orders
```

**示例：**
```http
GET /api/v1/users/123?fields=id,name&expand=orders
```

**响应：**
```json
{
  "id": 123,
  "name": "张三",
  "orders": [
    {"id": 1, "total": 299.99},
    {"id": 2, "total": 199.99}
  ]
}
```

**说明：** 这种方式可以减少API调用次数，但会增加响应大小和服务器负载。

## 认证机制

### 基本原则

API必须实现安全的认证机制，验证客户端身份，确保只有合法用户才能访问受保护的资源。

### Bearer Token认证（推荐）

使用Bearer Token认证方案，基于JWT（JSON Web Token）实现。

#### 认证流程

```
1. 用户登录 → 服务器验证凭证
2. 服务器生成JWT token → 返回给客户端
3. 客户端在后续请求中携带token
4. 服务器验证token → 允许或拒绝访问
```

### Token类型

#### Access Token（访问令牌）

用于访问受保护的API资源。

**特性：**
- 短期有效（推荐15分钟）
- 包含用户身份和权限信息
- 无状态验证

**格式：**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

**JWT Payload示例：**
```json
{
  "sub": "123",
  "name": "张三",
  "email": "zhangsan@example.com",
  "role": "user",
  "iat": 1710842400,
  "exp": 1710843300
}
```

**字段说明：**
- `sub`：用户ID（subject）
- `iat`：签发时间（issued at）
- `exp`：过期时间（expiration）
- 其他自定义字段（name, email, role等）

#### Refresh Token（刷新令牌）

用于获取新的Access Token，无需重新登录。

**特性：**
- 长期有效（推荐7天）
- 只能用于刷新Access Token
- 应安全存储（HttpOnly Cookie或安全存储）

### 认证端点

#### 1. 登录

**请求：**
```http
POST /api/v1/auth/login HTTP/1.1
Content-Type: application/json

{
  "email": "zhangsan@example.com",
  "password": "SecurePass123!"
}
```

**响应（成功）：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tokenType": "Bearer",
  "expiresIn": 900,
  "user": {
    "id": 123,
    "name": "张三",
    "email": "zhangsan@example.com",
    "role": "user"
  }
}
```

**字段说明：**
- `accessToken`：访问令牌
- `refreshToken`：刷新令牌
- `tokenType`：令牌类型（固定为"Bearer"）
- `expiresIn`：访问令牌有效期（秒）
- `user`：用户基本信息

**响应（失败）：**
```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "邮箱或密码错误"
  }
}
```

#### 2. 刷新Token

**请求：**
```http
POST /api/v1/auth/refresh HTTP/1.1
Content-Type: application/json

{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应（成功）：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tokenType": "Bearer",
  "expiresIn": 900
}
```

**响应（失败 - Token无效）：**
```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "error": {
    "code": "INVALID_REFRESH_TOKEN",
    "message": "刷新令牌无效或已过期"
  }
}
```

#### 3. 登出

**请求：**
```http
POST /api/v1/auth/logout HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应：**
```http
HTTP/1.1 204 No Content
```

**说明：** 服务端应将Refresh Token加入黑名单，使其失效。

### 使用Access Token访问API

#### 请求格式

在Authorization请求头中携带Access Token。

```http
GET /api/v1/users/me HTTP/1.1
Host: api.example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**格式：**
```
Authorization: Bearer <access_token>
```

#### 成功响应

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": 123,
  "name": "张三",
  "email": "zhangsan@example.com"
}
```

#### 错误响应

##### 未提供Token

```http
GET /api/v1/users/me HTTP/1.1
```

**响应：**
```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer realm="API"
Content-Type: application/json

{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "未提供认证令牌"
  }
}
```

##### Token无效

```http
GET /api/v1/users/me HTTP/1.1
Authorization: Bearer invalid_token
```

**响应：**
```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer error="invalid_token"
Content-Type: application/json

{
  "error": {
    "code": "INVALID_TOKEN",
    "message": "认证令牌无效"
  }
}
```

##### Token过期

```http
GET /api/v1/users/me HTTP/1.1
Authorization: Bearer expired_token
```

**响应：**
```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer error="invalid_token", error_description="The access token expired"
Content-Type: application/json

{
  "error": {
    "code": "TOKEN_EXPIRED",
    "message": "访问令牌已过期，请刷新令牌"
  }
}
```

### Token生命周期管理

#### Access Token过期处理

客户端应实现自动刷新机制。

**流程：**
```
1. 请求API → 收到401 TOKEN_EXPIRED
2. 使用Refresh Token请求新的Access Token
3. 使用新Token重试原请求
```

**JavaScript示例：**
```javascript
async function fetchWithAuth(url, options = {}) {
  let token = getAccessToken();

  let response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${token}`
    }
  });

  // Token过期，尝试刷新
  if (response.status === 401) {
    const error = await response.json();
    if (error.error.code === 'TOKEN_EXPIRED') {
      // 刷新token
      const newToken = await refreshAccessToken();
      setAccessToken(newToken);

      // 重试请求
      response = await fetch(url, {
        ...options,
        headers: {
          ...options.headers,
          'Authorization': `Bearer ${newToken}`
        }
      });
    }
  }

  return response;
}

async function refreshAccessToken() {
  const refreshToken = getRefreshToken();

  const response = await fetch('/api/v1/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken })
  });

  if (!response.ok) {
    // Refresh token也失效，需要重新登录
    redirectToLogin();
    throw new Error('Refresh token expired');
  }

  const data = await response.json();
  return data.accessToken;
}
```

#### Refresh Token轮换

为提高安全性，每次刷新Access Token时，同时生成新的Refresh Token。

**请求：**
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refreshToken": "old_refresh_token"
}
```

**响应：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "accessToken": "new_access_token",
  "refreshToken": "new_refresh_token",
  "tokenType": "Bearer",
  "expiresIn": 900
}
```

**说明：** 旧的Refresh Token立即失效。

### Token存储

#### 客户端存储建议

| 存储方式 | 安全性 | 适用场景 |
|---------|--------|---------|
| Memory（内存） | 高 | 单页应用（刷新页面会丢失） |
| LocalStorage | 低 | 不推荐（易受XSS攻击） |
| SessionStorage | 中 | 短期会话 |
| HttpOnly Cookie | 高 | Web应用（推荐） |
| Secure Storage | 高 | 移动应用 |

**推荐方案：**
- **Web应用**：Access Token存内存，Refresh Token存HttpOnly Cookie
- **移动应用**：使用平台提供的安全存储（Keychain、KeyStore）
- **桌面应用**：使用加密存储

### 安全最佳实践

#### 1. HTTPS强制

所有认证相关的请求必须使用HTTPS。

```
✓ https://api.example.com/auth/login
✗ http://api.example.com/auth/login
```

#### 2. Token签名验证

使用强加密算法（如HS256、RS256）签名JWT。

```javascript
// 服务端验证
const jwt = require('jsonwebtoken');

function verifyToken(token) {
  try {
    const decoded = jwt.verify(token, SECRET_KEY);
    return decoded;
  } catch (error) {
    throw new Error('Invalid token');
  }
}
```

#### 3. 短期Access Token

Access Token有效期应尽可能短（推荐15分钟）。

```javascript
const accessToken = jwt.sign(
  { sub: userId, role: userRole },
  SECRET_KEY,
  { expiresIn: '15m' }
);
```

#### 4. Refresh Token黑名单

维护Refresh Token黑名单，支持强制登出。

```javascript
// 登出时将Refresh Token加入黑名单
async function logout(refreshToken) {
  await redis.set(`blacklist:${refreshToken}`, '1', 'EX', 604800); // 7天
}

// 刷新时检查黑名单
async function refresh(refreshToken) {
  const isBlacklisted = await redis.exists(`blacklist:${refreshToken}`);
  if (isBlacklisted) {
    throw new Error('Token has been revoked');
  }
  // 继续刷新流程...
}
```

#### 5. 限制登录尝试

防止暴力破解，限制登录失败次数。

```javascript
// 登录失败计数
async function checkLoginAttempts(email) {
  const key = `login_attempts:${email}`;
  const attempts = await redis.incr(key);

  if (attempts === 1) {
    await redis.expire(key, 900); // 15分钟过期
  }

  if (attempts > 5) {
    throw new Error('Too many login attempts. Please try again later.');
  }
}
```

#### 6. 不在URL中传递Token

永远不要在URL查询参数中传递Token。

**✗ 错误：**
```
GET /api/v1/users?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**✓ 正确：**
```
GET /api/v1/users
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 7. Token内容最小化

JWT中只包含必要的信息，避免敏感数据。

**✓ 推荐：**
```json
{
  "sub": "123",
  "role": "user",
  "iat": 1710842400,
  "exp": 1710843300
}
```

**✗ 不推荐：**
```json
{
  "sub": "123",
  "password": "hashed_password",
  "creditCard": "1234-5678-9012-3456",
  "ssn": "123-45-6789"
}
```

### 其他认证方式（可选）

#### API Key认证

适用于服务端到服务端的调用。

**请求：**
```http
GET /api/v1/data HTTP/1.1
X-API-Key: your_api_key_here
```

**特点：**
- 长期有效
- 无需用户交互
- 适合自动化脚本和服务集成

#### OAuth 2.0

适用于第三方应用授权。

**流程：**
```
1. 用户授权第三方应用
2. 第三方应用获取Authorization Code
3. 使用Code换取Access Token
4. 使用Token访问用户资源
```

**适用场景：**
- 社交登录（微信、GitHub等）
- 第三方应用集成
- 需要细粒度权限控制

### 认证最佳实践总结

1. **使用Bearer Token + JWT**：标准、无状态、易于扩展
2. **短期Access Token**：15分钟有效期
3. **长期Refresh Token**：7天有效期，支持轮换
4. **HTTPS强制**：所有认证请求必须加密
5. **Token黑名单**：支持强制登出
6. **限制登录尝试**：防止暴力破解
7. **安全存储**：Web用HttpOnly Cookie，移动用安全存储
8. **自动刷新**：客户端实现Token自动刷新机制

## 授权机制

### 基本原则

API必须实现细粒度的授权控制，确保用户只能访问有权限的资源，遵循最小权限原则。

### 授权模型

#### 基于角色的访问控制（RBAC）

根据用户角色授予权限。

**常见角色：**
- `admin`：管理员，拥有所有权限
- `moderator`：版主，拥有内容管理权限
- `user`：普通用户，基本权限
- `guest`：访客，只读权限

**示例：**
```json
{
  "sub": "123",
  "role": "admin",
  "permissions": ["users:read", "users:write", "users:delete"]
}
```

#### 基于资源的访问控制

验证用户是否有权访问特定资源。

**规则：**
- 用户只能访问自己的资源
- 管理员可以访问所有资源
- 特定角色可以访问特定类型的资源

### 授权检查

#### 1. 角色权限检查

验证用户角色是否满足端点要求。

**示例：删除用户（需要admin角色）**

**请求：**
```http
DELETE /api/v1/users/123 HTTP/1.1
Authorization: Bearer <token_with_role_user>
```

**响应（权限不足）：**
```http
HTTP/1.1 403 Forbidden
Content-Type: application/json

{
  "error": {
    "code": "FORBIDDEN",
    "message": "您没有权限执行此操作",
    "requiredRole": "admin",
    "currentRole": "user"
  }
}
```

**请求（admin角色）：**
```http
DELETE /api/v1/users/123 HTTP/1.1
Authorization: Bearer <token_with_role_admin>
```

**响应（成功）：**
```http
HTTP/1.1 204 No Content
```

#### 2. 资源所有权验证

验证用户是否为资源所有者。

**示例：更新用户信息**

**请求（更新自己的信息）：**
```http
PATCH /api/v1/users/123 HTTP/1.1
Authorization: Bearer <token_user_id_123>
Content-Type: application/json

{
  "name": "新名字"
}
```

**响应（成功）：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": 123,
  "name": "新名字"
}
```

**请求（更新他人信息）：**
```http
PATCH /api/v1/users/456 HTTP/1.1
Authorization: Bearer <token_user_id_123>
Content-Type: application/json

{
  "name": "新名字"
}
```

**响应（权限不足）：**
```http
HTTP/1.1 403 Forbidden
Content-Type: application/json

{
  "error": {
    "code": "FORBIDDEN",
    "message": "您只能修改自己的信息"
  }
}
```

#### 3. 细粒度权限检查

基于具体操作的权限验证。

**权限格式：**
```
resource:action
```

**示例：**
- `users:read` - 读取用户
- `users:write` - 创建/更新用户
- `users:delete` - 删除用户
- `orders:read` - 读取订单
- `orders:write` - 创建/更新订单

**JWT Payload：**
```json
{
  "sub": "123",
  "role": "moderator",
  "permissions": [
    "users:read",
    "users:write",
    "posts:read",
    "posts:write",
    "posts:delete"
  ]
}
```

**请求：**
```http
DELETE /api/v1/users/456 HTTP/1.1
Authorization: Bearer <token_with_moderator_role>
```

**响应（无删除用户权限）：**
```http
HTTP/1.1 403 Forbidden
Content-Type: application/json

{
  "error": {
    "code": "FORBIDDEN",
    "message": "您没有权限删除用户",
    "requiredPermission": "users:delete",
    "currentPermissions": ["users:read", "users:write", "posts:read", "posts:write", "posts:delete"]
  }
}
```

### 授权策略

#### 端点级授权

在路由层面定义所需权限。

**示例（伪代码）：**
```javascript
// 需要admin角色
router.delete('/api/v1/users/:id', requireRole('admin'), deleteUser);

// 需要特定权限
router.post('/api/v1/posts', requirePermission('posts:write'), createPost);

// 需要认证但无特定角色要求
router.get('/api/v1/users/me', requireAuth, getCurrentUser);

// 公开端点，无需认证
router.get('/api/v1/posts', getPosts);
```

#### 资源级授权

在业务逻辑中验证资源访问权限。

**示例（伪代码）：**
```javascript
async function updateOrder(orderId, userId, data) {
  const order = await Order.findById(orderId);

  if (!order) {
    throw new NotFoundError('订单不存在');
  }

  // 验证所有权
  if (order.userId !== userId && !isAdmin(userId)) {
    throw new ForbiddenError('您只能修改自己的订单');
  }

  // 验证状态
  if (order.status === 'completed') {
    throw new ForbiddenError('已完成的订单无法修改');
  }

  return await order.update(data);
}
```

#### 字段级授权

控制用户可以访问或修改的字段。

**示例：**

**普通用户更新自己的信息：**
```http
PATCH /api/v1/users/123 HTTP/1.1
Authorization: Bearer <token_user_id_123>
Content-Type: application/json

{
  "name": "新名字",
  "role": "admin"  // 尝试提升权限
}
```

**响应：**
```http
HTTP/1.1 403 Forbidden
Content-Type: application/json

{
  "error": {
    "code": "FORBIDDEN",
    "message": "您没有权限修改角色字段",
    "details": [
      {
        "field": "role",
        "message": "只有管理员可以修改角色"
      }
    ]
  }
}
```

**管理员更新用户信息：**
```http
PATCH /api/v1/users/123 HTTP/1.1
Authorization: Bearer <token_admin>
Content-Type: application/json

{
  "name": "新名字",
  "role": "moderator"
}
```

**响应（成功）：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": 123,
  "name": "新名字",
  "role": "moderator"
}
```

### 常见授权场景

#### 1. 公开资源

无需认证即可访问。

```http
GET /api/v1/posts HTTP/1.1
```

**响应：**
```http
HTTP/1.1 200 OK
```

#### 2. 需要认证

需要有效的Access Token。

```http
GET /api/v1/users/me HTTP/1.1
Authorization: Bearer <valid_token>
```

**响应：**
```http
HTTP/1.1 200 OK
```

#### 3. 需要特定角色

需要特定角色才能访问。

```http
GET /api/v1/admin/dashboard HTTP/1.1
Authorization: Bearer <token_with_admin_role>
```

**响应：**
```http
HTTP/1.1 200 OK
```

#### 4. 资源所有者或管理员

资源所有者或管理员可以访问。

```http
GET /api/v1/users/123/orders HTTP/1.1
Authorization: Bearer <token_user_id_123_or_admin>
```

**响应：**
```http
HTTP/1.1 200 OK
```

#### 5. 组织/团队权限

用户必须属于特定组织或团队。

```http
GET /api/v1/organizations/1/projects HTTP/1.1
Authorization: Bearer <token_member_of_org_1>
```

**响应：**
```http
HTTP/1.1 200 OK
```

### 授权错误响应

#### 401 vs 403

**401 Unauthorized：** 未认证或认证失败
```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer realm="API"

{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "请先登录"
  }
}
```

**403 Forbidden：** 已认证但无权限
```http
HTTP/1.1 403 Forbidden

{
  "error": {
    "code": "FORBIDDEN",
    "message": "您没有权限访问此资源"
  }
}
```

#### 详细错误信息

提供足够的信息帮助客户端理解权限问题，但不暴露敏感信息。

**✓ 好的错误消息：**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "您没有权限删除此文章",
    "reason": "只有文章作者或管理员可以删除文章"
  }
}
```

**✗ 不好的错误消息：**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Access denied"
  }
}
```

**✗ 暴露过多信息：**
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "User ID 123 does not have admin role in database table users"
  }
}
```

### 授权最佳实践

#### 1. 默认拒绝

采用白名单策略，默认拒绝所有访问，明确授予权限。

```javascript
// ✓ 正确：明确要求权限
router.post('/api/v1/posts', requireAuth, requirePermission('posts:write'), createPost);

// ✗ 错误：默认允许
router.post('/api/v1/posts', createPost);
```

#### 2. 最小权限原则

只授予完成任务所需的最小权限。

```javascript
// ✓ 正确：细粒度权限
const moderatorPermissions = [
  'posts:read',
  'posts:write',
  'posts:delete',
  'comments:read',
  'comments:delete'
];

// ✗ 错误：过度授权
const moderatorPermissions = ['*:*']; // 所有权限
```

#### 3. 分离认证和授权

认证验证身份，授权验证权限，两者分离。

```javascript
// 认证中间件
function requireAuth(req, res, next) {
  const token = extractToken(req);
  const user = verifyToken(token);
  req.user = user;
  next();
}

// 授权中间件
function requireRole(role) {
  return (req, res, next) => {
    if (req.user.role !== role) {
      return res.status(403).json({ error: 'Forbidden' });
    }
    next();
  };
}

// 使用
router.delete('/api/v1/users/:id', requireAuth, requireRole('admin'), deleteUser);
```

#### 4. 服务端验证

永远在服务端验证权限，不要依赖客户端。

```javascript
// ✗ 错误：只在客户端检查
if (user.role === 'admin') {
  showDeleteButton();
}

// ✓ 正确：服务端验证
router.delete('/api/v1/users/:id', requireAuth, requireRole('admin'), deleteUser);
```

#### 5. 审计日志

记录所有授权决策，便于审计和调试。

```javascript
function logAuthorizationDecision(userId, resource, action, granted) {
  logger.info({
    userId,
    resource,
    action,
    granted,
    timestamp: new Date()
  });
}

// 使用
if (!hasPermission(user, 'users:delete')) {
  logAuthorizationDecision(user.id, 'users', 'delete', false);
  throw new ForbiddenError();
}
logAuthorizationDecision(user.id, 'users', 'delete', true);
```

#### 6. 一致的权限检查

在所有相关端点使用一致的权限检查逻辑。

```javascript
// 定义权限检查函数
function canDeleteUser(currentUser, targetUserId) {
  return currentUser.role === 'admin' || currentUser.id === targetUserId;
}

// 在多个端点使用
router.delete('/api/v1/users/:id', requireAuth, async (req, res) => {
  if (!canDeleteUser(req.user, req.params.id)) {
    return res.status(403).json({ error: 'Forbidden' });
  }
  // 删除逻辑...
});

router.patch('/api/v1/users/:id/deactivate', requireAuth, async (req, res) => {
  if (!canDeleteUser(req.user, req.params.id)) {
    return res.status(403).json({ error: 'Forbidden' });
  }
  // 停用逻辑...
});
```

#### 7. 资源不存在 vs 无权限

当资源不存在时返回404，而不是403，避免信息泄露。

```javascript
// ✓ 正确
async function getOrder(orderId, userId) {
  const order = await Order.findById(orderId);

  if (!order) {
    throw new NotFoundError('订单不存在'); // 404
  }

  if (order.userId !== userId && !isAdmin(userId)) {
    throw new ForbiddenError('无权访问此订单'); // 403
  }

  return order;
}

// ✗ 错误：暴露资源存在性
async function getOrder(orderId, userId) {
  const order = await Order.findById(orderId);

  if (order && order.userId !== userId && !isAdmin(userId)) {
    throw new ForbiddenError('无权访问此订单'); // 403，暴露了订单存在
  }

  if (!order) {
    throw new NotFoundError('订单不存在'); // 404
  }

  return order;
}
```

### 授权实现示例

#### 中间件实现

```javascript
// 角色检查中间件
function requireRole(...roles) {
  return (req, res, next) => {
    if (!req.user) {
      return res.status(401).json({
        error: {
          code: 'UNAUTHORIZED',
          message: '请先登录'
        }
      });
    }

    if (!roles.includes(req.user.role)) {
      return res.status(403).json({
        error: {
          code: 'FORBIDDEN',
          message: '您没有权限执行此操作',
          requiredRole: roles,
          currentRole: req.user.role
        }
      });
    }

    next();
  };
}

// 权限检查中间件
function requirePermission(...permissions) {
  return (req, res, next) => {
    if (!req.user) {
      return res.status(401).json({
        error: {
          code: 'UNAUTHORIZED',
          message: '请先登录'
        }
      });
    }

    const hasPermission = permissions.some(p =>
      req.user.permissions.includes(p)
    );

    if (!hasPermission) {
      return res.status(403).json({
        error: {
          code: 'FORBIDDEN',
          message: '您没有权限执行此操作',
          requiredPermission: permissions,
          currentPermissions: req.user.permissions
        }
      });
    }

    next();
  };
}

// 资源所有权检查
function requireOwnership(getResourceUserId) {
  return async (req, res, next) => {
    const resourceUserId = await getResourceUserId(req);

    if (req.user.id !== resourceUserId && req.user.role !== 'admin') {
      return res.status(403).json({
        error: {
          code: 'FORBIDDEN',
          message: '您只能访问自己的资源'
        }
      });
    }

    next();
  };
}

// 使用示例
router.delete('/api/v1/users/:id', requireAuth, requireRole('admin'), deleteUser);

router.post('/api/v1/posts', requireAuth, requirePermission('posts:write'), createPost);

router.patch('/api/v1/orders/:id', requireAuth, requireOwnership(async (req) => {
  const order = await Order.findById(req.params.id);
  return order?.userId;
}), updateOrder);
```

## 请求限流

### 基本原则

API必须实现请求限流（Rate Limiting），防止滥用、保护服务稳定性，确保公平的资源分配。

### 限流策略

#### 固定窗口限流

在固定时间窗口内限制请求数量。

**示例：** 每分钟最多100个请求

```
时间窗口：00:00-00:59
允许请求：100次
窗口重置：01:00
```

**优点：** 实现简单
**缺点：** 窗口边界可能出现流量突刺

#### 滑动窗口限流（推荐）

使用滑动时间窗口，更平滑地限制请求。

**示例：** 过去60秒内最多100个请求

**优点：** 更精确，避免突刺
**缺点：** 实现稍复杂

#### 令牌桶算法

以固定速率生成令牌，请求消耗令牌。

**优点：** 允许短时突发流量
**缺点：** 实现复杂

### 限流响应头

API必须在响应中包含限流信息。

#### 标准响应头

| 响应头 | 说明 | 示例 |
|--------|------|------|
| X-RateLimit-Limit | 时间窗口内的请求限制 | 100 |
| X-RateLimit-Remaining | 剩余可用请求数 | 95 |
| X-RateLimit-Reset | 限制重置的时间戳（Unix时间） | 1710842460 |

#### 示例

**正常请求：**
```http
GET /api/v1/users HTTP/1.1
Authorization: Bearer token
```

**响应：**
```http
HTTP/1.1 200 OK
Content-Type: application/json
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1710842460

{
  "data": [...]
}
```

**超出限制：**
```http
GET /api/v1/users HTTP/1.1
Authorization: Bearer token
```

**响应：**
```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 60
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1710842460

{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "请求过于频繁，请60秒后再试",
    "retryAfter": 60
  }
}
```

### 限流维度

#### 1. 基于用户

每个用户独立计算限流。

**标识：** 用户ID（从Access Token中提取）

**示例：**
```
用户123：100次/分钟
用户456：100次/分钟
```

#### 2. 基于IP地址

针对未认证请求，基于IP地址限流。

**标识：** 客户端IP地址

**示例：**
```
IP 192.168.1.1：20次/分钟（未认证）
IP 192.168.1.2：20次/分钟（未认证）
```

#### 3. 基于API Key

针对服务端集成，基于API Key限流。

**标识：** API Key

**示例：**
```
API Key abc123：1000次/小时
API Key def456：1000次/小时
```

#### 4. 基于端点

不同端点使用不同的限流策略。

**示例：**
```
GET /api/v1/users：100次/分钟
POST /api/v1/users：10次/分钟（创建操作更严格）
POST /api/v1/auth/login：5次/分钟（防止暴力破解）
```

### 差异化限流策略

根据用户类型或订阅级别设置不同的限流。

#### 示例

| 用户类型 | 限流 |
|---------|------|
| 免费用户 | 100次/小时 |
| 基础订阅 | 1000次/小时 |
| 高级订阅 | 10000次/小时 |
| 企业订阅 | 无限制或100000次/小时 |

**响应头示例（免费用户）：**
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 85
X-RateLimit-Reset: 1710846000
X-RateLimit-Tier: free
```

**响应头示例（高级用户）：**
```http
X-RateLimit-Limit: 10000
X-RateLimit-Remaining: 9850
X-RateLimit-Reset: 1710846000
X-RateLimit-Tier: premium
```

### 限流实现

#### Redis实现（推荐）

使用Redis的原子操作实现限流。

**固定窗口实现：**
```javascript
async function checkRateLimit(userId, limit, windowSeconds) {
  const key = `rate_limit:${userId}:${Math.floor(Date.now() / 1000 / windowSeconds)}`;

  const current = await redis.incr(key);

  if (current === 1) {
    await redis.expire(key, windowSeconds);
  }

  const ttl = await redis.ttl(key);
  const resetTime = Math.floor(Date.now() / 1000) + ttl;

  return {
    allowed: current <= limit,
    limit: limit,
    remaining: Math.max(0, limit - current),
    reset: resetTime
  };
}

// 使用
const result = await checkRateLimit('user123', 100, 60);

if (!result.allowed) {
  return res.status(429).json({
    error: {
      code: 'RATE_LIMIT_EXCEEDED',
      message: '请求过于频繁'
    }
  });
}

res.set('X-RateLimit-Limit', result.limit);
res.set('X-RateLimit-Remaining', result.remaining);
res.set('X-RateLimit-Reset', result.reset);
```

**滑动窗口实现：**
```javascript
async function checkRateLimitSlidingWindow(userId, limit, windowSeconds) {
  const key = `rate_limit:${userId}`;
  const now = Date.now();
  const windowStart = now - windowSeconds * 1000;

  // 使用Redis Sorted Set
  const pipeline = redis.pipeline();

  // 移除过期的请求记录
  pipeline.zremrangebyscore(key, 0, windowStart);

  // 添加当前请求
  pipeline.zadd(key, now, `${now}-${Math.random()}`);

  // 计数
  pipeline.zcard(key);

  // 设置过期时间
  pipeline.expire(key, windowSeconds);

  const results = await pipeline.exec();
  const count = results[2][1];

  const resetTime = Math.floor((now + windowSeconds * 1000) / 1000);

  return {
    allowed: count <= limit,
    limit: limit,
    remaining: Math.max(0, limit - count),
    reset: resetTime
  };
}
```

#### 中间件实现

```javascript
function rateLimitMiddleware(options = {}) {
  const {
    limit = 100,
    windowSeconds = 60,
    keyGenerator = (req) => req.user?.id || req.ip
  } = options;

  return async (req, res, next) => {
    const key = keyGenerator(req);

    if (!key) {
      return next();
    }

    const result = await checkRateLimit(key, limit, windowSeconds);

    res.set('X-RateLimit-Limit', result.limit);
    res.set('X-RateLimit-Remaining', result.remaining);
    res.set('X-RateLimit-Reset', result.reset);

    if (!result.allowed) {
      const retryAfter = result.reset - Math.floor(Date.now() / 1000);
      res.set('Retry-After', retryAfter);

      return res.status(429).json({
        error: {
          code: 'RATE_LIMIT_EXCEEDED',
          message: `请求过于频繁，请${retryAfter}秒后再试`,
          retryAfter: retryAfter
        }
      });
    }

    next();
  };
}

// 使用
app.use('/api/v1', rateLimitMiddleware({
  limit: 100,
  windowSeconds: 60
}));

// 特定端点使用更严格的限流
app.post('/api/v1/auth/login', rateLimitMiddleware({
  limit: 5,
  windowSeconds: 60
}), loginHandler);
```

### 限流绕过

某些情况下可以绕过限流。

#### 白名单

特定用户或IP地址不受限流限制。

```javascript
const WHITELIST = ['admin_user_id', '192.168.1.100'];

function rateLimitMiddleware(options = {}) {
  return async (req, res, next) => {
    const key = keyGenerator(req);

    if (WHITELIST.includes(key)) {
      return next(); // 绕过限流
    }

    // 正常限流逻辑...
  };
}
```

#### 内部服务

来自内部服务的请求不受限流限制。

```javascript
function rateLimitMiddleware(options = {}) {
  return async (req, res, next) => {
    // 检查是否为内部服务请求
    if (req.headers['x-internal-service'] === process.env.INTERNAL_SECRET) {
      return next();
    }

    // 正常限流逻辑...
  };
}
```

### 限流最佳实践

#### 1. 合理的限流值

根据实际负载能力设置限流值。

**考虑因素：**
- 服务器处理能力
- 数据库连接数
- 第三方API限制
- 用户体验

**示例：**
```
读操作：100次/分钟
写操作：10次/分钟
搜索操作：20次/分钟
登录操作：5次/分钟
```

#### 2. 清晰的错误消息

提供明确的错误消息和重试时间。

**✓ 好的错误消息：**
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "您已超出请求限制（100次/分钟），请60秒后再试",
    "retryAfter": 60,
    "limit": 100,
    "resetAt": "2024-03-19T10:01:00Z"
  }
}
```

**✗ 不好的错误消息：**
```json
{
  "error": "Too many requests"
}
```

#### 3. 渐进式限流

对于重复违规的用户，逐步增加限制。

```javascript
async function getLimit(userId) {
  const violations = await redis.get(`violations:${userId}`);

  if (violations > 10) {
    return 10; // 严重违规，降低到10次/分钟
  } else if (violations > 5) {
    return 50; // 多次违规，降低到50次/分钟
  }

  return 100; // 正常限制
}
```

#### 4. 监控和告警

监控限流触发情况，识别异常流量。

```javascript
async function checkRateLimit(userId, limit, windowSeconds) {
  const result = await checkRateLimitImpl(userId, limit, windowSeconds);

  if (!result.allowed) {
    // 记录限流事件
    logger.warn({
      event: 'rate_limit_exceeded',
      userId,
      limit,
      timestamp: new Date()
    });

    // 增加违规计数
    await redis.incr(`violations:${userId}`);
    await redis.expire(`violations:${userId}`, 3600);
  }

  return result;
}
```

#### 5. 文档说明

在API文档中明确说明限流策略。

**示例文档：**
```markdown
## 限流策略

所有API端点都受到限流限制，以确保服务稳定性。

### 限流规则

- 认证用户：100次/分钟
- 未认证用户：20次/分钟
- 登录端点：5次/分钟

### 响应头

- X-RateLimit-Limit：时间窗口内的请求限制
- X-RateLimit-Remaining：剩余可用请求数
- X-RateLimit-Reset：限制重置的时间戳

### 超出限制

当超出限流时，API返回429状态码和Retry-After响应头。

### 提升限制

如需更高的限流限制，请升级到高级订阅或联系我们。
```

#### 6. 客户端重试策略

客户端应实现指数退避重试。

**JavaScript示例：**
```javascript
async function fetchWithRetry(url, options = {}, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    const response = await fetch(url, options);

    if (response.status !== 429) {
      return response;
    }

    // 获取重试时间
    const retryAfter = parseInt(response.headers.get('Retry-After') || '60');

    if (i < maxRetries - 1) {
      // 等待后重试
      await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));
    } else {
      // 最后一次重试失败
      throw new Error('Rate limit exceeded');
    }
  }
}
```

#### 7. 成本优化

对于高频端点，考虑使用缓存减少后端负载。

```javascript
// 缓存GET请求结果
app.get('/api/v1/posts', cacheMiddleware(60), rateLimitMiddleware(), getPosts);
```

### 限流与其他机制结合

#### 限流 + 认证

未认证用户使用更严格的限流。

```javascript
function rateLimitMiddleware() {
  return async (req, res, next) => {
    const limit = req.user ? 100 : 20; // 认证用户100次，未认证20次
    const key = req.user?.id || req.ip;

    const result = await checkRateLimit(key, limit, 60);
    // 处理限流...
  };
}
```

#### 限流 + 订阅级别

根据订阅级别设置不同限流。

```javascript
function getRateLimitByTier(tier) {
  const limits = {
    free: 100,
    basic: 1000,
    premium: 10000,
    enterprise: 100000
  };
  return limits[tier] || limits.free;
}

function rateLimitMiddleware() {
  return async (req, res, next) => {
    const tier = req.user?.subscriptionTier || 'free';
    const limit = getRateLimitByTier(tier);

    const result = await checkRateLimit(req.user.id, limit, 3600);
    // 处理限流...
  };
}
```

### 限流测试

#### 单元测试

```javascript
describe('Rate Limiting', () => {
  it('should allow requests within limit', async () => {
    for (let i = 0; i < 100; i++) {
      const result = await checkRateLimit('user123', 100, 60);
      expect(result.allowed).toBe(true);
    }
  });

  it('should block requests exceeding limit', async () => {
    for (let i = 0; i < 100; i++) {
      await checkRateLimit('user456', 100, 60);
    }

    const result = await checkRateLimit('user456', 100, 60);
    expect(result.allowed).toBe(false);
  });

  it('should reset after window expires', async () => {
    // 达到限制
    for (let i = 0; i < 100; i++) {
      await checkRateLimit('user789', 100, 1);
    }

    // 等待窗口过期
    await new Promise(resolve => setTimeout(resolve, 1100));

    // 应该可以再次请求
    const result = await checkRateLimit('user789', 100, 1);
    expect(result.allowed).toBe(true);
  });
});
```

## 幂等性保证

### 基本原则

API必须确保幂等操作的幂等性，即多次执行相同操作产生相同结果，避免重复操作导致的问题。

### 幂等性定义

**幂等操作：** 多次执行与执行一次产生相同的结果和副作用。

**示例：**
- 设置用户名为"张三"：无论执行多少次，结果都是用户名为"张三"
- 删除ID为123的用户：无论执行多少次，结果都是该用户被删除

**非幂等操作：**
- 创建用户：每次执行都创建一个新用户
- 增加余额：每次执行都增加一次

### HTTP方法的幂等性

| 方法 | 幂等性 | 说明 |
|------|--------|------|
| GET | ✓ | 只读操作，天然幂等 |
| HEAD | ✓ | 只读操作，天然幂等 |
| OPTIONS | ✓ | 只读操作，天然幂等 |
| PUT | ✓ | 完整替换资源，幂等 |
| DELETE | ✓ | 删除资源，幂等 |
| PATCH | ✓ | 部分更新，应设计为幂等 |
| POST | ✗ | 创建资源，非幂等 |

### PUT操作的幂等性

PUT操作必须是幂等的，多次执行相同的PUT请求，资源状态保持一致。

#### 示例

**第一次请求：**
```http
PUT /api/v1/users/123 HTTP/1.1
Content-Type: application/json

{
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000"
}
```

**响应：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": 123,
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "updatedAt": "2024-03-19T10:00:00Z"
}
```

**第二次相同请求：**
```http
PUT /api/v1/users/123 HTTP/1.1
Content-Type: application/json

{
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000"
}
```

**响应（资源状态相同）：**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "id": 123,
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "updatedAt": "2024-03-19T10:00:00Z"
}
```

**说明：** updatedAt可能不同，但资源的业务状态相同。

### DELETE操作的幂等性

DELETE操作必须是幂等的，多次删除同一资源，结果一致。

#### 示例

**第一次删除：**
```http
DELETE /api/v1/users/123 HTTP/1.1
```

**响应：**
```http
HTTP/1.1 204 No Content
```

**第二次删除（资源已不存在）：**
```http
DELETE /api/v1/users/123 HTTP/1.1
```

**响应：**
```http
HTTP/1.1 404 Not Found
Content-Type: application/json

{
  "error": {
    "code": "NOT_FOUND",
    "message": "用户不存在"
  }
}
```

**说明：** 第二次返回404，但这是幂等的——资源确实不存在。

**替代方案（返回204）：**
某些API设计选择第二次删除也返回204，认为"确保资源不存在"是幂等的。

```http
HTTP/1.1 204 No Content
```

**推荐：** 返回404更准确，但两种方式都可接受。

### PATCH操作的幂等性

PATCH操作应设计为幂等的，但需要注意操作语义。

#### 幂等的PATCH

**设置字段值（幂等）：**
```http
PATCH /api/v1/users/123 HTTP/1.1
Content-Type: application/json

{
  "phone": "13900139000"
}
```

多次执行，phone始终为"13900139000"。

#### 非幂等的PATCH（应避免）

**增量操作（非幂等）：**
```http
PATCH /api/v1/users/123 HTTP/1.1
Content-Type: application/json

{
  "balance": { "$increment": 100 }
}
```

每次执行都增加100，非幂等。

**推荐替代方案：**
使用专门的端点处理增量操作，并使用幂等性令牌。

```http
POST /api/v1/users/123/transactions HTTP/1.1
Content-Type: application/json
Idempotency-Key: unique-key-123

{
  "amount": 100,
  "type": "deposit"
}
```

### POST操作的幂等性

POST操作天然非幂等，但可以通过幂等性令牌实现幂等。

#### 幂等性令牌（Idempotency Key）

客户端在请求头中提供唯一的幂等性令牌，服务端保证相同令牌的请求只执行一次。

##### 请求格式

```http
POST /api/v1/orders HTTP/1.1
Authorization: Bearer token
Content-Type: application/json
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000

{
  "productId": 123,
  "quantity": 2
}
```

##### 首次请求

**响应：**
```http
HTTP/1.1 201 Created
Location: /api/v1/orders/456
Content-Type: application/json

{
  "id": 456,
  "productId": 123,
  "quantity": 2,
  "status": "pending",
  "createdAt": "2024-03-19T10:00:00Z"
}
```

##### 重复请求（相同Idempotency-Key）

**响应（返回相同结果）：**
```http
HTTP/1.1 201 Created
Location: /api/v1/orders/456
Content-Type: application/json

{
  "id": 456,
  "productId": 123,
  "quantity": 2,
  "status": "pending",
  "createdAt": "2024-03-19T10:00:00Z"
}
```

**说明：** 服务端识别出相同的Idempotency-Key，返回首次请求的结果，不创建新订单。

##### 不同请求体（相同Idempotency-Key）

如果使用相同的Idempotency-Key但请求体不同，返回错误。

**请求：**
```http
POST /api/v1/orders HTTP/1.1
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000

{
  "productId": 456,
  "quantity": 3
}
```

**响应：**
```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json

{
  "error": {
    "code": "IDEMPOTENCY_KEY_MISMATCH",
    "message": "该幂等性令牌已被使用，但请求内容不同"
  }
}
```

### 幂等性令牌实现

#### 生成幂等性令牌

客户端应生成唯一的幂等性令牌。

**推荐格式：** UUID v4

**JavaScript示例：**
```javascript
function generateIdempotencyKey() {
  return crypto.randomUUID();
}

// 使用
const idempotencyKey = generateIdempotencyKey();
// "550e8400-e29b-41d4-a716-446655440000"

await fetch('/api/v1/orders', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Idempotency-Key': idempotencyKey
  },
  body: JSON.stringify({ productId: 123, quantity: 2 })
});
```

**Python示例：**
```python
import uuid

def generate_idempotency_key():
    return str(uuid.uuid4())

# 使用
idempotency_key = generate_idempotency_key()

response = requests.post(
    '/api/v1/orders',
    headers={
        'Content-Type': 'application/json',
        'Idempotency-Key': idempotency_key
    },
    json={'productId': 123, 'quantity': 2}
)
```

#### 服务端实现

使用Redis存储幂等性令牌和响应结果。

**实现示例：**
```javascript
async function handleIdempotentRequest(idempotencyKey, requestBody, handler) {
  if (!idempotencyKey) {
    // 没有幂等性令牌，正常处理
    return await handler();
  }

  const cacheKey = `idempotency:${idempotencyKey}`;

  // 检查是否已处理
  const cached = await redis.get(cacheKey);

  if (cached) {
    const cachedData = JSON.parse(cached);

    // 验证请求体是否相同
    if (cachedData.requestHash !== hashRequest(requestBody)) {
      throw new Error('Idempotency key mismatch');
    }

    // 返回缓存的响应
    return cachedData.response;
  }

  // 首次请求，执行处理
  const response = await handler();

  // 缓存结果（24小时）
  await redis.setex(
    cacheKey,
    86400,
    JSON.stringify({
      requestHash: hashRequest(requestBody),
      response: response
    })
  );

  return response;
}

function hashRequest(body) {
  return crypto.createHash('sha256').update(JSON.stringify(body)).digest('hex');
}

// 使用
app.post('/api/v1/orders', async (req, res) => {
  const idempotencyKey = req.headers['idempotency-key'];

  try {
    const result = await handleIdempotentRequest(
      idempotencyKey,
      req.body,
      async () => {
        // 实际的业务逻辑
        return await createOrder(req.body);
      }
    );

    res.status(201).json(result);
  } catch (error) {
    if (error.message === 'Idempotency key mismatch') {
      return res.status(422).json({
        error: {
          code: 'IDEMPOTENCY_KEY_MISMATCH',
          message: '该幂等性令牌已被使用，但请求内容不同'
        }
      });
    }
    throw error;
  }
});
```

### 幂等性令牌的生命周期

#### 过期时间

幂等性令牌应有合理的过期时间。

**推荐：** 24小时

**原因：**
- 足够长，覆盖大多数重试场景
- 不会无限占用存储空间

#### 清理策略

使用Redis的TTL自动清理过期令牌。

```javascript
await redis.setex(cacheKey, 86400, data); // 24小时后自动删除
```

### 幂等性最佳实践

#### 1. 明确幂等性要求

在API文档中明确说明哪些端点支持幂等性令牌。

**示例文档：**
```markdown
## POST /api/v1/orders

创建新订单。

### 幂等性

此端点支持幂等性令牌。在Idempotency-Key请求头中提供UUID，
确保重复请求不会创建多个订单。

### 请求头

- Idempotency-Key (可选): UUID格式的幂等性令牌

### 示例

```http
POST /api/v1/orders
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
```
```

#### 2. 幂等性令牌为可选

不强制要求幂等性令牌，但推荐使用。

```javascript
// 没有幂等性令牌时正常处理
if (!idempotencyKey) {
  return await handler();
}
```

#### 3. 适用场景

幂等性令牌特别适用于：
- 支付操作
- 订单创建
- 资金转账
- 重要的状态变更

#### 4. 客户端重试

客户端在网络错误时应使用相同的幂等性令牌重试。

**JavaScript示例：**
```javascript
async function createOrderWithRetry(orderData, maxRetries = 3) {
  const idempotencyKey = generateIdempotencyKey();

  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch('/api/v1/orders', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Idempotency-Key': idempotencyKey // 使用相同的令牌
        },
        body: JSON.stringify(orderData)
      });

      if (response.ok) {
        return await response.json();
      }

      if (response.status === 422) {
        // 幂等性令牌冲突，不重试
        throw new Error('Idempotency key mismatch');
      }

      // 其他错误，重试
    } catch (error) {
      if (i === maxRetries - 1) {
        throw error;
      }
      // 等待后重试
      await new Promise(resolve => setTimeout(resolve, 1000 * (i + 1)));
    }
  }
}
```

#### 5. 监控和日志

记录幂等性令牌的使用情况。

```javascript
async function handleIdempotentRequest(idempotencyKey, requestBody, handler) {
  if (!idempotencyKey) {
    return await handler();
  }

  const cached = await redis.get(`idempotency:${idempotencyKey}`);

  if (cached) {
    logger.info({
      event: 'idempotency_key_reused',
      key: idempotencyKey,
      timestamp: new Date()
    });
    return JSON.parse(cached).response;
  }

  logger.info({
    event: 'idempotency_key_first_use',
    key: idempotencyKey,
    timestamp: new Date()
  });

  const response = await handler();
  // 缓存结果...
  return response;
}
```

#### 6. 安全考虑

防止幂等性令牌被滥用。

**限制：**
- 幂等性令牌应与用户绑定
- 不同用户不能使用相同的令牌

**实现：**
```javascript
const cacheKey = `idempotency:${userId}:${idempotencyKey}`;
```

#### 7. 测试幂等性

编写测试验证幂等性。

**测试示例：**
```javascript
describe('Order Creation Idempotency', () => {
  it('should create order only once with same idempotency key', async () => {
    const idempotencyKey = generateIdempotencyKey();
    const orderData = { productId: 123, quantity: 2 };

    // 第一次请求
    const response1 = await createOrder(orderData, idempotencyKey);
    expect(response1.status).toBe(201);
    const order1 = await response1.json();

    // 第二次请求（相同令牌）
    const response2 = await createOrder(orderData, idempotencyKey);
    expect(response2.status).toBe(201);
    const order2 = await response2.json();

    // 应返回相同的订单
    expect(order1.id).toBe(order2.id);

    // 数据库中只有一个订单
    const ordersCount = await Order.count({ id: order1.id });
    expect(ordersCount).toBe(1);
  });

  it('should reject different request with same idempotency key', async () => {
    const idempotencyKey = generateIdempotencyKey();

    await createOrder({ productId: 123, quantity: 2 }, idempotencyKey);

    const response = await createOrder(
      { productId: 456, quantity: 3 },
      idempotencyKey
    );

    expect(response.status).toBe(422);
  });
});
```

### 幂等性总结

| 操作 | 幂等性 | 实现方式 |
|------|--------|---------|
| GET | 天然幂等 | 无需特殊处理 |
| PUT | 必须幂等 | 完整替换资源 |
| DELETE | 必须幂等 | 删除后返回404或204 |
| PATCH | 应该幂等 | 避免增量操作 |
| POST | 可选幂等 | 使用Idempotency-Key |

---

## 代码示例

### 完整的Express.js API示例

以下是一个集成所有实现规范的完整示例。

#### 项目结构

```
src/
├── middleware/
│   ├── auth.js           # 认证中间件
│   ├── rateLimit.js      # 限流中间件
│   └── errorHandler.js   # 错误处理中间件
├── routes/
│   └── users.js          # 用户路由
├── controllers/
│   └── userController.js # 用户控制器
└── app.js                # 应用入口
```

#### 1. 认证中间件

```javascript
// middleware/auth.js
const jwt = require('jsonwebtoken');

function requireAuth(req, res, next) {
  const authHeader = req.headers.authorization;

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({
      error: {
        code: 'UNAUTHORIZED',
        message: '未提供认证令牌'
      }
    });
  }

  const token = authHeader.substring(7);

  try {
    const decoded = jwt.verify(token, process.env.JWT_SECRET);
    req.user = decoded;
    next();
  } catch (error) {
    if (error.name === 'TokenExpiredError') {
      return res.status(401).json({
        error: {
          code: 'TOKEN_EXPIRED',
          message: '访问令牌已过期，请刷新令牌'
        }
      });
    }

    return res.status(401).json({
      error: {
        code: 'INVALID_TOKEN',
        message: '认证令牌无效'
      }
    });
  }
}

function requireRole(...roles) {
  return (req, res, next) => {
    if (!req.user) {
      return res.status(401).json({
        error: {
          code: 'UNAUTHORIZED',
          message: '请先登录'
        }
      });
    }

    if (!roles.includes(req.user.role)) {
      return res.status(403).json({
        error: {
          code: 'FORBIDDEN',
          message: '您没有权限执行此操作',
          requiredRole: roles,
          currentRole: req.user.role
        }
      });
    }

    next();
  };
}

module.exports = { requireAuth, requireRole };
```

#### 2. 限流中间件

```javascript
// middleware/rateLimit.js
const redis = require('redis');
const client = redis.createClient();

async function checkRateLimit(key, limit, windowSeconds) {
  const now = Math.floor(Date.now() / 1000);
  const windowKey = `rate_limit:${key}:${Math.floor(now / windowSeconds)}`;

  const current = await client.incr(windowKey);

  if (current === 1) {
    await client.expire(windowKey, windowSeconds);
  }

  const ttl = await client.ttl(windowKey);
  const resetTime = now + ttl;

  return {
    allowed: current <= limit,
    limit: limit,
    remaining: Math.max(0, limit - current),
    reset: resetTime
  };
}

function rateLimitMiddleware(options = {}) {
  const {
    limit = 100,
    windowSeconds = 60,
    keyGenerator = (req) => req.user?.id || req.ip
  } = options;

  return async (req, res, next) => {
    const key = keyGenerator(req);

    if (!key) {
      return next();
    }

    try {
      const result = await checkRateLimit(key, limit, windowSeconds);

      res.set('X-RateLimit-Limit', result.limit);
      res.set('X-RateLimit-Remaining', result.remaining);
      res.set('X-RateLimit-Reset', result.reset);

      if (!result.allowed) {
        const retryAfter = result.reset - Math.floor(Date.now() / 1000);
        res.set('Retry-After', retryAfter);

        return res.status(429).json({
          error: {
            code: 'RATE_LIMIT_EXCEEDED',
            message: `请求过于频繁，请${retryAfter}秒后再试`,
            retryAfter: retryAfter
          }
        });
      }

      next();
    } catch (error) {
      console.error('Rate limit error:', error);
      next(); // 限流失败时放行，避免影响服务
    }
  };
}

module.exports = { rateLimitMiddleware };
```

#### 3. 错误处理中间件

```javascript
// middleware/errorHandler.js
class ApiError extends Error {
  constructor(statusCode, code, message, details = null) {
    super(message);
    this.statusCode = statusCode;
    this.code = code;
    this.details = details;
  }
}

class ValidationError extends ApiError {
  constructor(message, details) {
    super(400, 'VALIDATION_ERROR', message, details);
  }
}

class NotFoundError extends ApiError {
  constructor(message = '资源不存在') {
    super(404, 'NOT_FOUND', message);
  }
}

class ForbiddenError extends ApiError {
  constructor(message = '您没有权限访问此资源') {
    super(403, 'FORBIDDEN', message);
  }
}

function errorHandler(err, req, res, next) {
  console.error('Error:', err);

  if (err instanceof ApiError) {
    const response = {
      error: {
        code: err.code,
        message: err.message
      }
    };

    if (err.details) {
      response.error.details = err.details;
    }

    return res.status(err.statusCode).json(response);
  }

  // 未知错误
  res.status(500).json({
    error: {
      code: 'INTERNAL_ERROR',
      message: '服务器内部错误，请稍后重试'
    }
  });
}

module.exports = {
  ApiError,
  ValidationError,
  NotFoundError,
  ForbiddenError,
  errorHandler
};
```

#### 4. 用户控制器

```javascript
// controllers/userController.js
const { ValidationError, NotFoundError, ForbiddenError } = require('../middleware/errorHandler');

class UserController {
  async getUsers(req, res) {
    // 解析查询参数
    const offset = parseInt(req.query.offset) || 0;
    const limit = Math.min(parseInt(req.query.limit) || 20, 100);
    const status = req.query.status;
    const sort = req.query.sort || '-createdAt';
    const fields = req.query.fields?.split(',');

    // 构建查询
    const query = {};
    if (status) {
      query.status = status;
    }

    // 解析排序
    const sortField = sort.startsWith('-') ? sort.substring(1) : sort;
    const sortOrder = sort.startsWith('-') ? -1 : 1;

    // 查询数据库
    const users = await User.find(query)
      .sort({ [sortField]: sortOrder })
      .skip(offset)
      .limit(limit)
      .select(fields?.join(' '));

    const total = await User.countDocuments(query);

    res.json({
      data: users,
      total: total,
      offset: offset,
      limit: limit
    });
  }

  async getUser(req, res) {
    const userId = req.params.id;
    const fields = req.query.fields?.split(',');

    const user = await User.findById(userId).select(fields?.join(' '));

    if (!user) {
      throw new NotFoundError('用户不存在');
    }

    res.json(user);
  }

  async createUser(req, res) {
    // 验证请求体
    const errors = [];

    if (!req.body.name) {
      errors.push({ field: 'name', message: '姓名不能为空' });
    }

    if (!req.body.email) {
      errors.push({ field: 'email', message: '邮箱不能为空' });
    } else if (!isValidEmail(req.body.email)) {
      errors.push({ field: 'email', message: '邮箱格式不正确' });
    }

    if (errors.length > 0) {
      throw new ValidationError('请求参数验证失败', errors);
    }

    // 检查邮箱是否已存在
    const existingUser = await User.findOne({ email: req.body.email });
    if (existingUser) {
      return res.status(409).json({
        error: {
          code: 'DUPLICATE_RESOURCE',
          message: '该邮箱已被注册'
        }
      });
    }

    // 创建用户
    const user = await User.create(req.body);

    res.status(201)
      .location(`/api/v1/users/${user.id}`)
      .json(user);
  }

  async updateUser(req, res) {
    const userId = req.params.id;

    // 检查权限
    if (req.user.id !== userId && req.user.role !== 'admin') {
      throw new ForbiddenError('您只能修改自己的信息');
    }

    // 普通用户不能修改角色
    if (req.body.role && req.user.role !== 'admin') {
      throw new ForbiddenError('只有管理员可以修改角色');
    }

    const user = await User.findById(userId);

    if (!user) {
      throw new NotFoundError('用户不存在');
    }

    // 更新用户
    Object.assign(user, req.body);
    await user.save();

    res.json(user);
  }

  async deleteUser(req, res) {
    const userId = req.params.id;

    const user = await User.findById(userId);

    if (!user) {
      throw new NotFoundError('用户不存在');
    }

    await user.remove();

    res.status(204).send();
  }
}

function isValidEmail(email) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

module.exports = new UserController();
```

#### 5. 用户路由

```javascript
// routes/users.js
const express = require('express');
const router = express.Router();
const userController = require('../controllers/userController');
const { requireAuth, requireRole } = require('../middleware/auth');
const { rateLimitMiddleware } = require('../middleware/rateLimit');

// 获取用户列表（需要认证）
router.get('/',
  requireAuth,
  rateLimitMiddleware({ limit: 100, windowSeconds: 60 }),
  asyncHandler(userController.getUsers)
);

// 获取单个用户（需要认证）
router.get('/:id',
  requireAuth,
  rateLimitMiddleware({ limit: 100, windowSeconds: 60 }),
  asyncHandler(userController.getUser)
);

// 创建用户（需要admin角色）
router.post('/',
  requireAuth,
  requireRole('admin'),
  rateLimitMiddleware({ limit: 10, windowSeconds: 60 }),
  asyncHandler(userController.createUser)
);

// 更新用户（需要认证，只能更新自己或admin）
router.patch('/:id',
  requireAuth,
  rateLimitMiddleware({ limit: 20, windowSeconds: 60 }),
  asyncHandler(userController.updateUser)
);

// 删除用户（需要admin角色）
router.delete('/:id',
  requireAuth,
  requireRole('admin'),
  rateLimitMiddleware({ limit: 10, windowSeconds: 60 }),
  asyncHandler(userController.deleteUser)
);

// 异步错误处理包装器
function asyncHandler(fn) {
  return (req, res, next) => {
    Promise.resolve(fn(req, res, next)).catch(next);
  };
}

module.exports = router;
```

#### 6. 应用入口

```javascript
// app.js
const express = require('express');
const userRoutes = require('./routes/users');
const { errorHandler } = require('./middleware/errorHandler');

const app = express();

// 中间件
app.use(express.json());

// 路由
app.use('/api/v1/users', userRoutes);

// 错误处理（必须在最后）
app.use(errorHandler);

// 启动服务器
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});

module.exports = app;
```

### Python Flask示例

```python
# app.py
from flask import Flask, request, jsonify
from functools import wraps
import jwt
import redis
import time

app = Flask(__name__)
redis_client = redis.Redis(host='localhost', port=6379, decode_responses=True)

# 认证装饰器
def require_auth(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        auth_header = request.headers.get('Authorization')

        if not auth_header or not auth_header.startswith('Bearer '):
            return jsonify({
                'error': {
                    'code': 'UNAUTHORIZED',
                    'message': '未提供认证令牌'
                }
            }), 401

        token = auth_header[7:]

        try:
            decoded = jwt.decode(token, app.config['JWT_SECRET'], algorithms=['HS256'])
            request.user = decoded
            return f(*args, **kwargs)
        except jwt.ExpiredSignatureError:
            return jsonify({
                'error': {
                    'code': 'TOKEN_EXPIRED',
                    'message': '访问令牌已过期'
                }
            }), 401
        except jwt.InvalidTokenError:
            return jsonify({
                'error': {
                    'code': 'INVALID_TOKEN',
                    'message': '认证令牌无效'
                }
            }), 401

    return decorated_function

# 角色检查装饰器
def require_role(*roles):
    def decorator(f):
        @wraps(f)
        def decorated_function(*args, **kwargs):
            if not hasattr(request, 'user'):
                return jsonify({
                    'error': {
                        'code': 'UNAUTHORIZED',
                        'message': '请先登录'
                    }
                }), 401

            if request.user.get('role') not in roles:
                return jsonify({
                    'error': {
                        'code': 'FORBIDDEN',
                        'message': '您没有权限执行此操作'
                    }
                }), 403

            return f(*args, **kwargs)
        return decorated_function
    return decorator

# 限流装饰器
def rate_limit(limit=100, window_seconds=60):
    def decorator(f):
        @wraps(f)
        def decorated_function(*args, **kwargs):
            key = request.user.get('id') if hasattr(request, 'user') else request.remote_addr
            now = int(time.time())
            window_key = f"rate_limit:{key}:{now // window_seconds}"

            current = redis_client.incr(window_key)

            if current == 1:
                redis_client.expire(window_key, window_seconds)

            ttl = redis_client.ttl(window_key)
            reset_time = now + ttl

            # 设置响应头
            response = f(*args, **kwargs)
            if isinstance(response, tuple):
                data, status_code = response
                response = jsonify(data), status_code

            response.headers['X-RateLimit-Limit'] = str(limit)
            response.headers['X-RateLimit-Remaining'] = str(max(0, limit - current))
            response.headers['X-RateLimit-Reset'] = str(reset_time)

            if current > limit:
                retry_after = reset_time - now
                return jsonify({
                    'error': {
                        'code': 'RATE_LIMIT_EXCEEDED',
                        'message': f'请求过于频繁，请{retry_after}秒后再试'
                    }
                }), 429

            return response
        return decorated_function
    return decorator

# 用户路由
@app.route('/api/v1/users', methods=['GET'])
@require_auth
@rate_limit(limit=100, window_seconds=60)
def get_users():
    offset = int(request.args.get('offset', 0))
    limit = min(int(request.args.get('limit', 20)), 100)
    status = request.args.get('status')

    # 查询逻辑...
    users = []  # 从数据库获取

    return jsonify({
        'data': users,
        'total': len(users),
        'offset': offset,
        'limit': limit
    })

@app.route('/api/v1/users/<user_id>', methods=['GET'])
@require_auth
@rate_limit(limit=100, window_seconds=60)
def get_user(user_id):
    # 查询逻辑...
    user = {}  # 从数据库获取

    if not user:
        return jsonify({
            'error': {
                'code': 'NOT_FOUND',
                'message': '用户不存在'
            }
        }), 404

    return jsonify(user)

@app.route('/api/v1/users', methods=['POST'])
@require_auth
@require_role('admin')
@rate_limit(limit=10, window_seconds=60)
def create_user():
    data = request.get_json()

    # 验证逻辑...
    errors = []

    if not data.get('name'):
        errors.append({'field': 'name', 'message': '姓名不能为空'})

    if errors:
        return jsonify({
            'error': {
                'code': 'VALIDATION_ERROR',
                'message': '请求参数验证失败',
                'details': errors
            }
        }), 400

    # 创建用户...
    user = {}  # 保存到数据库

    return jsonify(user), 201

if __name__ == '__main__':
    app.run(debug=True)
```

### 客户端使用示例

#### JavaScript/TypeScript

```typescript
// api-client.ts
class ApiClient {
  private baseURL: string;
  private accessToken: string | null = null;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  setAccessToken(token: string) {
    this.accessToken = token;
  }

  async request(endpoint: string, options: RequestInit = {}) {
    const url = `${this.baseURL}${endpoint}`;

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.accessToken) {
      headers['Authorization'] = `Bearer ${this.accessToken}`;
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    // 处理限流
    if (response.status === 429) {
      const retryAfter = parseInt(response.headers.get('Retry-After') || '60');
      throw new Error(`Rate limit exceeded. Retry after ${retryAfter} seconds`);
    }

    // 处理错误
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error.message);
    }

    return response.json();
  }

  async getUsers(params: { offset?: number; limit?: number; status?: string } = {}) {
    const queryString = new URLSearchParams(params as any).toString();
    return this.request(`/api/v1/users?${queryString}`);
  }

  async getUser(id: string, fields?: string[]) {
    const queryString = fields ? `?fields=${fields.join(',')}` : '';
    return this.request(`/api/v1/users/${id}${queryString}`);
  }

  async createUser(data: any) {
    return this.request('/api/v1/users', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateUser(id: string, data: any) {
    return this.request(`/api/v1/users/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async deleteUser(id: string) {
    return this.request(`/api/v1/users/${id}`, {
      method: 'DELETE',
    });
  }
}

// 使用
const client = new ApiClient('https://api.example.com');
client.setAccessToken('your_access_token');

const users = await client.getUsers({ offset: 0, limit: 20, status: 'active' });
```
