package pve

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"hyperflow/internal/logger"
)

// Client 封装与 PVE REST API 的通信
type Client struct {
	baseURL     string
	tokenID     string
	tokenSecret string
	httpClient  *http.Client
	logWriter   logger.Logger
}

// PveError 表示 PVE API 返回的错误
type PveError struct {
	StatusCode int
	Message    string
}

func (e *PveError) Error() string {
	return fmt.Sprintf("PVE API error %d: %s", e.StatusCode, e.Message)
}

// NewClient 从环境变量读取配置并创建 PveClient，缺失必要配置时返回错误
func NewClient(logWriter logger.Logger) (*Client, error) {
	host := os.Getenv("PVE_HOST")
	tokenID := os.Getenv("PVE_TOKEN_ID")
	tokenSecret := os.Getenv("PVE_TOKEN_SECRET")

	if host == "" {
		return nil, fmt.Errorf("PVE_HOST environment variable is required")
	}
	if tokenID == "" {
		return nil, fmt.Errorf("PVE_TOKEN_ID environment variable is required")
	}
	if tokenSecret == "" {
		return nil, fmt.Errorf("PVE_TOKEN_SECRET environment variable is required")
	}

	insecure := strings.ToLower(os.Getenv("PVE_INSECURE")) == "true"
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	return &Client{
		baseURL:     strings.TrimRight(host, "/") + "/api2/json",
		tokenID:     tokenID,
		tokenSecret: tokenSecret,
		httpClient:  &http.Client{Transport: transport},
		logWriter:   logWriter,
	}, nil
}

func (c *Client) authHeader() string {
	return fmt.Sprintf("PVEAPIToken=%s=%s", c.tokenID, c.tokenSecret)
}

// do 执行普通 JSON 请求，绑定上游 context，并在返回前记录一次 pve.call 日志。
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (json.RawMessage, error) {
	ctx = contextOrBackground(ctx)
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	startedAt := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		pveErr := &PveError{StatusCode: 502, Message: err.Error()}
		c.logCall(ctx, method, path, pveErr.StatusCode, startedAt, pveErr)
		return nil, pveErr
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logCall(ctx, method, path, resp.StatusCode, startedAt, err)
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Errors map[string]string `json:"errors"`
		}
		msg := string(rawBody)
		if jsonErr := json.Unmarshal(rawBody, &errResp); jsonErr == nil && len(errResp.Errors) > 0 {
			parts := make([]string, 0, len(errResp.Errors))
			for k, v := range errResp.Errors {
				parts = append(parts, k+": "+v)
			}
			msg = strings.Join(parts, "; ")
		}
		pveErr := &PveError{StatusCode: resp.StatusCode, Message: msg}
		c.logCall(ctx, method, path, resp.StatusCode, startedAt, pveErr)
		return nil, pveErr
	}

	c.logCall(ctx, method, path, resp.StatusCode, startedAt, nil)

	// PVE 把数据包在 {"data": ...} 中
	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rawBody, &wrapper); err != nil {
		return rawBody, nil
	}
	return wrapper.Data, nil
}

func (c *Client) Get(ctx context.Context, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) Post(ctx context.Context, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, path, nil)
}

func (c *Client) PostWithBody(ctx context.Context, path string, body io.Reader) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *Client) PutWithBody(ctx context.Context, path string, body io.Reader) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPut, path, body)
}

func (c *Client) Delete(ctx context.Context, path string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodDelete, path, nil)
}

// UploadMultipart 向 PVE 上传 multipart/form-data 文件（用于 Snippets 等存储内容）
func (c *Client) UploadMultipart(ctx context.Context, path string, fields map[string]string, fileField, filename string, fileData []byte) (json.RawMessage, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	fw, err := w.CreateFormFile(fileField, filename)
	if err != nil {
		return nil, err
	}
	_, err = fw.Write(fileData)
	if err != nil {
		return nil, err
	}
	w.Close()

	return c.doMultipart(ctx, path, &buf, w.FormDataContentType())
}

// doMultipart 执行 multipart POST 请求，行为与 do 保持一致，但单独处理 Content-Type。
func (c *Client) doMultipart(ctx context.Context, path string, body io.Reader, contentType string) (json.RawMessage, error) {
	ctx = contextOrBackground(ctx)
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", contentType)

	startedAt := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		pveErr := &PveError{StatusCode: 502, Message: err.Error()}
		c.logCall(ctx, http.MethodPost, path, pveErr.StatusCode, startedAt, pveErr)
		return nil, pveErr
	}
	defer resp.Body.Close()
	rawBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		c.logCall(ctx, http.MethodPost, path, resp.StatusCode, startedAt, readErr)
		return nil, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		pveErr := &PveError{StatusCode: resp.StatusCode, Message: string(rawBody)}
		c.logCall(ctx, http.MethodPost, path, resp.StatusCode, startedAt, pveErr)
		return nil, pveErr
	}
	c.logCall(ctx, http.MethodPost, path, resp.StatusCode, startedAt, nil)
	return rawBody, nil
}

// logCall 统一写入 PVE 出站调用日志，保证成功和失败路径字段结构一致。
func (c *Client) logCall(ctx context.Context, method, path string, statusCode int, startedAt time.Time, err error) {
	if c.logWriter == nil {
		return
	}

	level := "INFO"
	message := ""
	if err != nil {
		level = "ERROR"
		message = err.Error()
	}

	c.logWriter.Log(logger.Entry{
		RequestID:  logger.RequestIDFromContext(ctx),
		Level:      level,
		Event:      "pve.call",
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
		DurationMs: time.Since(startedAt).Milliseconds(),
		Message:    message,
	})
}

// contextOrBackground 兜底 nil context，确保下游 HTTP 请求始终可安全绑定 context。
func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
