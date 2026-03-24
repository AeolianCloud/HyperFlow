package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"hyperflow/internal/logger"
)

const requestIDContextKey = "request_id"

// requestLoggingMiddleware 为每个入站请求生成 request_id，并在请求结束后写入 http.request 日志。
func requestLoggingMiddleware(logWriter logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, err := newRequestID()
		if err != nil {
			respondError(c, http.StatusInternalServerError, "InternalServerError", "failed to generate request id")
			c.Abort()
			return
		}

		c.Set(requestIDContextKey, requestID)

		startedAt := time.Now()
		c.Next()

		if logWriter == nil {
			return
		}

		logWriter.Log(logger.Entry{
			RequestID:  requestID,
			Level:      httpLogLevel(c.Writer.Status()),
			Event:      "http.request",
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			DurationMs: time.Since(startedAt).Milliseconds(),
		})
	}
}

// requestContextFromGin 将 gin.Context 中的 request_id 注入标准 context，供下游 service / client 透传。
func requestContextFromGin(c *gin.Context) context.Context {
	requestID := requestIDFromGin(c)
	if requestID == "" {
		return c.Request.Context()
	}
	return context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
}

// requestIDFromGin 读取中间件写入的 request_id。
func requestIDFromGin(c *gin.Context) string {
	return c.GetString(requestIDContextKey)
}

// newRequestID 生成 16 字节随机 request_id，并以 32 位 hex 字符串返回。
func newRequestID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// httpLogLevel 根据 HTTP 状态码选择日志级别，5xx 视为 ERROR，其余记为 INFO。
func httpLogLevel(statusCode int) string {
	if statusCode >= http.StatusInternalServerError {
		return "ERROR"
	}
	return "INFO"
}
