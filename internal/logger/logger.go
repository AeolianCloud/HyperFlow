package logger

import (
	"context"
	"database/sql"
	"fmt"
	"hyperflow/internal/timeutil"
	"os"
	"sync"
	"time"
)

type contextKey string

// RequestIDKey 是在 context.Context 中存储 request_id 的键
const RequestIDKey contextKey = "request_id"

// RequestIDFromContext 从 context 中提取 request_id，不存在时返回空字符串
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(RequestIDKey).(string); ok {
		return v
	}
	return ""
}

// Entry 表示一条结构化日志记录
type Entry struct {
	RequestID   string
	Timestamp   time.Time
	Level       string // INFO / WARN / ERROR
	Event       string // http.request / pve.call / ws.connect / ws.disconnect / operation.change
	Method      string
	Path        string
	StatusCode  int
	DurationMs  int64
	OperationID string
	Node        string
	Message     string
}

// Logger 定义日志写入接口
type Logger interface {
	Log(e Entry)
	Shutdown(ctx context.Context)
}

// MySQLLogger 将日志异步写入 MySQL logs 表
type MySQLLogger struct {
	db           *sql.DB
	ch           chan Entry
	done         chan struct{}
	mu           sync.RWMutex
	shutdownOnce sync.Once
	closed       bool
}

// NewMySQLLogger 创建并启动异步日志写入器
func NewMySQLLogger(db *sql.DB) *MySQLLogger {
	l := &MySQLLogger{
		db:   db,
		ch:   make(chan Entry, 1000),
		done: make(chan struct{}),
	}
	go l.run()
	return l
}

// CreateTable 确保 logs 表存在
func (l *MySQLLogger) CreateTable() error {
	_, err := l.db.Exec(`CREATE TABLE IF NOT EXISTS logs (
		id            BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
		request_id    VARCHAR(32)  NOT NULL,
		timestamp     DATETIME(3)  NOT NULL,
		level         VARCHAR(10)  NOT NULL,
		event         VARCHAR(100) NOT NULL,
		method        VARCHAR(10)  NULL,
		path          VARCHAR(500) NULL,
		status_code   INT          NULL,
		duration_ms   BIGINT       NULL,
		operation_id  VARCHAR(32)  NULL,
		node          VARCHAR(100) NULL,
		message       TEXT         NULL,
		INDEX idx_request_id (request_id),
		INDEX idx_timestamp  (timestamp)
	)`)
	return err
}

// Log 将日志条目投入异步 channel；channel 满时丢弃并向 stderr 告警
func (l *MySQLLogger) Log(e Entry) {
	if e.Timestamp.IsZero() {
		e.Timestamp = timeutil.NowShanghai()
	} else {
		e.Timestamp = timeutil.InShanghai(e.Timestamp)
	}

	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.closed {
		return
	}

	select {
	case l.ch <- e:
	default:
		fmt.Fprintf(os.Stderr, "logger: channel full, dropping log entry event=%s request_id=%s\n", e.Event, e.RequestID)
	}
}

func (l *MySQLLogger) run() {
	defer close(l.done)
	for e := range l.ch {
		// 单后台 goroutine 串行消费，避免并发写库带来的额外复杂度。
		l.write(e)
	}
}

// write 将单条日志持久化到 logs 表；写库失败只打 stderr，不反向影响业务请求。
func (l *MySQLLogger) write(e Entry) {
	_, err := l.db.Exec(
		`INSERT INTO logs (
			request_id, timestamp, level, event, method, path,
			status_code, duration_ms, operation_id, node, message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.RequestID,
		e.Timestamp,
		e.Level,
		e.Event,
		nullableString(e.Method),
		nullableString(e.Path),
		nullableStatusCode(e),
		nullableDurationMs(e),
		nullableString(e.OperationID),
		nullableString(e.Node),
		nullableString(e.Message),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: failed to persist log entry event=%s request_id=%s err=%v\n", e.Event, e.RequestID, err)
	}
}

// Shutdown 关闭 channel 并等待所有条目写入完成或 ctx 超时
func (l *MySQLLogger) Shutdown(ctx context.Context) {
	l.shutdownOnce.Do(func() {
		l.mu.Lock()
		l.closed = true
		close(l.ch)
		l.mu.Unlock()
	})

	select {
	case <-l.done:
	case <-ctx.Done():
	}
}

// nullableString 将空字符串映射为 SQL NULL，保持日志表的可空列语义。
func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

// nullableStatusCode 仅在包含 HTTP/PVE 响应码的事件上写入 status_code。
func nullableStatusCode(e Entry) any {
	if e.Event != "http.request" && e.Event != "pve.call" {
		return nil
	}
	return e.StatusCode
}

// nullableDurationMs 仅在包含耗时语义的事件上写入 duration_ms。
func nullableDurationMs(e Entry) any {
	if e.Event != "http.request" && e.Event != "pve.call" {
		return nil
	}
	return e.DurationMs
}
