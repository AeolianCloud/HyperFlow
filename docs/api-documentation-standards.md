# API文档规范

本文档定义API文档的标准和要求，确保API文档的完整性、准确性和易用性。

## 目录

1. [OpenAPI规范](#openapi规范)
2. [端点文档要求](#端点文档要求)
3. [交互式文档](#交互式文档)
4. [文档版本管理](#文档版本管理)
5. [代码示例](#代码示例)
6. [文档自动生成](#文档自动生成)
7. [文档可访问性](#文档可访问性)
8. [文档完整性验证](#文档完整性验证)

---

## OpenAPI规范

### 基本原则

API文档必须使用OpenAPI 3.0+规范格式，确保标准化、工具兼容性和自动化支持。

### OpenAPI 3.0文档结构

#### 基本结构

```yaml
openapi: 3.0.3
info:
  title: API标题
  version: 1.0.0
  description: API描述
  contact:
    name: API支持
    email: support@example.com
  license:
    name: MIT
    url: https://opensource.org/licenses/MIT

servers:
  - url: https://api.example.com/v1
    description: 生产环境
  - url: https://staging-api.example.com/v1
    description: 测试环境

paths:
  /users:
    get:
      summary: 获取用户列表
      # ...

components:
  schemas:
    User:
      type: object
      # ...
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - bearerAuth: []
```

### info对象

定义API的基本信息。

```yaml
info:
  title: 用户管理API
  version: 1.0.0
  description: |
    用户管理API提供用户的CRUD操作。

    ## 功能特性
    - 用户注册和登录
    - 用户信息管理
    - 角色和权限控制

  contact:
    name: API支持团队
    email: api-support@example.com
    url: https://support.example.com
  license:
    name: Apache 2.0
    url: https://www.apache.org/licenses/LICENSE-2.0.html
  termsOfService: https://example.com/terms
```

**必需字段：**
- `title`：API标题
- `version`：API版本号

**推荐字段：**
- `description`：详细描述（支持Markdown）
- `contact`：联系信息
- `license`：许可证信息

### servers对象

定义API服务器地址。

```yaml
servers:
  - url: https://api.example.com/v1
    description: 生产环境
  - url: https://staging-api.example.com/v1
    description: 测试环境
  - url: http://localhost:3000/v1
    description: 本地开发环境
```

**支持变量：**
```yaml
servers:
  - url: https://{environment}.example.com/v1
    description: 可配置环境
    variables:
      environment:
        default: api
        enum:
          - api
          - staging
          - dev
```

### paths对象

定义API端点和操作。

#### 基本路径定义

```yaml
paths:
  /users:
    get:
      summary: 获取用户列表
      description: 返回分页的用户列表
      operationId: getUsers
      tags:
        - Users
      parameters:
        - name: offset
          in: query
          description: 偏移量
          schema:
            type: integer
            default: 0
        - name: limit
          in: query
          description: 每页数量
          schema:
            type: integer
            default: 20
            maximum: 100
      responses:
        '200':
          description: 成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserListResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'
      security:
        - bearerAuth: []
```

#### 路径参数

```yaml
paths:
  /users/{userId}:
    get:
      summary: 获取单个用户
      parameters:
        - name: userId
          in: path
          required: true
          description: 用户ID
          schema:
            type: string
      responses:
        '200':
          description: 成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          $ref: '#/components/responses/NotFound'
```

#### 请求体

```yaml
paths:
  /users:
    post:
      summary: 创建用户
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
            examples:
              example1:
                summary: 基本用户
                value:
                  name: 张三
                  email: zhangsan@example.com
      responses:
        '201':
          description: 创建成功
          headers:
            Location:
              description: 新用户的URI
              schema:
                type: string
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
```

### components对象

定义可复用的组件。

#### schemas（数据模型）

```yaml
components:
  schemas:
    User:
      type: object
      required:
        - id
        - name
        - email
      properties:
        id:
          type: integer
          format: int64
          description: 用户ID
          example: 123
        name:
          type: string
          description: 用户名
          example: 张三
          minLength: 1
          maxLength: 100
        email:
          type: string
          format: email
          description: 邮箱地址
          example: zhangsan@example.com
        phone:
          type: string
          description: 电话号码
          example: "13800138000"
          nullable: true
        status:
          type: string
          enum:
            - active
            - inactive
            - suspended
          description: 用户状态
          example: active
        role:
          type: string
          enum:
            - admin
            - user
            - guest
          description: 用户角色
          example: user
        createdAt:
          type: string
          format: date-time
          description: 创建时间
          example: "2024-03-19T10:00:00Z"
        updatedAt:
          type: string
          format: date-time
          description: 更新时间
          example: "2024-03-19T10:00:00Z"

    CreateUserRequest:
      type: object
      required:
        - name
        - email
      properties:
        name:
          type: string
          minLength: 1
          maxLength: 100
        email:
          type: string
          format: email
        phone:
          type: string

    UserListResponse:
      type: object
      required:
        - data
        - total
        - offset
        - limit
      properties:
        data:
          type: array
          items:
            $ref: '#/components/schemas/User'
        total:
          type: integer
          description: 总记录数
          example: 156
        offset:
          type: integer
          description: 偏移量
          example: 0
        limit:
          type: integer
          description: 每页数量
          example: 20

    Error:
      type: object
      required:
        - error
      properties:
        error:
          type: object
          required:
            - code
            - message
          properties:
            code:
              type: string
              description: 错误代码
              example: VALIDATION_ERROR
            message:
              type: string
              description: 错误消息
              example: 请求参数验证失败
            details:
              type: array
              description: 错误详情
              items:
                type: object
                properties:
                  field:
                    type: string
                    example: email
                  message:
                    type: string
                    example: 邮箱格式不正确
```

#### responses（响应模板）

```yaml
components:
  responses:
    Unauthorized:
      description: 未认证
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: UNAUTHORIZED
              message: 未提供认证令牌

    Forbidden:
      description: 无权限
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: FORBIDDEN
              message: 您没有权限访问此资源

    NotFound:
      description: 资源不存在
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: NOT_FOUND
              message: 资源不存在

    ValidationError:
      description: 参数验证失败
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: VALIDATION_ERROR
              message: 请求参数验证失败
              details:
                - field: email
                  message: 邮箱格式不正确

    RateLimitExceeded:
      description: 请求频率超限
      headers:
        X-RateLimit-Limit:
          schema:
            type: integer
          description: 限流上限
        X-RateLimit-Remaining:
          schema:
            type: integer
          description: 剩余请求数
        X-RateLimit-Reset:
          schema:
            type: integer
          description: 重置时间戳
        Retry-After:
          schema:
            type: integer
          description: 重试等待秒数
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: RATE_LIMIT_EXCEEDED
              message: 请求过于频繁，请60秒后再试
```

#### securitySchemes（认证方案）

```yaml
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: |
        使用JWT Bearer Token认证。

        在Authorization请求头中提供token：
        ```
        Authorization: Bearer <token>
        ```

    apiKey:
      type: apiKey
      in: header
      name: X-API-Key
      description: API Key认证

    oauth2:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://example.com/oauth/authorize
          tokenUrl: https://example.com/oauth/token
          scopes:
            read: 读取权限
            write: 写入权限
            admin: 管理员权限
```

#### parameters（参数模板）

```yaml
components:
  parameters:
    offsetParam:
      name: offset
      in: query
      description: 偏移量，从0开始
      schema:
        type: integer
        default: 0
        minimum: 0

    limitParam:
      name: limit
      in: query
      description: 每页数量，最大100
      schema:
        type: integer
        default: 20
        minimum: 1
        maximum: 100

    sortParam:
      name: sort
      in: query
      description: |
        排序字段，使用逗号分隔多个字段。
        使用减号前缀表示降序。

        示例：
        - `sort=name` - 按name升序
        - `sort=-createdAt` - 按createdAt降序
        - `sort=status,-createdAt` - 先按status升序，再按createdAt降序
      schema:
        type: string
        example: "-createdAt"

    fieldsParam:
      name: fields
      in: query
      description: |
        指定返回的字段，使用逗号分隔。

        示例：
        - `fields=id,name,email`
      schema:
        type: string
        example: "id,name,email"
```

### security对象

定义全局安全要求。

```yaml
# 全局要求认证
security:
  - bearerAuth: []

paths:
  /public/posts:
    get:
      summary: 获取公开文章（无需认证）
      security: []  # 覆盖全局设置
      # ...

  /admin/users:
    get:
      summary: 管理员获取用户列表
      security:
        - bearerAuth: []  # 需要认证
      # ...
```

### tags对象

定义标签用于分组端点。

```yaml
tags:
  - name: Users
    description: 用户管理相关接口
  - name: Orders
    description: 订单管理相关接口
  - name: Auth
    description: 认证相关接口

paths:
  /users:
    get:
      tags:
        - Users
      # ...

  /orders:
    get:
      tags:
        - Orders
      # ...
```

### OpenAPI最佳实践

#### 1. 使用$ref引用

避免重复定义，使用引用提高可维护性。

**✓ 推荐：**
```yaml
responses:
  '200':
    description: 成功
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/User'
  '404':
    $ref: '#/components/responses/NotFound'
```

**✗ 不推荐：**
```yaml
responses:
  '200':
    description: 成功
    content:
      application/json:
        schema:
          type: object
          properties:
            id:
              type: integer
            name:
              type: string
            # ... 重复定义
```

#### 2. 提供示例

为schema和响应提供示例值。

```yaml
components:
  schemas:
    User:
      type: object
      properties:
        name:
          type: string
          example: 张三
        email:
          type: string
          example: zhangsan@example.com
```

#### 3. 详细的描述

为所有字段、参数、响应提供清晰的描述。

```yaml
parameters:
  - name: status
    in: query
    description: |
      过滤用户状态。

      可选值：
      - `active`: 活跃用户
      - `inactive`: 非活跃用户
      - `suspended`: 已暂停用户
    schema:
      type: string
      enum: [active, inactive, suspended]
```

#### 4. 使用operationId

为每个操作提供唯一的operationId，便于代码生成。

```yaml
paths:
  /users:
    get:
      operationId: listUsers
      # ...
    post:
      operationId: createUser
      # ...
  /users/{userId}:
    get:
      operationId: getUser
      # ...
```

#### 5. 文档化所有响应

包括成功和所有可能的错误响应。

```yaml
responses:
  '200':
    description: 成功
  '400':
    $ref: '#/components/responses/ValidationError'
  '401':
    $ref: '#/components/responses/Unauthorized'
  '403':
    $ref: '#/components/responses/Forbidden'
  '404':
    $ref: '#/components/responses/NotFound'
  '429':
    $ref: '#/components/responses/RateLimitExceeded'
  '500':
    description: 服务器内部错误
```

## 端点文档要求

### 基本原则

每个API端点必须提供完整的文档，包括描述、参数、请求体、响应和错误处理。

### 必需元素

#### 1. 摘要和描述

**summary（必需）：** 简短的操作描述（一句话）

**description（推荐）：** 详细的操作说明，支持Markdown格式

```yaml
paths:
  /users/{userId}:
    get:
      summary: 获取用户信息
      description: |
        根据用户ID获取用户的详细信息。

        ## 权限要求
        - 用户可以获取自己的信息
        - 管理员可以获取任何用户的信息

        ## 注意事项
        - 敏感字段（如密码）不会返回
        - 已删除的用户返回404
```

#### 2. 操作ID

**operationId（推荐）：** 唯一标识操作，用于代码生成

```yaml
paths:
  /users:
    get:
      operationId: listUsers
    post:
      operationId: createUser
  /users/{userId}:
    get:
      operationId: getUser
    patch:
      operationId: updateUser
    delete:
      operationId: deleteUser
```

#### 3. 参数文档

所有参数必须文档化，包括路径参数、查询参数、请求头。

```yaml
parameters:
  - name: userId
    in: path
    required: true
    description: 用户ID
    schema:
      type: string
    example: "123"

  - name: status
    in: query
    required: false
    description: 过滤用户状态
    schema:
      type: string
      enum: [active, inactive, suspended]
      default: active
```

#### 4. 请求体文档

POST、PUT、PATCH请求必须文档化请求体。

```yaml
requestBody:
  required: true
  description: 用户创建数据
  content:
    application/json:
      schema:
        $ref: '#/components/schemas/CreateUserRequest'
      examples:
        basicUser:
          summary: 基本用户
          value:
            name: 张三
            email: zhangsan@example.com
```

#### 5. 响应文档

必须文档化所有可能的响应状态码。

```yaml
responses:
  '200':
    description: 成功
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/User'
  '400':
    $ref: '#/components/responses/ValidationError'
  '401':
    $ref: '#/components/responses/Unauthorized'
  '403':
    $ref: '#/components/responses/Forbidden'
  '404':
    $ref: '#/components/responses/NotFound'
```

### 文档质量检查清单

- [ ] 提供了summary和description
- [ ] 设置了operationId
- [ ] 添加了tags标签
- [ ] 文档化了所有参数
- [ ] 文档化了请求体（如适用）
- [ ] 文档化了所有响应状态码
- [ ] 提供了响应示例
- [ ] 说明了认证和权限要求

## 交互式文档

### 基本原则

API文档应提供交互式界面，允许开发者直接在文档中测试API，提高开发效率和文档可用性。

### Swagger UI集成

#### 基本配置

使用Swagger UI展示OpenAPI文档。

**HTML示例：**
```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>API文档</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.SwaggerUIStandalonePreset
        ],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>
```

#### 自定义配置

```javascript
SwaggerUIBundle({
  url: "/openapi.yaml",
  dom_id: '#swagger-ui',

  // 深度链接
  deepLinking: true,

  // 显示请求持续时间
  displayRequestDuration: true,

  // 默认展开级别
  docExpansion: "list", // none, list, full

  // 过滤
  filter: true,

  // 显示扩展
  showExtensions: true,

  // 显示通用扩展
  showCommonExtensions: true,

  // 默认模型展开深度
  defaultModelsExpandDepth: 1,

  // 默认模型渲染
  defaultModelRendering: "example", // example, model

  // 持久化授权
  persistAuthorization: true,

  // 预设
  presets: [
    SwaggerUIBundle.presets.apis,
    SwaggerUIBundle.SwaggerUIStandalonePreset
  ],

  // 布局
  layout: "BaseLayout"
});
```

### Try it out功能

#### 启用测试功能

Swagger UI默认启用"Try it out"功能，允许用户直接发送请求。

**功能特性：**
- 填写参数值
- 编辑请求体
- 发送实际请求
- 查看响应结果
- 查看请求详情（URL、Headers、Body）

#### 配置CORS

为了支持浏览器中的API测试，服务器必须配置CORS。

**Go示例：**
```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### 认证测试支持

#### Bearer Token认证

Swagger UI支持在界面中输入认证凭证。

**OpenAPI配置：**
```yaml
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: |
        输入JWT token进行认证。

        获取token：
        1. 调用 POST /api/v1/auth/login
        2. 复制返回的accessToken
        3. 点击右上角"Authorize"按钮
        4. 粘贴token并点击"Authorize"

security:
  - bearerAuth: []
```

**用户操作流程：**
1. 点击"Authorize"按钮
2. 在弹出框中输入token
3. 点击"Authorize"确认
4. 后续请求自动携带token

#### API Key认证

```yaml
components:
  securitySchemes:
    apiKey:
      type: apiKey
      in: header
      name: X-API-Key
      description: |
        输入API Key进行认证。

        获取API Key：
        1. 登录控制台
        2. 进入"API密钥"页面
        3. 创建新密钥
        4. 复制密钥值

security:
  - apiKey: []
```

### Redoc集成（可选）

Redoc提供更美观的文档展示，但不支持"Try it out"。

**HTML示例：**
```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>API文档</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    body {
      margin: 0;
      padding: 0;
    }
  </style>
</head>
<body>
  <redoc spec-url="/openapi.yaml"></redoc>
  <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>
```

**配置选项：**
```html
<redoc
  spec-url="/openapi.yaml"
  hide-download-button
  hide-hostname
  theme='{
    "colors": {
      "primary": {
        "main": "#1976d2"
      }
    },
    "typography": {
      "fontSize": "14px",
      "fontFamily": "Arial, sans-serif"
    }
  }'
></redoc>
```

### 文档主题定制

#### Swagger UI主题

**自定义CSS：**
```css
/* 自定义颜色 */
.swagger-ui .topbar {
  background-color: #1976d2;
}

.swagger-ui .btn.authorize {
  background-color: #4caf50;
  border-color: #4caf50;
}

.swagger-ui .btn.authorize:hover {
  background-color: #45a049;
}

/* 自定义字体 */
.swagger-ui {
  font-family: 'Arial', sans-serif;
}

/* 隐藏Swagger UI logo */
.swagger-ui .topbar-wrapper img {
  display: none;
}

/* 自定义logo */
.swagger-ui .topbar-wrapper::before {
  content: "我的API文档";
  font-size: 20px;
  font-weight: bold;
  color: white;
}
```

#### 添加自定义logo

```javascript
SwaggerUIBundle({
  url: "/openapi.yaml",
  dom_id: '#swagger-ui',

  // 自定义顶部栏
  customCss: `
    .topbar-wrapper img {
      content: url('/logo.png');
    }
  `,

  // 自定义站点标题
  customSiteTitle: "我的API文档"
});
```

### 多环境支持

允许用户在文档中切换不同的API环境。

**OpenAPI配置：**
```yaml
servers:
  - url: https://api.example.com/v1
    description: 生产环境
  - url: https://staging-api.example.com/v1
    description: 测试环境
  - url: http://localhost:3000/v1
    description: 本地开发环境
```

**Swagger UI自动显示服务器选择下拉框。**

### 交互式文档最佳实践

#### 1. 提供测试账号

在文档中提供测试账号，方便开发者测试。

```yaml
info:
  description: |
    # API文档

    ## 测试账号

    您可以使用以下测试账号进行API测试：

    - 用户名: `test@example.com`
    - 密码: `Test123!`

    **注意：** 测试环境数据会定期清理。
```

#### 2. 提供完整示例

为每个端点提供完整的请求和响应示例。

```yaml
paths:
  /users:
    post:
      requestBody:
        content:
          application/json:
            examples:
              complete:
                summary: 完整示例
                description: 包含所有可选字段的完整示例
                value:
                  name: 张三
                  email: zhangsan@example.com
                  phone: "13800138000"
                  address:
                    city: 北京
                    district: 朝阳区
              minimal:
                summary: 最小示例
                description: 只包含必需字段的最小示例
                value:
                  name: 张三
                  email: zhangsan@example.com
```

#### 3. 说明限流规则

在文档中明确说明限流规则，避免测试时触发限流。

```yaml
info:
  description: |
    ## 限流规则

    - 认证用户: 100次/分钟
    - 未认证用户: 20次/分钟
    - 登录端点: 5次/分钟

    超出限流时返回429状态码。
```

#### 4. 提供错误示例

为常见错误提供示例，帮助开发者理解错误处理。

```yaml
responses:
  '400':
    description: 请求参数错误
    content:
      application/json:
        examples:
          missingField:
            summary: 缺少必需字段
            value:
              error:
                code: VALIDATION_ERROR
                message: 请求参数验证失败
                details:
                  - field: email
                    message: 邮箱不能为空
          invalidFormat:
            summary: 格式错误
            value:
              error:
                code: VALIDATION_ERROR
                message: 请求参数验证失败
                details:
                  - field: email
                    message: 邮箱格式不正确
```

#### 5. 响应时间显示

启用响应时间显示，帮助开发者了解API性能。

```javascript
SwaggerUIBundle({
  displayRequestDuration: true
});
```

#### 6. 持久化认证

启用认证持久化，避免刷新页面后需要重新认证。

```javascript
SwaggerUIBundle({
  persistAuthorization: true
});
```

### Go语言集成示例

```go
package main

import (
    "embed"
    "net/http"

    "github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var openapiSpec embed.FS

//go:embed swagger-ui/*
var swaggerUI embed.FS

func main() {
    r := gin.Default()

    // 提供OpenAPI规范文件
    r.GET("/openapi.yaml", func(c *gin.Context) {
        data, _ := openapiSpec.ReadFile("openapi.yaml")
        c.Data(http.StatusOK, "application/yaml", data)
    })

    // 提供Swagger UI静态文件
    r.StaticFS("/docs", http.FS(swaggerUI))

    // API路由
    api := r.Group("/api/v1")
    {
        api.GET("/users", getUsers)
        api.POST("/users", createUser)
    }

    r.Run(":8080")
}
```

**访问文档：** `http://localhost:8080/docs`

## 文档版本管理

### 基本原则

API文档必须与API版本同步，每个API版本都应有对应的文档版本，确保文档准确反映当前API状态。

### 版本号标识

#### OpenAPI文档中的版本

```yaml
openapi: 3.0.3
info:
  title: 用户管理API
  version: 1.0.0  # API版本号
  description: |
    当前版本: v1.0.0
    发布日期: 2024-03-19

servers:
  - url: https://api.example.com/v1
    description: 生产环境 (v1)
```

**版本号格式：** 使用语义化版本（Semantic Versioning）

- `MAJOR.MINOR.PATCH`
- `1.0.0` → `1.0.1`（补丁）→ `1.1.0`（新功能）→ `2.0.0`（破坏性变更）

### 多版本文档并存

#### 目录结构

```
docs/
├── v1/
│   ├── openapi.yaml
│   └── index.html
├── v2/
│   ├── openapi.yaml
│   └── index.html
└── latest/  # 指向最新版本
    ├── openapi.yaml -> ../v2/openapi.yaml
    └── index.html -> ../v2/index.html
```

#### URL结构

```
https://docs.example.com/v1/     # v1文档
https://docs.example.com/v2/     # v2文档
https://docs.example.com/latest/ # 最新版本
```

#### 版本选择器

在文档界面中提供版本选择器。

**HTML示例：**
```html
<div class="version-selector">
  <label>API版本:</label>
  <select onchange="window.location.href=this.value">
    <option value="/docs/v2/" selected>v2 (最新)</option>
    <option value="/docs/v1/">v1</option>
  </select>
</div>
```

**Swagger UI集成：**
```javascript
SwaggerUIBundle({
  url: "/docs/v2/openapi.yaml",
  dom_id: '#swagger-ui',

  // 自定义顶部栏
  onComplete: function() {
    // 添加版本选择器
    const topbar = document.querySelector('.topbar');
    const versionSelector = document.createElement('div');
    versionSelector.innerHTML = `
      <select onchange="window.location.href='/docs/'+this.value+'/'">
        <option value="v2" selected>v2 (最新)</option>
        <option value="v1">v1</option>
      </select>
    `;
    topbar.appendChild(versionSelector);
  }
});
```

### 变更日志

#### 在文档中包含变更日志

```yaml
info:
  description: |
    # 用户管理API

    ## 变更日志

    ### v2.0.0 (2024-03-19)

    **破坏性变更:**
    - 用户资源的email字段移至contacts对象
    - 删除了已废弃的/api/users/search端点

    **新增功能:**
    - 新增用户批量操作端点
    - 支持字段选择（fields参数）

    **改进:**
    - 优化了分页性能
    - 改进了错误消息

    ### v1.2.0 (2024-02-15)

    **新增功能:**
    - 新增用户导出功能
    - 支持按角色过滤

    **Bug修复:**
    - 修复了分页偏移量计算错误

    ### v1.1.0 (2024-01-10)

    **新增功能:**
    - 新增用户状态管理
    - 支持排序参数

    ### v1.0.0 (2024-01-01)

    **初始版本**
```

#### 独立的CHANGELOG文件

```markdown
# Changelog

All notable changes to this API will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2024-03-19

### Breaking Changes
- Moved `email` field to `contacts` object in User resource
- Removed deprecated `/api/users/search` endpoint

### Added
- Batch operations for users
- Field selection support via `fields` parameter

### Changed
- Improved pagination performance
- Enhanced error messages

### Fixed
- Fixed timezone handling in date filters

## [1.2.0] - 2024-02-15

### Added
- User export functionality
- Role-based filtering

### Fixed
- Fixed pagination offset calculation

## [1.1.0] - 2024-01-10

### Added
- User status management
- Sort parameter support

## [1.0.0] - 2024-01-01

### Added
- Initial release
- User CRUD operations
- Authentication and authorization
```

### 版本废弃通知

#### 在文档中标记废弃版本

```yaml
info:
  title: 用户管理API v1
  version: 1.0.0
  description: |
    # ⚠️ 此版本已废弃

    **废弃日期:** 2024-03-01
    **停止服务日期:** 2024-09-01

    请迁移到v2版本: https://docs.example.com/v2/

    ## 迁移指南

    查看完整的迁移指南: https://docs.example.com/migration/v1-to-v2
```

#### 废弃端点标记

```yaml
paths:
  /users/search:
    get:
      summary: 搜索用户（已废弃）
      deprecated: true
      description: |
        ⚠️ **此端点已废弃，将在v2中移除。**

        请使用 `GET /users?search=keyword` 代替。

        **废弃日期:** 2024-01-01
        **移除日期:** 2024-06-01
```

### 迁移指南

#### 创建版本迁移文档

```markdown
# API v1 到 v2 迁移指南

## 概述

本指南帮助您从API v1迁移到v2。

## 破坏性变更

### 1. 用户资源结构变更

**v1:**
```json
{
  "id": 123,
  "name": "张三",
  "email": "zhangsan@example.com"
}
```

**v2:**
```json
{
  "id": 123,
  "name": "张三",
  "contacts": {
    "email": "zhangsan@example.com"
  }
}
```

**迁移步骤:**
1. 更新API基础URL: `/api/v1` → `/api/v2`
2. 更新代码访问email字段: `user.email` → `user.contacts.email`

### 2. 搜索端点变更

**v1:**
```
GET /api/v1/users/search?q=keyword
```

**v2:**
```
GET /api/v2/users?search=keyword
```

**迁移步骤:**
1. 将 `/users/search` 改为 `/users`
2. 将查询参数 `q` 改为 `search`

## 新增功能

### 字段选择

v2支持字段选择，减少数据传输：

```
GET /api/v2/users?fields=id,name,contacts.email
```

### 批量操作

v2支持批量创建和更新：

```
POST /api/v2/users/batch
```

## 时间表

- **2024-03-19:** v2发布
- **2024-06-01:** v1标记为废弃
- **2024-09-01:** v1停止服务

## 支持

如有问题，请联系: api-support@example.com
```

### 文档版本控制

#### 使用Git管理文档

```bash
docs/
├── .git/
├── v1/
│   └── openapi.yaml
├── v2/
│   └── openapi.yaml
└── README.md
```

**Git标签：**
```bash
git tag -a v1.0.0 -m "API v1.0.0"
git tag -a v1.1.0 -m "API v1.1.0"
git tag -a v2.0.0 -m "API v2.0.0"
```

#### 文档发布流程

```bash
# 1. 更新文档
vim docs/v2/openapi.yaml

# 2. 提交变更
git add docs/v2/openapi.yaml
git commit -m "docs: update v2 API documentation"

# 3. 创建标签
git tag -a v2.1.0 -m "API v2.1.0"

# 4. 推送到远程
git push origin main --tags

# 5. 触发文档部署
# (通过CI/CD自动部署)
```

### 文档版本管理最佳实践

#### 1. 版本号一致性

API版本号、文档版本号、代码版本号保持一致。

```yaml
info:
  version: 2.1.0  # 与API版本一致
```

#### 2. 保留历史版本

至少保留最近3个主版本的文档。

```
docs/
├── v1/  # 保留
├── v2/  # 保留
└── v3/  # 当前
```

#### 3. 清晰的废弃通知

提前至少3个月通知版本废弃。

```yaml
info:
  description: |
    ⚠️ **重要通知**

    v1将于2024-09-01停止服务，请尽快迁移到v2。

    迁移指南: https://docs.example.com/migration/v1-to-v2
```

#### 4. 自动化文档生成

使用CI/CD自动生成和部署文档。

```yaml
# .github/workflows/docs.yml
name: Deploy Docs

on:
  push:
    tags:
      - 'v*'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Extract version
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/v}" >> $GITHUB_OUTPUT

      - name: Deploy docs
        run: |
          # 部署到文档服务器
          rsync -avz docs/ user@docs-server:/var/www/docs/
```

#### 5. 版本对比工具

提供工具对比不同版本的差异。

```bash
# 使用openapi-diff工具
npx openapi-diff docs/v1/openapi.yaml docs/v2/openapi.yaml
```

## 代码示例

### 基本原则

API文档应提供多种编程语言的代码示例，降低集成门槛，帮助开发者快速上手。

### 支持的语言

推荐为以下语言提供代码示例：

1. **Go** - 主要开发语言
2. **JavaScript/TypeScript** - Web前端
3. **Python** - 数据处理和脚本
4. **Java** - 企业应用

### 示例内容

每个代码示例应包含：

1. **完整的请求构造**
2. **认证处理**
3. **响应解析**
4. **错误处理**

### Go语言示例

#### 基本请求

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

const (
    BaseURL = "https://api.example.com/v1"
    Token   = "your_access_token"
)

// User 用户结构
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Status    string    `json:"status"`
    CreatedAt string    `json:"createdAt"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
    Data   []User `json:"data"`
    Total  int    `json:"total"`
    Offset int    `json:"offset"`
    Limit  int    `json:"limit"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
    Error struct {
        Code    string `json:"code"`
        Message string `json:"message"`
        Details []struct {
            Field   string `json:"field"`
            Message string `json:"message"`
        } `json:"details,omitempty"`
    } `json:"error"`
}

// GetUsers 获取用户列表
func GetUsers(offset, limit int, status string) (*UserListResponse, error) {
    url := fmt.Sprintf("%s/users?offset=%d&limit=%d", BaseURL, offset, limit)
    if status != "" {
        url += fmt.Sprintf("&status=%s", status)
    }

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+Token)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode != http.StatusOK {
        var errResp ErrorResponse
        if err := json.Unmarshal(body, &errResp); err != nil {
            return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
        }
        return nil, fmt.Errorf("%s: %s", errResp.Error.Code, errResp.Error.Message)
    }

    var result UserListResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, err
    }

    return &result, nil
}

// CreateUser 创建用户
func CreateUser(name, email string) (*User, error) {
    payload := map[string]string{
        "name":  name,
        "email": email,
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", BaseURL+"/users", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+Token)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode != http.StatusCreated {
        var errResp ErrorResponse
        if err := json.Unmarshal(body, &errResp); err != nil {
            return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
        }
        return nil, fmt.Errorf("%s: %s", errResp.Error.Code, errResp.Error.Message)
    }

    var user User
    if err := json.Unmarshal(body, &user); err != nil {
        return nil, err
    }

    return &user, nil
}

func main() {
    // 获取用户列表
    users, err := GetUsers(0, 20, "active")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Total users: %d\n", users.Total)

    // 创建用户
    user, err := CreateUser("张三", "zhangsan@example.com")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Created user: %s (ID: %d)\n", user.Name, user.ID)
}
```

#### 使用SDK封装

```go
package apiclient

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// Client API客户端
type Client struct {
    BaseURL    string
    HTTPClient *http.Client
    Token      string
}

// NewClient 创建新的API客户端
func NewClient(baseURL, token string) *Client {
    return &Client{
        BaseURL: baseURL,
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        Token: token,
    }
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, int, error) {
    var reqBody io.Reader
    if body != nil {
        jsonData, err := json.Marshal(body)
        if err != nil {
            return nil, 0, err
        }
        reqBody = bytes.NewBuffer(jsonData)
    }

    req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
    if err != nil {
        return nil, 0, err
    }

    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, 0, err
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, resp.StatusCode, err
    }

    return respBody, resp.StatusCode, nil
}

// GetUsers 获取用户列表
func (c *Client) GetUsers(params map[string]string) (*UserListResponse, error) {
    path := "/users"
    if len(params) > 0 {
        path += "?"
        for k, v := range params {
            path += fmt.Sprintf("%s=%s&", k, v)
        }
        path = path[:len(path)-1]
    }

    body, status, err := c.doRequest("GET", path, nil)
    if err != nil {
        return nil, err
    }

    if status != http.StatusOK {
        var errResp ErrorResponse
        json.Unmarshal(body, &errResp)
        return nil, fmt.Errorf("%s: %s", errResp.Error.Code, errResp.Error.Message)
    }

    var result UserListResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, err
    }

    return &result, nil
}

// 使用示例
func Example() {
    client := NewClient("https://api.example.com/v1", "your_token")

    users, err := client.GetUsers(map[string]string{
        "offset": "0",
        "limit":  "20",
        "status": "active",
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d users\n", users.Total)
}
```

### JavaScript/TypeScript示例

```typescript
// api-client.ts
interface User {
  id: number;
  name: string;
  email: string;
  status: string;
  createdAt: string;
}

interface UserListResponse {
  data: User[];
  total: number;
  offset: number;
  limit: number;
}

interface ErrorResponse {
  error: {
    code: string;
    message: string;
    details?: Array<{
      field: string;
      message: string;
    }>;
  };
}

class ApiClient {
  private baseURL: string;
  private token: string;

  constructor(baseURL: string, token: string) {
    this.baseURL = baseURL;
    this.token = token;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: any
  ): Promise<T> {
    const response = await fetch(`${this.baseURL}${path}`, {
      method,
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json',
      },
      body: body ? JSON.stringify(body) : undefined,
    });

    const data = await response.json();

    if (!response.ok) {
      const error = data as ErrorResponse;
      throw new Error(`${error.error.code}: ${error.error.message}`);
    }

    return data as T;
  }

  async getUsers(params: {
    offset?: number;
    limit?: number;
    status?: string;
  } = {}): Promise<UserListResponse> {
    const queryString = new URLSearchParams(
      params as Record<string, string>
    ).toString();

    return this.request<UserListResponse>(
      'GET',
      `/users${queryString ? '?' + queryString : ''}`
    );
  }

  async createUser(data: {
    name: string;
    email: string;
  }): Promise<User> {
    return this.request<User>('POST', '/users', data);
  }

  async updateUser(id: number, data: Partial<User>): Promise<User> {
    return this.request<User>('PATCH', `/users/${id}`, data);
  }

  async deleteUser(id: number): Promise<void> {
    await this.request<void>('DELETE', `/users/${id}`);
  }
}

// 使用示例
const client = new ApiClient('https://api.example.com/v1', 'your_token');

// 获取用户列表
const users = await client.getUsers({ offset: 0, limit: 20, status: 'active' });
console.log(`Total users: ${users.total}`);

// 创建用户
const newUser = await client.createUser({
  name: '张三',
  email: 'zhangsan@example.com',
});
console.log(`Created user: ${newUser.name}`);
```

### Python示例

```python
import requests
from typing import Optional, Dict, List

class ApiClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.token = token
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        })

    def _request(self, method: str, path: str, **kwargs) -> requests.Response:
        url = f"{self.base_url}{path}"
        response = self.session.request(method, url, **kwargs)

        if not response.ok:
            error = response.json()
            raise Exception(
                f"{error['error']['code']}: {error['error']['message']}"
            )

        return response

    def get_users(
        self,
        offset: int = 0,
        limit: int = 20,
        status: Optional[str] = None
    ) -> Dict:
        params = {'offset': offset, 'limit': limit}
        if status:
            params['status'] = status

        response = self._request('GET', '/users', params=params)
        return response.json()

    def create_user(self, name: str, email: str) -> Dict:
        data = {'name': name, 'email': email}
        response = self._request('POST', '/users', json=data)
        return response.json()

    def update_user(self, user_id: int, **kwargs) -> Dict:
        response = self._request('PATCH', f'/users/{user_id}', json=kwargs)
        return response.json()

    def delete_user(self, user_id: int) -> None:
        self._request('DELETE', f'/users/{user_id}')

# 使用示例
client = ApiClient('https://api.example.com/v1', 'your_token')

# 获取用户列表
users = client.get_users(offset=0, limit=20, status='active')
print(f"Total users: {users['total']}")

# 创建用户
new_user = client.create_user('张三', 'zhangsan@example.com')
print(f"Created user: {new_user['name']}")
```

### 错误处理示例

#### Go

```go
func HandleAPIError(err error) {
    if err == nil {
        return
    }

    // 类型断言检查是否为API错误
    if apiErr, ok := err.(*APIError); ok {
        switch apiErr.Code {
        case "UNAUTHORIZED":
            fmt.Println("请先登录")
            // 跳转到登录页面
        case "FORBIDDEN":
            fmt.Println("权限不足")
        case "NOT_FOUND":
            fmt.Println("资源不存在")
        case "RATE_LIMIT_EXCEEDED":
            fmt.Printf("请求过于频繁，请%d秒后重试\n", apiErr.RetryAfter)
            time.Sleep(time.Duration(apiErr.RetryAfter) * time.Second)
            // 重试请求
        default:
            fmt.Printf("错误: %s\n", apiErr.Message)
        }
    } else {
        fmt.Printf("网络错误: %v\n", err)
    }
}
```

#### JavaScript

```javascript
async function handleAPICall() {
  try {
    const users = await client.getUsers();
    console.log(users);
  } catch (error) {
    if (error.code === 'UNAUTHORIZED') {
      // 跳转到登录页面
      window.location.href = '/login';
    } else if (error.code === 'RATE_LIMIT_EXCEEDED') {
      // 等待后重试
      await new Promise(resolve => setTimeout(resolve, error.retryAfter * 1000));
      return handleAPICall(); // 重试
    } else {
      console.error('API错误:', error.message);
    }
  }
}
```

### 代码示例最佳实践

#### 1. 完整可运行

示例代码应该是完整的，可以直接复制运行。

**✓ 好的示例：**
```go
package main

import "fmt"

func main() {
    // 完整的示例代码
}
```

**✗ 不好的示例：**
```go
// 缺少package和import
client.GetUsers()
```

#### 2. 包含错误处理

展示正确的错误处理方式。

```go
users, err := client.GetUsers()
if err != nil {
    log.Fatal(err)
}
```

#### 3. 使用真实数据

使用真实的示例数据，而不是占位符。

**✓ 好：**
```go
CreateUser("张三", "zhangsan@example.com")
```

**✗ 不好：**
```go
CreateUser("name", "email")
```

#### 4. 添加注释

为关键步骤添加注释说明。

```go
// 设置认证头
req.Header.Set("Authorization", "Bearer "+token)

// 发送请求
resp, err := client.Do(req)
```

## 文档自动生成

### 基本原则

API文档应支持从代码或配置自动生成，减少手动维护成本，确保文档与代码同步。

### Go语言文档生成

#### 使用swag生成OpenAPI文档

**安装swag：**
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

**在代码中添加注解：**

```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    _ "myapp/docs" // 导入生成的文档
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           用户管理API
// @version         1.0
// @description     用户管理API提供用户的CRUD操作
// @termsOfService  https://example.com/terms

// @contact.name   API支持
// @contact.url    https://example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   https://www.apache.org/licenses/LICENSE-2.0.html

// @host      api.example.com
// @BasePath  /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入Bearer token，格式: Bearer <token>

func main() {
    r := gin.Default()

    // Swagger文档路由
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // API路由
    r.GET("/users", GetUsers)
    r.POST("/users", CreateUser)
    r.GET("/users/:id", GetUser)
    r.PATCH("/users/:id", UpdateUser)
    r.DELETE("/users/:id", DeleteUser)

    r.Run(":8080")
}

// GetUsers godoc
// @Summary      获取用户列表
// @Description  返回分页的用户列表
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        offset  query     int     false  "偏移量"  default(0)
// @Param        limit   query     int     false  "每页数量"  default(20)  maximum(100)
// @Param        status  query     string  false  "用户状态"  Enums(active, inactive, suspended)
// @Param        sort    query     string  false  "排序字段"  default(-createdAt)
// @Success      200  {object}  UserListResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      429  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users [get]
func GetUsers(c *gin.Context) {
    // 实现逻辑
}

// CreateUser godoc
// @Summary      创建用户
// @Description  创建新用户账户
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      CreateUserRequest  true  "用户数据"
// @Success      201   {object}  User
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users [post]
func CreateUser(c *gin.Context) {
    // 实现逻辑
}

// GetUser godoc
// @Summary      获取用户信息
// @Description  根据用户ID获取用户详细信息
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id      path      string  true   "用户ID"
// @Param        fields  query     string  false  "返回字段"  example(id,name,email)
// @Success      200     {object}  User
// @Failure      401     {object}  ErrorResponse
// @Failure      403     {object}  ErrorResponse
// @Failure      404     {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/{id} [get]
func GetUser(c *gin.Context) {
    // 实现逻辑
}

// UpdateUser godoc
// @Summary      更新用户信息
// @Description  部分更新用户信息
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      string             true  "用户ID"
// @Param        user  body      UpdateUserRequest  true  "更新数据"
// @Success      200   {object}  User
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Failure      403   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/{id} [patch]
func UpdateUser(c *gin.Context) {
    // 实现逻辑
}

// DeleteUser godoc
// @Summary      删除用户
// @Description  删除指定用户
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "用户ID"
// @Success      204  "删除成功"
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users/{id} [delete]
func DeleteUser(c *gin.Context) {
    // 实现逻辑
}

// 数据模型定义

// User 用户模型
type User struct {
    ID        int    `json:"id" example:"123"`
    Name      string `json:"name" example:"张三"`
    Email     string `json:"email" example:"zhangsan@example.com"`
    Phone     string `json:"phone,omitempty" example:"13800138000"`
    Status    string `json:"status" example:"active" enums:"active,inactive,suspended"`
    Role      string `json:"role" example:"user" enums:"admin,user,guest"`
    CreatedAt string `json:"createdAt" example:"2024-03-19T10:00:00Z"`
    UpdatedAt string `json:"updatedAt" example:"2024-03-19T10:00:00Z"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required" example:"张三"`
    Email string `json:"email" binding:"required,email" example:"zhangsan@example.com"`
    Phone string `json:"phone,omitempty" example:"13800138000"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
    Name  string `json:"name,omitempty" example:"新名字"`
    Email string `json:"email,omitempty" example:"newemail@example.com"`
    Phone string `json:"phone,omitempty" example:"13900139000"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
    Data   []User `json:"data"`
    Total  int    `json:"total" example:"156"`
    Offset int    `json:"offset" example:"0"`
    Limit  int    `json:"limit" example:"20"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
    Code    string        `json:"code" example:"VALIDATION_ERROR"`
    Message string        `json:"message" example:"请求参数验证失败"`
    Details []FieldError  `json:"details,omitempty"`
}

// FieldError 字段错误
type FieldError struct {
    Field   string `json:"field" example:"email"`
    Message string `json:"message" example:"邮箱格式不正确"`
}
```

**生成文档：**
```bash
# 在项目根目录执行
swag init

# 生成的文件
# docs/
#   ├── docs.go
#   ├── swagger.json
#   └── swagger.yaml
```

**访问文档：**
```
http://localhost:8080/swagger/index.html
```

#### swag注解说明

| 注解 | 说明 | 示例 |
|------|------|------|
| @Summary | 简短描述 | @Summary 获取用户列表 |
| @Description | 详细描述 | @Description 返回分页的用户列表 |
| @Tags | 标签分组 | @Tags users |
| @Accept | 接受的内容类型 | @Accept json |
| @Produce | 生成的内容类型 | @Produce json |
| @Param | 参数定义 | @Param id path string true "用户ID" |
| @Success | 成功响应 | @Success 200 {object} User |
| @Failure | 失败响应 | @Failure 404 {object} ErrorResponse |
| @Security | 安全要求 | @Security BearerAuth |
| @Router | 路由定义 | @Router /users [get] |

### 文档验证

#### 使用openapi-generator验证

**安装：**
```bash
npm install -g @openapitools/openapi-generator-cli
```

**验证OpenAPI文档：**
```bash
openapi-generator-cli validate -i docs/swagger.yaml
```

**输出示例：**
```
Validating spec (docs/swagger.yaml)
Spec is valid.
```

#### 使用spectral进行高级验证

**安装：**
```bash
npm install -g @stoplight/spectral-cli
```

**创建规则文件 `.spectral.yaml`：**
```yaml
extends: [[spectral:oas, all]]

rules:
  # 要求所有操作都有summary
  operation-summary: error

  # 要求所有操作都有description
  operation-description: warn

  # 要求所有操作都有tags
  operation-tags: error

  # 要求所有参数都有description
  parameter-description: error

  # 要求所有响应都有description
  response-description: error

  # 要求所有schema都有example
  schema-example: warn
```

**验证文档：**
```bash
spectral lint docs/swagger.yaml
```

### CI/CD集成

#### GitHub Actions示例

```yaml
# .github/workflows/docs.yml
name: Generate and Validate API Docs

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  docs:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install swag
        run: go install github.com/swaggo/swag/cmd/swag@latest

      - name: Generate OpenAPI docs
        run: swag init

      - name: Install spectral
        run: npm install -g @stoplight/spectral-cli

      - name: Validate OpenAPI docs
        run: spectral lint docs/swagger.yaml

      - name: Check for changes
        run: |
          if [[ -n $(git status -s docs/) ]]; then
            echo "Documentation needs to be regenerated"
            echo "Run: swag init"
            exit 1
          fi

      - name: Upload docs artifact
        uses: actions/upload-artifact@v3
        with:
          name: api-docs
          path: docs/

      - name: Deploy to GitHub Pages
        if: github.ref == 'refs/heads/main'
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs
```

### 文档覆盖率检查

#### 自定义脚本检查文档完整性

```go
// tools/check-docs-coverage.go
package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
)

type OpenAPISpec struct {
    Paths map[string]map[string]Operation `json:"paths"`
}

type Operation struct {
    Summary     string `json:"summary"`
    Description string `json:"description"`
    Tags        []string `json:"tags"`
}

func main() {
    data, err := ioutil.ReadFile("docs/swagger.json")
    if err != nil {
        fmt.Println("Error reading swagger.json:", err)
        os.Exit(1)
    }

    var spec OpenAPISpec
    if err := json.Unmarshal(data, &spec); err != nil {
        fmt.Println("Error parsing swagger.json:", err)
        os.Exit(1)
    }

    totalOps := 0
    missingDocs := 0

    for path, methods := range spec.Paths {
        for method, op := range methods {
            totalOps++

            if op.Summary == "" {
                fmt.Printf("Missing summary: %s %s\n", method, path)
                missingDocs++
            }

            if op.Description == "" {
                fmt.Printf("Missing description: %s %s\n", method, path)
                missingDocs++
            }

            if len(op.Tags) == 0 {
                fmt.Printf("Missing tags: %s %s\n", method, path)
                missingDocs++
            }
        }
    }

    coverage := float64(totalOps-missingDocs) / float64(totalOps) * 100
    fmt.Printf("\nDocumentation Coverage: %.2f%%\n", coverage)
    fmt.Printf("Total Operations: %d\n", totalOps)
    fmt.Printf("Missing Documentation: %d\n", missingDocs)

    if coverage < 100 {
        os.Exit(1)
    }
}
```

**在CI中运行：**
```yaml
- name: Check docs coverage
  run: go run tools/check-docs-coverage.go
```

### 文档自动生成最佳实践

#### 1. 保持注解与代码同步

在修改API时同时更新注解。

```go
// ✓ 好的做法：修改代码时更新注解
// @Param limit query int false "每页数量" default(20) maximum(100)
func GetUsers(c *gin.Context) {
    limit := c.DefaultQuery("limit", "20")
    // 实现逻辑
}
```

#### 2. 使用结构体标签

利用Go的结构体标签自动生成schema。

```go
type User struct {
    ID    int    `json:"id" example:"123"`
    Name  string `json:"name" binding:"required" example:"张三"`
    Email string `json:"email" binding:"required,email" example:"zhangsan@example.com"`
}
```

#### 3. 提供示例值

为所有字段提供example标签。

```go
type User struct {
    Status string `json:"status" example:"active" enums:"active,inactive,suspended"`
}
```

#### 4. 文档生成自动化

在pre-commit hook中自动生成文档。

```bash
# .git/hooks/pre-commit
#!/bin/bash

# 生成文档
swag init

# 检查是否有变更
if [[ -n $(git status -s docs/) ]]; then
    git add docs/
    echo "API documentation updated"
fi
```

#### 5. 版本控制

将生成的文档提交到版本控制。

```gitignore
# 不要忽略生成的文档
# docs/
```

## 文档可访问性

### 基本原则

API文档应易于访问和搜索，支持开发者快速找到所需信息，提供良好的用户体验。

### 在线文档托管

#### 静态站点托管

**GitHub Pages：**
```yaml
# .github/workflows/deploy-docs.yml
name: Deploy Docs

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Generate docs
        run: swag init

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs
```

**访问地址：** `https://username.github.io/repo-name/`

**Netlify：**
```toml
# netlify.toml
[build]
  command = "swag init"
  publish = "docs"

[[redirects]]
  from = "/*"
  to = "/index.html"
  status = 200
```

**Vercel：**
```json
{
  "buildCommand": "swag init",
  "outputDirectory": "docs"
}
```

#### 自托管

**使用Nginx托管：**
```nginx
server {
    listen 80;
    server_name docs.example.com;

    root /var/www/docs;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # 启用gzip压缩
    gzip on;
    gzip_types text/plain text/css application/json application/javascript;

    # 缓存静态资源
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

**使用Docker部署：**
```dockerfile
FROM nginx:alpine

COPY docs/ /usr/share/nginx/html/

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

```bash
docker build -t api-docs .
docker run -d -p 80:80 api-docs
```

### 搜索功能

#### Swagger UI内置搜索

Swagger UI默认支持过滤功能。

```javascript
SwaggerUIBundle({
  url: "/openapi.yaml",
  dom_id: '#swagger-ui',
  filter: true,  // 启用搜索过滤
  deepLinking: true
});
```

#### 自定义搜索

**使用Algolia DocSearch：**
```html
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@docsearch/css@3" />

<div id="docsearch"></div>

<script src="https://cdn.jsdelivr.net/npm/@docsearch/js@3"></script>
<script>
  docsearch({
    appId: 'YOUR_APP_ID',
    apiKey: 'YOUR_API_KEY',
    indexName: 'api-docs',
    container: '#docsearch',
  });
</script>
```

**使用Lunr.js本地搜索：**
```javascript
// 构建搜索索引
const idx = lunr(function () {
  this.ref('id');
  this.field('title');
  this.field('description');
  this.field('path');

  documents.forEach(function (doc) {
    this.add(doc);
  }, this);
});

// 搜索
const results = idx.search(query);
```

### 分类和标签

#### 使用tags组织端点

```yaml
tags:
  - name: Users
    description: 用户管理相关接口
    externalDocs:
      description: 用户管理文档
      url: https://docs.example.com/users

  - name: Orders
    description: 订单管理相关接口
    externalDocs:
      description: 订单管理文档
      url: https://docs.example.com/orders

  - name: Auth
    description: 认证相关接口
    externalDocs:
      description: 认证文档
      url: https://docs.example.com/auth

paths:
  /users:
    get:
      tags:
        - Users
      summary: 获取用户列表

  /orders:
    get:
      tags:
        - Orders
      summary: 获取订单列表
```

#### 自定义分组

**按功能模块分组：**
```yaml
x-tagGroups:
  - name: 用户管理
    tags:
      - Users
      - User Profiles
      - User Settings

  - name: 订单管理
    tags:
      - Orders
      - Order Items
      - Payments

  - name: 系统管理
    tags:
      - Auth
      - Admin
```

### 导航和目录

#### 侧边栏导航

**Redoc示例：**
```html
<redoc
  spec-url="/openapi.yaml"
  scroll-y-offset="nav"
  hide-download-button
  theme='{
    "sidebar": {
      "backgroundColor": "#f5f5f5",
      "textColor": "#333"
    }
  }'
></redoc>
```

#### 面包屑导航

```html
<nav aria-label="breadcrumb">
  <ol class="breadcrumb">
    <li class="breadcrumb-item"><a href="/">首页</a></li>
    <li class="breadcrumb-item"><a href="/docs">文档</a></li>
    <li class="breadcrumb-item active">API参考</li>
  </ol>
</nav>
```

### 多语言支持

#### 国际化文档

**目录结构：**
```
docs/
├── zh-CN/
│   ├── openapi.yaml
│   └── index.html
├── en-US/
│   ├── openapi.yaml
│   └── index.html
└── index.html  # 语言选择页
```

**语言选择器：**
```html
<div class="language-selector">
  <select onchange="window.location.href='/docs/'+this.value+'/'">
    <option value="zh-CN" selected>简体中文</option>
    <option value="en-US">English</option>
  </select>
</div>
```

**OpenAPI中的多语言描述：**
```yaml
info:
  title: User Management API
  description: |
    User management API provides CRUD operations for users.

    ---

    用户管理API提供用户的CRUD操作。
```

### 响应式设计

#### 移动端适配

```css
/* 响应式样式 */
@media (max-width: 768px) {
  .swagger-ui .topbar {
    padding: 10px;
  }

  .swagger-ui .info {
    margin: 20px 10px;
  }

  .swagger-ui .scheme-container {
    padding: 10px;
  }
}
```

#### 触摸优化

```css
/* 增大可点击区域 */
.swagger-ui .opblock-summary {
  min-height: 48px;
  padding: 12px;
}

/* 优化按钮大小 */
.swagger-ui .btn {
  min-height: 44px;
  padding: 10px 20px;
}
```

### 性能优化

#### 文档加载优化

**延迟加载：**
```javascript
SwaggerUIBundle({
  url: "/openapi.yaml",
  dom_id: '#swagger-ui',
  // 默认折叠所有操作
  docExpansion: "none",
  // 延迟加载模型
  defaultModelsExpandDepth: -1
});
```

**CDN加速：**
```html
<!-- 使用CDN加载Swagger UI -->
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
```

#### 文档压缩

```bash
# 压缩OpenAPI文档
gzip -k docs/swagger.yaml

# Nginx配置
gzip on;
gzip_types application/json application/yaml text/yaml;
```

### 访问控制

#### 公开文档

```nginx
# 公开访问
location /docs {
    root /var/www;
    index index.html;
}
```

#### 需要认证的文档

```nginx
# 基本认证
location /docs {
    auth_basic "API Documentation";
    auth_basic_user_file /etc/nginx/.htpasswd;
    root /var/www;
}
```

**生成密码文件：**
```bash
htpasswd -c /etc/nginx/.htpasswd username
```

#### IP白名单

```nginx
location /docs {
    allow 192.168.1.0/24;
    allow 10.0.0.0/8;
    deny all;
    root /var/www;
}
```

### 文档监控

#### 访问统计

**使用Google Analytics：**
```html
<!-- Google Analytics -->
<script async src="https://www.googletagmanager.com/gtag/js?id=GA_MEASUREMENT_ID"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', 'GA_MEASUREMENT_ID');
</script>
```

**自定义事件跟踪：**
```javascript
// 跟踪API端点查看
document.querySelectorAll('.opblock').forEach(block => {
  block.addEventListener('click', function() {
    const path = this.dataset.path;
    const method = this.dataset.method;
    gtag('event', 'view_endpoint', {
      'endpoint_path': path,
      'http_method': method
    });
  });
});
```

### 文档反馈

#### 反馈按钮

```html
<div class="feedback-widget">
  <button onclick="openFeedback()">
    文档有问题？点击反馈
  </button>
</div>

<script>
function openFeedback() {
  const currentPage = window.location.href;
  const issueUrl = `https://github.com/org/repo/issues/new?title=文档反馈&body=页面: ${currentPage}`;
  window.open(issueUrl, '_blank');
}
</script>
```

#### 评分系统

```html
<div class="doc-rating">
  <p>这个文档有帮助吗？</p>
  <button onclick="rate('helpful')">👍 有帮助</button>
  <button onclick="rate('not-helpful')">👎 没帮助</button>
</div>

<script>
function rate(rating) {
  fetch('/api/feedback', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      page: window.location.pathname,
      rating: rating
    })
  });
  alert('感谢您的反馈！');
}
</script>
```

### 文档可访问性最佳实践

#### 1. 提供稳定的URL

确保文档URL长期有效。

```
✓ https://docs.example.com/v1/
✓ https://api.example.com/docs/v1/
✗ https://temp-docs.example.com/
```

#### 2. 支持深度链接

允许直接链接到特定端点。

```
https://docs.example.com/v1/#/Users/getUser
```

#### 3. 提供离线访问

支持下载文档供离线查看。

```html
<a href="/openapi.yaml" download>下载OpenAPI规范</a>
<a href="/docs.pdf" download>下载PDF文档</a>
```

#### 4. 快速加载

优化文档加载速度，首屏加载时间<3秒。

```javascript
// 使用懒加载
SwaggerUIBundle({
  docExpansion: "none",
  defaultModelsExpandDepth: -1
});
```

#### 5. 清晰的导航

提供清晰的导航结构，用户能快速找到所需内容。

```yaml
tags:
  - name: Getting Started
  - name: Authentication
  - name: Users
  - name: Orders
