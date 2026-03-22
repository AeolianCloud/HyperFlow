## 1. 安装依赖

- [x] 1.1 运行 `go get github.com/swaggo/swag/cmd/swag` 安装 swag CLI
- [x] 1.2 运行 `go get github.com/swaggo/gin-swagger` 添加 gin-swagger 中间件依赖
- [x] 1.3 运行 `go get github.com/swaggo/files` 添加 swagger 静态文件依赖

## 2. 添加全局 API 注释

- [x] 2.1 在 `cmd/main.go` 顶部添加 `@title`、`@version`、`@description`、`@host`、`@BasePath` 等 swaggo 全局注释

## 3. 为 Handler 添加接口注释

- [x] 3.1 为 `GET /api/pve/nodes` 添加 swaggo 注释（summary、tags、produce、success、failure）
- [x] 3.2 为 `GET /api/pve/nodes/:node` 添加 swaggo 注释（含路径参数 node）
- [x] 3.3 为 `GET /api/pve/nodes/:node/vms` 添加 swaggo 注释（含路径参数 node）
- [x] 3.4 为 `GET /api/pve/nodes/:node/vms/:vmid` 添加 swaggo 注释（含路径参数 node、vmid）
- [x] 3.5 为 `POST /api/pve/nodes/:node/vms/:vmid/start` 添加 swaggo 注释
- [x] 3.6 为 `POST /api/pve/nodes/:node/vms/:vmid/stop` 添加 swaggo 注释
- [x] 3.7 为 `DELETE /api/pve/nodes/:node/vms/:vmid` 添加 swaggo 注释
- [x] 3.8 为 `GET /api/pve/storage` 添加 swaggo 注释

## 4. 生成文档并注册路由

- [x] 4.1 在项目根目录运行 `swag init -g cmd/main.go` 生成 `docs/` 包
- [x] 4.2 在 `cmd/main.go` 中 import `docs` 包和 `gin-swagger`，注册 `/swagger/*any` 路由
- [x] 4.3 运行 `go build ./cmd/` 确认编译通过

## 5. 验证

- [x] 5.1 启动服务，访问 `http://localhost:8080/swagger/index.html` 确认 Swagger UI 正常显示
- [x] 5.2 在 Swagger UI 中确认所有 8 个接口均有文档
- [ ] 5.3 在 Swagger UI 中对至少一个接口执行 "Try it out" 测试，确认请求/响应正常
