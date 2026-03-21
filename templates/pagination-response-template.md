# 分页响应模板

## 标准分页格式

所有返回列表的API端点使用统一的分页格式：

```json
{
  "data": [...],
  "total": 156,
  "offset": 0,
  "limit": 20
}
```

## 字段说明

- `data`: 当前页的数据数组
- `total`: 总记录数
- `offset`: 当前偏移量
- `limit`: 每页数量

## 分页参数

### 查询参数

- `offset`: 偏移量，默认0
- `limit`: 每页数量，默认20，最大100

### 示例请求

```http
GET /api/v1/users?offset=0&limit=20
GET /api/v1/users?offset=20&limit=20
GET /api/v1/users?offset=40&limit=20
```

## Go语言实现

### 分页结构定义

```go
package models

// PaginatedResponse 分页响应
type PaginatedResponse struct {
    Data   interface{} `json:"data"`
    Total  int         `json:"total"`
    Offset int         `json:"offset"`
    Limit  int         `json:"limit"`
}

// PaginationParams 分页参数
type PaginationParams struct {
    Offset int `form:"offset" binding:"min=0"`
    Limit  int `form:"limit" binding:"min=1,max=100"`
}

// GetOffset 获取偏移量，提供默认值
func (p *PaginationParams) GetOffset() int {
    if p.Offset < 0 {
        return 0
    }
    return p.Offset
}

// GetLimit 获取每页数量，提供默认值
func (p *PaginationParams) GetLimit() int {
    if p.Limit <= 0 {
        return 20
    }
    if p.Limit > 100 {
        return 100
    }
    return p.Limit
}
```

### 分页辅助函数

```go
package utils

import "myapp/models"

// NewPaginatedResponse 创建分页响应
func NewPaginatedResponse(data interface{}, total, offset, limit int) models.PaginatedResponse {
    return models.PaginatedResponse{
        Data:   data,
        Total:  total,
        Offset: offset,
        Limit:  limit,
    }
}

// CalculateTotalPages 计算总页数
func CalculateTotalPages(total, limit int) int {
    if limit <= 0 {
        return 0
    }
    return (total + limit - 1) / limit
}

// CalculateCurrentPage 计算当前页码（从1开始）
func CalculateCurrentPage(offset, limit int) int {
    if limit <= 0 {
        return 1
    }
    return (offset / limit) + 1
}
```

### 使用示例

```go
package handlers

import (
    "myapp/models"
    "myapp/utils"

    "github.com/gin-gonic/gin"
)

// GetUsers 获取用户列表
func GetUsers(c *gin.Context) {
    // 解析分页参数
    var params models.PaginationParams
    if err := c.ShouldBindQuery(&params); err != nil {
        c.JSON(400, gin.H{"error": "无效的分页参数"})
        return
    }

    offset := params.GetOffset()
    limit := params.GetLimit()

    // 查询数据库
    users, total, err := getUsersFromDB(offset, limit)
    if err != nil {
        c.JSON(500, gin.H{"error": "查询失败"})
        return
    }

    // 返回分页响应
    response := utils.NewPaginatedResponse(users, total, offset, limit)
    c.JSON(200, response)
}

// getUsersFromDB 从数据库查询用户（示例）
func getUsersFromDB(offset, limit int) ([]models.User, int, error) {
    var users []models.User
    var total int64

    // 查询总数
    db.Model(&models.User{}).Count(&total)

    // 查询数据
    err := db.Offset(offset).Limit(limit).Find(&users).Error

    return users, int(total), err
}
```

### 带过滤和排序的分页

```go
package handlers

import (
    "myapp/models"
    "myapp/utils"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// ListUsersParams 用户列表参数
type ListUsersParams struct {
    models.PaginationParams
    Status string `form:"status"`
    Sort   string `form:"sort" binding:"omitempty"`
}

// GetUsersWithFilters 获取用户列表（带过滤和排序）
func GetUsersWithFilters(c *gin.Context) {
    var params ListUsersParams
    if err := c.ShouldBindQuery(&params); err != nil {
        c.JSON(400, gin.H{"error": "无效的参数"})
        return
    }

    offset := params.GetOffset()
    limit := params.GetLimit()

    // 构建查询
    query := db.Model(&models.User{})

    // 应用过滤
    if params.Status != "" {
        query = query.Where("status = ?", params.Status)
    }

    // 应用排序
    if params.Sort != "" {
        orderClause := parseSortParam(params.Sort)
        query = query.Order(orderClause)
    } else {
        query = query.Order("created_at DESC") // 默认排序
    }

    // 查询总数
    var total int64
    query.Count(&total)

    // 查询数据
    var users []models.User
    err := query.Offset(offset).Limit(limit).Find(&users).Error
    if err != nil {
        c.JSON(500, gin.H{"error": "查询失败"})
        return
    }

    // 返回分页响应
    response := utils.NewPaginatedResponse(users, int(total), offset, limit)
    c.JSON(200, response)
}

// parseSortParam 解析排序参数
func parseSortParam(sort string) string {
    if sort == "" {
        return "created_at DESC"
    }

    // 处理降序（-前缀）
    if sort[0] == '-' {
        return sort[1:] + " DESC"
    }

    return sort + " ASC"
}
```