```

#### 6. 搜索优化

确保文档可被搜索引擎索引。

```html
<meta name="description" content="用户管理API文档">
<meta name="keywords" content="API, REST, 用户管理">
```

## 文档完整性验证

### 基本原则

API文档必须经过完整性检查，确保所有端点都有文档，所有必需元素都已提供。

### 缺失文档检测

#### 自动检测未文档化的端点

**Go脚本示例：**
```go
// tools/check-undocumented-endpoints.go
package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "strings"
)

type OpenAPISpec struct {
    Paths map[string]map[string]interface{} `json:"paths"`
}

type Route struct {
    Method string
    Path   string
}

func main() {
    // 读取OpenAPI文档
    data, err := ioutil.ReadFile("docs/swagger.json")
    if err != nil {
        fmt.Println("Error reading swagger.json:", err)
        os.Exit(1)
    }

    var spec OpenAPISpec
    if err := json.Unmarshal(data, &spec); err != nil {
        fmt.Println("Error parsing swagger.json:", err)
        os.Exit(1)
    }

    // 获取已文档化的端点
    documented := make(map[string]bool)
    for path, methods := range spec.Paths {
        for method := range methods {
            key := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
            documented[key] = true
        }
    }

    // 获取实际的路由（需要根据实际框架实现）
    actualRoutes := getActualRoutes()

    // 检查未文档化的端点
    undocumented := []string{}
    for _, route := range actualRoutes {
        key := fmt.Sprintf("%s %s", route.Method, route.Path)
        if !documented[key] {
            undocumented = append(undocumented, key)
        }
    }

    if len(undocumented) > 0 {
        fmt.Println("未文档化的端点:")
        for _, endpoint := range undocumented {
            fmt.Printf("  - %s\n", endpoint)
        }
        os.Exit(1)
    }

    fmt.Println("所有端点都已文档化 ✓")
}

