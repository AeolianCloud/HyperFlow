## 1. 依赖与基础设施

- [x] 1.1 执行 `go get github.com/gorilla/websocket` 并更新 go.mod / go.sum

## 2. 删除原 REST 接口

- [x] 2.1 删除 `cmd/handlers.go` 中的 `getOperation` handler 及其 swag 注释
- [x] 2.2 删除 `registerOperationsRoutes` 中 `rg.GET("/:id", getOperation)` 路由注册

## 3. WebSocket Handler

- [x] 3.1 在 `cmd/handlers.go` 中定义 `var upgrader websocket.Upgrader`（允许所有来源）
- [x] 3.2 实现 `watchOperation` handler：升级前校验操作存在性，不存在返回 404
- [x] 3.3 实现连接建立后的推送循环：1s ticker 调用 `operationsSvcGlobal.GetOperation`，推送 JSON 状态消息
- [x] 3.4 操作进入终态（Succeeded / Failed）时推送最终消息并以 code 1000 关闭连接
- [x] 3.5 客户端断开时（write 返回错误）退出 goroutine，不泄漏资源

## 4. 路由注册

- [x] 4.1 在 `registerOperationsRoutes` 中新增 `rg.GET("/:id/watch", watchOperation)`

## 5. 文档

- [x] 5.1 为 `watchOperation` 添加完整 swag 注释（Tags: operations，描述 WebSocket 升级行为及响应格式）
- [x] 5.2 执行 `swag init -g cmd/main.go` 重新生成 `docs/`
