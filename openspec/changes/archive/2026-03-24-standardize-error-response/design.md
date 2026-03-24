## Context

当前所有 API 错误响应使用 `{"error": "message string"}` 的扁平格式，所有 handler 直接调用 `gin.H{"error": ...}` 或通过 `handlePveError` 输出。微软 REST API Guidelines 要求错误响应结构为嵌套对象，包含 `code`（机器可读的错误码）和 `message`（人类可读说明）字段。

## Goals / Non-Goals

**Goals:**
- 统一所有错误响应为微软标准格式：`{"error": {"code": "...", "message": "..."}}`
- 定义可复用的 `ErrorResponse` / `ErrorDetail` 结构体用于 Swagger 文档引用
- 更新 `handlePveError` 和所有直接输出错误的 handler

**Non-Goals:**
- 引入多层嵌套的 `details`/`innererror` 字段（微软规范的可选扩展）
- 国际化错误消息
- 修改成功响应结构

## Decisions

### 决策：在 `cmd/handlers.go` 中定义错误结构体

```go
type ErrorDetail struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}
```

`code` 字段使用 PascalCase 英文字符串（如 `NotFound`、`Conflict`、`BadRequest`、`InternalServerError`、`BadGateway`），与微软规范一致。

`handlePveError` 根据 HTTP 状态码映射 code 值；直接 `c.JSON` 的错误调用点改为调用统一辅助函数 `respondError(c, statusCode, code, message)`。

**备选：使用独立 errors 包** — 项目规模小，handler 层内定义即可，无需新包。

## Risks / Trade-offs

- [Risk] Breaking change，已有客户端解析 `error` 为字符串会失败 → 在 proposal 中已标注 BREAKING，需客户端同步更新
- [Risk] PVE 返回的错误 message 中可能含有复杂格式 → 直接透传 message 字段，不做额外处理

## Migration Plan

1. 在 `handlers.go` 中新增 `ErrorDetail`、`ErrorResponse` 结构体及 `respondError` 辅助函数
2. 重构 `handlePveError` 使用新结构体
3. 替换所有 `c.JSON(..., gin.H{"error": ...})` 调用为 `respondError`
4. 更新所有 Swagger `@Failure` 注释引用新结构体