## OpenAPI定义

```yaml
components:
  schemas:
    PaginatedUserResponse:
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

  parameters:
    offsetParam:
      name: offset
      in: query
      description: 偏移量，从0开始
      schema:
        type: integer
        default: 0
        minimum: 0
      example: 0

    limitParam:
      name: limit
      in: query
      description: 每页数量，最大100
      schema:
        type: integer
        default: 20
        minimum: 1
        maximum: 100
      example: 20

paths:
  /users:
    get:
      summary: 获取用户列表
      parameters:
        - $ref: '#/components/parameters/offsetParam'
        - $ref: '#/components/parameters/limitParam'
        - name: status
          in: query
          description: 过滤用户状态
          schema:
            type: string
            enum: [active, inactive, suspended]
        - name: sort
          in: query
          description: 排序字段（使用-前缀表示降序）
          schema:
            type: string
          example: "-createdAt"
      responses:
        '200':
          description: 成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PaginatedUserResponse'
              example:
                data:
                  - id: 1
                    name: 张三
                    email: zhangsan@example.com
                    status: active
                  - id: 2
                    name: 李四
                    email: lisi@example.com
                    status: active
                total: 156
                offset: 0
                limit: 20
```

## 响应示例

### 第一页

```http
GET /api/v1/users?offset=0&limit=20
```

```json
{
  "data": [
    {
      "id": 1,
      "name": "张三",
      "email": "zhangsan@example.com",
      "status": "active",
      "createdAt": "2024-03-19T10:00:00Z"
    },
    {
      "id": 2,
      "name": "李四",
      "email": "lisi@example.com",
      "status": "active",
      "createdAt": "2024-03-18T15:00:00Z"
    }
  ],
  "total": 156,
  "offset": 0,
  "limit": 20
}
```

### 第二页

```http
GET /api/v1/users?offset=20&limit=20
```

```json
{
  "data": [
    {
      "id": 21,
      "name": "用户21",
      "email": "user21@example.com",
      "status": "active",
      "createdAt": "2024-03-17T10:00:00Z"
    }
  ],
  "total": 156,
  "offset": 20,
  "limit": 20
}
```

### 空结果

```http
GET /api/v1/users?offset=200&limit=20
```

```json
{
  "data": [],
  "total": 156,
  "offset": 200,
  "limit": 20
}
```

### 带过滤

```http
GET /api/v1/users?status=active&offset=0&limit=20
```

```json
{
  "data": [
    {
      "id": 1,
      "name": "张三",
      "email": "zhangsan@example.com",
      "status": "active",
      "createdAt": "2024-03-19T10:00:00Z"
    }
  ],
  "total": 120,
  "offset": 0,
  "limit": 20
}
```

## 客户端使用示例

### Go客户端

```go
package client

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type PaginatedResponse struct {
    Data   []map[string]interface{} `json:"data"`
    Total  int                      `json:"total"`
    Offset int                      `json:"offset"`
    Limit  int                      `json:"limit"`
}

func GetUsers(offset, limit int) (*PaginatedResponse, error) {
    url := fmt.Sprintf("https://api.example.com/v1/users?offset=%d&limit=%d", offset, limit)

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result PaginatedResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

// 遍历所有页
func GetAllUsers() ([]map[string]interface{}, error) {
    var allUsers []map[string]interface{}
    offset := 0
    limit := 100

    for {
        result, err := GetUsers(offset, limit)
        if err != nil {
            return nil, err
        }

        allUsers = append(allUsers, result.Data...)

        if offset+limit >= result.Total {
            break
        }

        offset += limit
    }

    return allUsers, nil
}
```

### JavaScript客户端

```javascript
async function getUsers(offset = 0, limit = 20) {
  const response = await fetch(
    `https://api.example.com/v1/users?offset=${offset}&limit=${limit}`
  );

  return await response.json();
}

// 遍历所有页
async function getAllUsers() {
  const allUsers = [];
  let offset = 0;
  const limit = 100;

  while (true) {
    const result = await getUsers(offset, limit);
    allUsers.push(...result.data);

    if (offset + limit >= result.total) {
      break;
    }

    offset += limit;
  }

  return allUsers;
}

// 计算分页信息
function getPaginationInfo(response) {
  const { total, offset, limit } = response;
  const currentPage = Math.floor(offset / limit) + 1;
  const totalPages = Math.ceil(total / limit);

  return {
    currentPage,
    totalPages,
    hasNextPage: offset + limit < total,
    hasPrevPage: offset > 0,
  };
}
```

## 最佳实践

1. **一致的响应格式**：所有分页端点使用相同的响应结构
2. **合理的默认值**：offset默认0，limit默认20
3. **限制最大值**：limit最大值100，防止过大请求
4. **返回总数**：始终返回total字段，便于客户端计算总页数
5. **空结果处理**：offset超出范围时返回空数组，不返回错误
6. **性能优化**：为常用的过滤和排序字段创建索引