func getActualRoutes() []Route {
    // 这里需要根据实际框架实现
    // 例如，解析Gin的路由注册
    return []Route{
        {Method: "GET", Path: "/api/v1/users"},
        {Method: "POST", Path: "/api/v1/users"},
        {Method: "GET", Path: "/api/v1/users/:id"},
        {Method: "PATCH", Path: "/api/v1/users/:id"},
        {Method: "DELETE", Path: "/api/v1/users/:id"},
    }
}
```

#### CI集成

```yaml
# .github/workflows/check-docs.yml
name: Check API Documentation

on:
  pull_request:
    paths:
      - '**.go'
      - 'docs/**'

jobs:
  check-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Generate docs
        run: swag init

      - name: Check for undocumented endpoints
        run: go run tools/check-undocumented-endpoints.go

      - name: Validate OpenAPI spec
        run: |
          npm install -g @stoplight/spectral-cli
          spectral lint docs/swagger.yaml
```

### 文档覆盖率报告

#### 生成覆盖率报告

```go
// tools/docs-coverage-report.go
package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
)

type OpenAPISpec struct {
    Paths map[string]map[string]Operation `json:"paths"`
}

type Operation struct {
    Summary     string                 `json:"summary"`
    Description string                 `json:"description"`
    Tags        []string               `json:"tags"`
    Parameters  []Parameter            `json:"parameters"`
    Responses   map[string]Response    `json:"responses"`
}

