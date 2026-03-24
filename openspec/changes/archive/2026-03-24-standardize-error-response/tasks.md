## 1. 定义错误响应结构体

- [x] 1.1 在 `cmd/handlers.go` 中新增 `ErrorDetail` 和 `ErrorResponse` 结构体
- [x] 1.2 新增 `respondError(c *gin.Context, status int, code, message string)` 辅助函数

## 2. 重构错误输出

- [x] 2.1 重构 `handlePveError`，使用 `respondError` 替换所有 `gin.H{"error": ...}` 输出，并映射各状态码对应的 `code` 字符串
- [x] 2.2 替换 `createVm` handler 中直接调用 `c.JSON(..., gin.H{"error": ...})` 的两处为 `respondError`

## 3. 更新 Swagger 注释

- [x] 3.1 将所有 `@Failure` 注释中的 `map[string]string` 替换为 `ErrorResponse`
