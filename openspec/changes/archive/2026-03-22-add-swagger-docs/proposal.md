## Why

项目目前没有接口文档，前端开发和测试人员无法直观了解 API 的参数、响应格式和错误码，调试效率低。引入 Swagger UI 可提供交互式接口文档，支持直接在浏览器中发起请求测试。

## What Changes

- 引入 `swaggo/swag` 工具链，为所有已有 API handler 添加注释
- 集成 `gin-swagger` 中间件，在 `/swagger/*any` 路由暴露 Swagger UI
- 为以下接口组生成文档：Nodes、VMs、Storage
- 添加全局 API 信息（title、version、base URL）

## Capabilities

### New Capabilities
- `swagger-ui`: 通过 `/swagger/index.html` 访问的交互式 API 文档页面，由 swaggo/gin-swagger 驱动

### Modified Capabilities

## Impact

- `cmd/main.go`：注册 swagger 路由
- `cmd/handlers.go`：为每个 handler 添加 swaggo 注释
- `go.mod`/`go.sum`：新增 `github.com/swaggo/swag`、`github.com/swaggo/gin-swagger`、`github.com/swaggo/files` 依赖
- 构建步骤：需先运行 `swag init` 生成 `docs/` 包