type Parameter struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type Response struct {
    Description string `json:"description"`
}

type CoverageReport struct {
    TotalEndpoints      int
    DocumentedEndpoints int
    MissingSummary      []string
    MissingDescription  []string
    MissingTags         []string
    MissingExamples     []string
    CoveragePercent     float64
}

func main() {
    data, err := ioutil.ReadFile("docs/swagger.json")
    if err != nil {
        fmt.Println("Error:", err)
        os.Exit(1)
    }

    var spec OpenAPISpec
    json.Unmarshal(data, &spec)

    report := CoverageReport{}

    for path, methods := range spec.Paths {
        for method, op := range methods {
            report.TotalEndpoints++
            endpoint := fmt.Sprintf("%s %s", method, path)

            issues := 0

            if op.Summary == "" {
                report.MissingSummary = append(report.MissingSummary, endpoint)
                issues++
            }

            if op.Description == "" {
                report.MissingDescription = append(report.MissingDescription, endpoint)
                issues++
            }

            if len(op.Tags) == 0 {
                report.MissingTags = append(report.MissingTags, endpoint)
                issues++
            }

            // 检查参数描述
            for _, param := range op.Parameters {
                if param.Description == "" {
                    issues++
                }
            }

            // 检查响应描述
            for _, resp := range op.Responses {
                if resp.Description == "" {
                    issues++
                }
            }

            if issues == 0 {
                report.DocumentedEndpoints++
            }
        }
    }

    report.CoveragePercent = float64(report.DocumentedEndpoints) / float64(report.TotalEndpoints) * 100

    // 输出报告
    fmt.Println("=== API文档覆盖率报告 ===")
    fmt.Printf("\n总端点数: %d\n", report.TotalEndpoints)
    fmt.Printf("完整文档: %d\n", report.DocumentedEndpoints)
    fmt.Printf("覆盖率: %.2f%%\n\n", report.CoveragePercent)

    if len(report.MissingSummary) > 0 {
        fmt.Println("缺少摘要的端点:")
        for _, ep := range report.MissingSummary {
            fmt.Printf("  - %s\n", ep)
        }
        fmt.Println()
    }

    if len(report.MissingDescription) > 0 {
        fmt.Println("缺少描述的端点:")
        for _, ep := range report.MissingDescription {
            fmt.Printf("  - %s\n", ep)
        }
        fmt.Println()
    }

    if len(report.MissingTags) > 0 {
        fmt.Println("缺少标签的端点:")
        for _, ep := range report.MissingTags {
            fmt.Printf("  - %s\n", ep)
        }
        fmt.Println()
    }

    // 生成HTML报告
    generateHTMLReport(report)

    if report.CoveragePercent < 100 {
        os.Exit(1)
    }
}

