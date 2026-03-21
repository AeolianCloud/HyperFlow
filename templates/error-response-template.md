# 错误响应模板

## 标准错误格式

所有错误响应使用统一的JSON格式：

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

## 常见错误代码

### 客户端错误 (4xx)

#### VALIDATION_ERROR (400)
请求参数验证失败

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "请求参数验证失败",
    "details": [
      {
        "field": "email",
        "message": "邮箱格式不正确"
      },
      {
        "field": "name",
        "message": "姓名不能为空"
      }
    ]
  }
}
```

#### INVALID_REQUEST (400)
请求格式错误

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "请求格式错误，请检查JSON格式"
  }
}
```

#### UNAUTHORIZED (401)
未认证或认证失败

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "未提供认证令牌"
  }
}
```

#### TOKEN_EXPIRED (401)
访问令牌已过期

```json
{
  "error": {
    "code": "TOKEN_EXPIRED",
    "message": "访问令牌已过期，请刷新令牌"
  }
}
```

#### INVALID_TOKEN (401)
认证令牌无效

```json
{
  "error": {
    "code": "INVALID_TOKEN",
    "message": "认证令牌无效"
  }
}
```

#### FORBIDDEN (403)
无权限访问

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "您没有权限访问此资源"
  }
}
```

#### NOT_FOUND (404)
资源不存在

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "用户不存在"
  }
}
```

#### METHOD_NOT_ALLOWED (405)
HTTP方法不支持

```json
{
  "error": {
    "code": "METHOD_NOT_ALLOWED",
    "message": "此端点不支持POST方法"
  }
}
```

#### CONFLICT (409)
资源冲突

```json
{
  "error": {
    "code": "CONFLICT",
    "message": "该邮箱已被注册"
  }
}
```

#### DUPLICATE_RESOURCE (409)
资源已存在

```json
{
  "error": {
    "code": "DUPLICATE_RESOURCE",
    "message": "该用户名已被使用"
  }
}
```

#### INVALID_DATA (422)
数据不符合业务规则

```json
{
  "error": {
    "code": "INVALID_DATA",
    "message": "库存不足，当前库存仅剩50件"
  }
}
```

#### RATE_LIMIT_EXCEEDED (429)
请求频率超限

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "请求过于频繁，请60秒后再试",
    "retryAfter": 60
  }
}
```

### 服务器错误 (5xx)

#### INTERNAL_ERROR (500)
服务器内部错误

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "服务器内部错误，请稍后重试"
  }
}
```

#### SERVICE_UNAVAILABLE (503)
服务不可用

```json
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "服务维护中，预计5分钟后恢复"
  }
}
```

## Go语言错误响应实现

### 错误结构定义

```go
package models

// ErrorResponse 错误响应
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
    Code    string       `json:"code"`
    Message string       `json:"message"`
    Details []FieldError `json:"details,omitempty"`
}

// FieldError 字段错误
type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

### 错误构造函数

```go
package errors

import (
    "net/http"
    "myapp/models"
)

// APIError API错误
type APIError struct {
    StatusCode int
    Response   models.ErrorResponse
}

func (e *APIError) Error() string {
    return e.Response.Error.Message
}

// NewValidationError 创建验证错误
func NewValidationError(message string, details []models.FieldError) *APIError {
    return &APIError{
        StatusCode: http.StatusBadRequest,
        Response: models.ErrorResponse{
            Error: models.ErrorDetail{
                Code:    "VALIDATION_ERROR",
                Message: message,
                Details: details,
            },
        },
    }
}

// NewUnauthorizedError 创建未认证错误
func NewUnauthorizedError(message string) *APIError {
    return &APIError{
        StatusCode: http.StatusUnauthorized,
        Response: models.ErrorResponse{
            Error: models.ErrorDetail{
                Code:    "UNAUTHORIZED",
                Message: message,
            },
        },
    }
}

// NewForbiddenError 创建无权限错误
func NewForbiddenError(message string) *APIError {
    return &APIError{
        StatusCode: http.StatusForbidden,
        Response: models.ErrorResponse{
            Error: models.ErrorDetail{
                Code:    "FORBIDDEN",
                Message: message,
            },
        },
    }
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(message string) *APIError {
    return &APIError{
        StatusCode: http.StatusNotFound,
        Response: models.ErrorResponse{
            Error: models.ErrorDetail{
                Code:    "NOT_FOUND",
                Message: message,
            },
        },
    }
}

// NewConflictError 创建资源冲突错误
func NewConflictError(message string) *APIError {
    return &APIError{
        StatusCode: http.StatusConflict,
        Response: models.ErrorResponse{
            Error: models.ErrorDetail{
                Code:    "CONFLICT",
                Message: message,
            },
        },
    }
}

// NewInternalError 创建服务器内部错误
func NewInternalError(message string) *APIError {
    return &APIError{
        StatusCode: http.StatusInternalServerError,
        Response: models.ErrorResponse{
            Error: models.ErrorDetail{
                Code:    "INTERNAL_ERROR",
                Message: message,
            },
        },
    }
}
```

### 错误处理中间件

```go
package middleware

import (
    "net/http"
    "myapp/errors"

    "github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // 检查是否有错误
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err

            // 处理API错误
            if apiErr, ok := err.(*errors.APIError); ok {
                c.JSON(apiErr.StatusCode, apiErr.Response)
                return
            }

            // 处理未知错误
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": gin.H{
                    "code":    "INTERNAL_ERROR",
                    "message": "服务器内部错误，请稍后重试",
                },
            })
        }
    }
}
```

### 使用示例

```go
package handlers

import (
    "myapp/errors"
    "myapp/models"

    "github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context) {
    var req models.CreateUserRequest

    // 绑定请求体
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Error(errors.NewValidationError("请求参数验证失败", []models.FieldError{
            {Field: "email", Message: "邮箱格式不正确"},
        }))
        return
    }

    // 检查邮箱是否已存在
    if userExists(req.Email) {
        c.Error(errors.NewConflictError("该邮箱已被注册"))
        return
    }

    // 创建用户
    user, err := createUser(req)
    if err != nil {
        c.Error(errors.NewInternalError("创建用户失败"))
        return
    }

    c.JSON(201, user)
}

func GetUser(c *gin.Context) {
    userID := c.Param("id")

    user, err := findUserByID(userID)
    if err != nil {
        c.Error(errors.NewNotFoundError("用户不存在"))
        return
    }

    // 检查权限
    if !canAccessUser(c, userID) {
        c.Error(errors.NewForbiddenError("您只能访问自己的信息"))
        return
    }

    c.JSON(200, user)
}
```

## OpenAPI错误响应定义

```yaml
components:
  schemas:
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
            message:
              type: string
              description: 错误消息
            details:
              type: array
              description: 错误详情
              items:
                type: object
                properties:
                  field:
                    type: string
                  message:
                    type: string

  responses:
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

    Conflict:
      description: 资源冲突
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: CONFLICT
              message: 该邮箱已被注册

    RateLimitExceeded:
      description: 请求频率超限
      headers:
        X-RateLimit-Limit:
          schema:
            type: integer
        X-RateLimit-Remaining:
          schema:
            type: integer
        X-RateLimit-Reset:
          schema:
            type: integer
        Retry-After:
          schema:
            type: integer
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: RATE_LIMIT_EXCEEDED
              message: 请求过于频繁，请60秒后再试

    InternalError:
      description: 服务器内部错误
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: INTERNAL_ERROR
              message: 服务器内部错误，请稍后重试
```
