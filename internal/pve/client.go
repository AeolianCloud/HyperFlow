package pve

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Client 封装与 PVE REST API 的通信
type Client struct {
	baseURL    string
	tokenID    string
	tokenSecret string
	httpClient *http.Client
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
func NewClient() (*Client, error) {
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
	}, nil
}

func (c *Client) authHeader() string {
	return fmt.Sprintf("PVEAPIToken=%s=%s", c.tokenID, c.tokenSecret)
}

func (c *Client) do(method, path string, body io.Reader) (json.RawMessage, error) {
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &PveError{StatusCode: 502, Message: err.Error()}
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
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
		return nil, &PveError{StatusCode: resp.StatusCode, Message: msg}
	}

	// PVE 把数据包在 {"data": ...} 中
	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(rawBody, &wrapper); err != nil {
		return rawBody, nil
	}
	return wrapper.Data, nil
}

func (c *Client) Get(path string) (json.RawMessage, error) {
	return c.do(http.MethodGet, path, nil)
}

func (c *Client) Post(path string) (json.RawMessage, error) {
	return c.do(http.MethodPost, path, nil)
}

func (c *Client) PostWithBody(path string, body io.Reader) (json.RawMessage, error) {
	return c.do(http.MethodPost, path, body)
}

func (c *Client) PutWithBody(path string, body io.Reader) (json.RawMessage, error) {
	return c.do(http.MethodPut, path, body)
}

func (c *Client) Delete(path string) (json.RawMessage, error) {
	return c.do(http.MethodDelete, path, nil)
}