func generateHTMLReport(report CoverageReport) {
    html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>API文档覆盖率报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .summary { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .metric { font-size: 24px; font-weight: bold; }
        .issues { margin-top: 20px; }
        .issue-list { list-style: none; padding: 0; }
        .issue-item { padding: 5px; border-left: 3px solid #ff9800; margin: 5px 0; padding-left: 10px; }
    </style>
</head>
<body>
    <h1>API文档覆盖率报告</h1>
    <div class="summary">
        <div class="metric">覆盖率: %.2f%%</div>
        <p>总端点数: %d</p>
        <p>完整文档: %d</p>
    </div>
    <div class="issues">
        <h2>待改进项</h2>
        <h3>缺少摘要 (%d)</h3>
        <ul class="issue-list">
`, report.CoveragePercent, report.TotalEndpoints, report.DocumentedEndpoints, len(report.MissingSummary))

    for _, ep := range report.MissingSummary {
        html += fmt.Sprintf(`            <li class="issue-item">%s</li>`, ep)
    }

    html += `
        </ul>
    </div>
</body>
</html>
`

    ioutil.WriteFile("docs/coverage-report.html", []byte(html), 0644)
    fmt.Println("HTML报告已生成: docs/coverage-report.html")
}
```

### 必需字段检查

#### Spectral规则配置

```yaml
# .spectral.yaml
extends: [[spectral:oas, all]]

rules:
  # 所有操作必须有summary
  operation-summary:
    severity: error
    message: "操作必须有summary"

  # 所有操作必须有description
  operation-description:
    severity: error
    message: "操作必须有description"

  # 所有操作必须有tags
  operation-tags:
    severity: error
    message: "操作必须有tags"

  # 所有操作必须有operationId
  operation-operationId:
    severity: error
    message: "操作必须有operationId"

  # 所有参数必须有description
  parameter-description:
    severity: error
    message: "参数必须有description"

  # 所有响应必须有description
  response-description:
    severity: error
    message: "响应必须有description"

  # 所有schema必须有example
  schema-example:
    severity: warn
    message: "Schema应该有example"

  # 所有属性必须有description
  schema-properties-descriptions:
    severity: warn
    message: "Schema属性应该有description"

  # 成功响应必须有2xx状态码
  operation-success-response:
    severity: error
    message: "操作必须至少有一个2xx响应"

  # 自定义规则：检查错误响应
  operation-error-responses:
    description: "操作应该文档化常见错误响应"
    severity: warn
    given: "$.paths[*][*]"
    then:
      - field: "responses"
        function: schema
        functionOptions:
          schema:
            type: object
            properties:
              "400": {}
              "401": {}
              "500": {}
```

### 文档质量门禁

#### 设置质量标准

```yaml
# .github/workflows/docs-quality-gate.yml
name: Documentation Quality Gate

on:
  pull_request:
    paths:
      - '**.go'
      - 'docs/**'

jobs:
  quality-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Generate docs
        run: swag init

      - name: Check coverage
        id: coverage
        run: |
          go run tools/docs-coverage-report.go > coverage.txt
          COVERAGE=$(grep "覆盖率:" coverage.txt | awk '{print $2}' | sed 's/%//')
          echo "coverage=$COVERAGE" >> $GITHUB_OUTPUT

      - name: Validate coverage threshold
        run: |
          THRESHOLD=95
          COVERAGE=${{ steps.coverage.outputs.coverage }}
          if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
            echo "文档覆盖率 ($COVERAGE%) 低于阈值 ($THRESHOLD%)"
            exit 1
          fi
          echo "文档覆盖率: $COVERAGE% ✓"

      - name: Validate OpenAPI spec
        run: |
          npm install -g @stoplight/spectral-cli
          spectral lint docs/swagger.yaml --fail-severity=error

      - name: Comment PR
        uses: actions/github-script@v6
        with:
          script: |
            const coverage = ${{ steps.coverage.outputs.coverage }};
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `## 📊 API文档质量报告\n\n✅ 文档覆盖率: ${coverage}%\n\n查看详细报告: [coverage-report.html](../docs/coverage-report.html)`
            });
```

### 文档审查清单

#### Pull Request模板

```markdown
<!-- .github/pull_request_template.md -->
## API变更

- [ ] 新增API端点
- [ ] 修改现有API
- [ ] 删除API端点

## 文档检查清单

- [ ] 所有新端点都已添加Swagger注解
- [ ] 所有端点都有summary和description
- [ ] 所有参数都有描述和示例
- [ ] 所有响应都有描述和示例
- [ ] 错误响应已文档化（400, 401, 403, 404, 500）
- [ ] 已添加适当的tags
- [ ] 已运行 `swag init` 生成文档
- [ ] 文档覆盖率 >= 95%
- [ ] Spectral验证通过

## 测试

- [ ] 在Swagger UI中测试了所有新端点
- [ ] 验证了请求和响应示例的正确性
```

### 持续监控

#### 文档质量仪表板

```go
// tools/docs-dashboard.go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type DashboardData struct {
    Timestamp       time.Time
    TotalEndpoints  int
    Coverage        float64
    Issues          int
    LastUpdate      string
}

