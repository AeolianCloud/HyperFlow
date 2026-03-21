# 示例API端点 - 用户资源CRUD操作

本示例展示完整的用户资源CRUD操作，包含所有API规范要点。

## Go语言实现（使用Gin框架）

### 完整代码

```go
package main

import (
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    _ "myapp/docs"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           用户管理API
// @version         1.0.0
// @description     用户管理API示例，展示RESTful API最佳实践
// @contact.name   API支持
// @contact.email  support@example.com
// @host      localhost:8080
// @BasePath  /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
    r := gin.Default()

    // Swagger文档
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    // API路由
    v1 := r.Group("/api/v1")
    {
        users := v1.Group("/users")
        {
            users.GET("", GetUsers)
            users.POST("", CreateUser)
            users.GET("/:id", GetUser)
            users.PATCH("/:id", UpdateUser)
            users.DELETE("/:id", DeleteUser)
        }
    }

    r.Run(":8080")
}

// User 用户模型
type User struct {
    ID        int       `json:"id" example:"123"`
    Name      string    `json:"name" example:"张三"`
    Email     string    `json:"email" example:"zhangsan@example.com"`
    Phone     string    `json:"phone,omitempty" example:"13800138000"`
    Status    string    `json:"status" example:"active" enums:"active,inactive,suspended"`
    Role      string    `json:"role" example:"user" enums:"admin,user,guest"`
    CreatedAt time.Time `json:"createdAt" example:"2024-03-19T10:00:00Z"`
    UpdatedAt time.Time `json:"updatedAt" example:"2024-03-19T10:00:00Z"`
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

// GetUsers godoc
// @Summary      获取用户列表
// @Description  返回分页的用户列表，支持过滤、排序和字段选择
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        offset  query     int     false  "偏移量"  default(0)  minimum(0)
// @Param        limit   query     int     false  "每页数量"  default(20)  minimum(1)  maximum(100)
// @Param        status  query     string  false  "用户状态"  Enums(active, inactive, suspended)
// @Param        sort    query     string  false  "排序字段"  default(-createdAt)
// @Param        fields  query     string  false  "返回字段"  example(id,name,email)
// @Success      200  {object}  UserListResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      429  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /users [get]
func GetUsers(c *gin.Context) {
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    status := c.Query("status")

    // 模拟数据
    users := []User{
        {ID: 1, Name: "张三", Email: "zhangsan@example.com", Status: "active", Role: "user", CreatedAt: time.Now()},
        {ID: 2, Name: "李四", Email: "lisi@example.com", Status: "active", Role: "user", CreatedAt: time.Now()},
    }

    c.JSON(http.StatusOK, UserListResponse{
        Data:   users,
        Total:  156,
        Offset: offset,
        Limit:  limit,
    })
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
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": gin.H{
                "code":    "VALIDATION_ERROR",
                "message": "请求参数验证失败",
            },
        })
        return
    }

    user := User{
        ID:        123,
        Name:      req.Name,
        Email:     req.Email,
        Phone:     req.Phone,
        Status:    "active",
        Role:      "user",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    c.Header("Location", "/api/v1/users/123")
    c.JSON(http.StatusCreated, user)
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
    id := c.Param("id")

    user := User{
        ID:        123,
        Name:      "张三",
        Email:     "zhangsan@example.com",
        Phone:     "13800138000",
        Status:    "active",
        Role:      "user",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    c.JSON(http.StatusOK, user)
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
    id := c.Param("id")
    var req UpdateUserRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": gin.H{
                "code":    "VALIDATION_ERROR",
                "message": "请求参数验证失败",
            },
        })
        return
    }

    user := User{
        ID:        123,
        Name:      req.Name,
        Email:     req.Email,
        Phone:     req.Phone,
        Status:    "active",
        Role:      "user",
        UpdatedAt: time.Now(),
    }

    c.JSON(http.StatusOK, user)
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
    id := c.Param("id")
    c.Status(http.StatusNoContent)
}

// ErrorResponse 错误响应
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
    Code    string       `json:"code" example:"VALIDATION_ERROR"`
    Message string       `json:"message" example:"请求参数验证失败"`
    Details []FieldError `json:"details,omitempty"`
}

// FieldError 字段错误
type FieldError struct {
    Field   string `json:"field" example:"email"`
    Message string `json:"message" example:"邮箱格式不正确"`
}
```

### 运行示例

```bash
# 安装依赖
go get -u github.com/gin-gonic/gin
go get -u github.com/swaggo/swag/cmd/swag
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files

# 生成文档
swag init

# 运行服务
go run main.go

# 访问文档
# http://localhost:8080/swagger/index.html
```

### 测试API

```bash
# 获取用户列表
curl http://localhost:8080/api/v1/users

# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"张三","email":"zhangsan@example.com"}'

# 获取单个用户
curl http://localhost:8080/api/v1/users/123

# 更新用户
curl -X PATCH http://localhost:8080/api/v1/users/123 \
  -H "Content-Type: application/json" \
  -d '{"name":"新名字"}'

# 删除用户
curl -X DELETE http://localhost:8080/api/v1/users/123
```
