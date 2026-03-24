## Why

当前所有 API 错误响应均使用非规范的 `{"error": "message"}` 格式，不符合微软 REST API 错误响应标准（[Microsoft REST API Guidelines](https://github.com/microsoft/api-guidelines/blob/vNext/azure/Guidelines.md#error-response)），导致客户端无法统一解析错误信息，也缺乏结构化的错误码和细节字段。

## What Changes

- **新增** 标准错误响应结构体 `ErrorResponse`，格式遵循微软 REST API 规范：
  ```json
  {
    "error": {
      "code": "NotFound",
      "message": "The requested resource was not found."
    }
  }
  ```
- **修改** `handlePveError` 及所有直接返回 `gin.H{"error": ...}` 的 handler，统一使用新结构体输出
- **修改** Swagger 注释中所有 `{object} map[string]string` 错误响应类型，改为引用新结构体
- **BREAKING** 错误响应体结构从 `{"error": "string"}` 变更为 `{"error": {"code": "...", "message": "..."}}`

## Capabilities

### New Capabilities
- `error-response`: 定义统一的错误响应结构，包含 `code` 和 `message` 字段，遵循微软 REST API Guidelines

### Modified Capabilities

## Impact

- 影响文件：`cmd/handlers.go`
- 所有 API 端点的错误响应结构变更（breaking change）
- Swagger 文档中错误响应类型更新
- 无新增外部依赖