func main() {
    http.HandleFunc("/api/docs-metrics", func(w http.ResponseWriter, r *http.Request) {
        // 读取最新的覆盖率报告
        report := getLatestReport()

        data := DashboardData{
            Timestamp:      time.Now(),
            TotalEndpoints: report.TotalEndpoints,
            Coverage:       report.CoveragePercent,
            Issues:         len(report.MissingSummary) + len(report.MissingDescription),
            LastUpdate:     time.Now().Format("2006-01-02 15:04:05"),
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(data)
    })

    http.Handle("/", http.FileServer(http.Dir("./dashboard")))

    fmt.Println("Dashboard running on http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}
```

**仪表板HTML：**
```html
<!-- dashboard/index.html -->
<!DOCTYPE html>
<html>
<head>
    <title>API文档质量仪表板</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <h1>API文档质量仪表板</h1>

    <div class="metrics">
        <div class="metric-card">
            <h2 id="coverage">--%</h2>
            <p>文档覆盖率</p>
        </div>
        <div class="metric-card">
            <h2 id="endpoints">--</h2>
            <p>总端点数</p>
        </div>
        <div class="metric-card">
            <h2 id="issues">--</h2>
            <p>待修复问题</p>
        </div>
    </div>

    <canvas id="coverageChart"></canvas>

    <script>
        async function updateMetrics() {
            const response = await fetch('/api/docs-metrics');
            const data = await response.json();

            document.getElementById('coverage').textContent = data.Coverage.toFixed(2) + '%';
            document.getElementById('endpoints').textContent = data.TotalEndpoints;
            document.getElementById('issues').textContent = data.Issues;
        }

        updateMetrics();
        setInterval(updateMetrics, 60000); // 每分钟更新
    </script>
</body>
</html>
```

### 文档完整性最佳实践

#### 1. 自动化检查

在CI/CD中自动检查文档完整性，不通过则阻止合并。

#### 2. 设置覆盖率目标

设定文档覆盖率目标（推荐95%以上），并持续监控。

#### 3. 定期审查

定期审查文档质量，更新过时的文档。

#### 4. 开发者培训

培训开发者正确编写API文档注解。

#### 5. 文档优先

在编写代码前先编写API文档（文档驱动开发）。

#### 6. 版本控制

将文档纳入版本控制，跟踪文档变更历史。
