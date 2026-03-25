package operations

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"hyperflow/internal/timeutil"
)

// Operation 表示一个异步操作记录，遵循微软 LRO 规范
type Operation struct {
	ID               string    `json:"id"`
	Status           string    `json:"status"` // Running / Succeeded / Failed
	PVENode          string    `json:"-"`
	PVEUpid          string    `json:"-"`
	ResourceLocation string    `json:"resourceLocation,omitempty"`
	ErrorCode        string    `json:"-"`
	ErrorMessage     string    `json:"-"`
	CreatorRequestID string    `json:"-"`
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
}

// Store 定义 Operation 持久化接口
type Store interface {
	CreateTable() error
	Insert(op *Operation) error
	GetByID(id string) (*Operation, error)
	ListRunning(limit int) ([]*Operation, error)
	CompleteOperation(op *Operation, event *OutboxEvent) (bool, error)
	ListPendingEvents(limit int) ([]*OutboxEvent, error)
	MarkEventPublished(id string) error
	MarkEventPublishFailed(id, lastError string) error
}

type mysqlStore struct {
	db *sql.DB
}

// NewMySQLStore 创建基于 MySQL 的 Store 实现
func NewMySQLStore(db *sql.DB) Store {
	return &mysqlStore{db: db}
}

// CreateTable 确保 operations 表存在
func (s *mysqlStore) CreateTable() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS operations (
		id                 VARCHAR(32)  NOT NULL PRIMARY KEY,
		status             VARCHAR(16)  NOT NULL,
		pve_node           VARCHAR(128) NOT NULL,
		pve_upid           VARCHAR(256) NOT NULL,
		resource_location  VARCHAR(256) NOT NULL DEFAULT '',
		error_code         VARCHAR(64)  NOT NULL DEFAULT '',
		error_message      TEXT         NOT NULL,
		creator_request_id VARCHAR(32)  NULL,
		created_at         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS operation_events_outbox (
		id            VARCHAR(32)  NOT NULL PRIMARY KEY,
		operation_id  VARCHAR(32)  NOT NULL,
		request_id    VARCHAR(32)  NULL,
		topic         VARCHAR(255) NOT NULL,
		payload       TEXT         NOT NULL,
		attempts      INT          NOT NULL DEFAULT 0,
		last_error    TEXT         NOT NULL,
		published_at  DATETIME     NULL,
		created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_operation_events_outbox_pending (published_at, created_at),
		INDEX idx_operation_events_outbox_operation_id (operation_id)
	)`)
	return err
}

// Insert 写入一条新 operation 记录
func (s *mysqlStore) Insert(op *Operation) error {
	now := timeutil.NowShanghai()
	_, err := s.db.Exec(
		`INSERT INTO operations (
			id, status, pve_node, pve_upid, resource_location,
			error_code, error_message, creator_request_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		op.ID, op.Status, op.PVENode, op.PVEUpid, op.ResourceLocation, op.ErrorCode, op.ErrorMessage, nullableString(op.CreatorRequestID), now, now,
	)
	return err
}

// GetByID 按 ID 查询 operation 记录，不存在时返回 nil, nil
func (s *mysqlStore) GetByID(id string) (*Operation, error) {
	row := s.db.QueryRow(
		`SELECT id, status, pve_node, pve_upid, resource_location, error_code, error_message, creator_request_id, created_at, updated_at FROM operations WHERE id = ?`, id,
	)
	op := &Operation{}
	var creatorRequestID sql.NullString
	err := row.Scan(&op.ID, &op.Status, &op.PVENode, &op.PVEUpid, &op.ResourceLocation, &op.ErrorCode, &op.ErrorMessage, &creatorRequestID, &op.CreatedAt, &op.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	op.CreatorRequestID = creatorRequestID.String
	op.CreatedAt = timeutil.InShanghai(op.CreatedAt)
	op.UpdatedAt = timeutil.InShanghai(op.UpdatedAt)
	return op, nil
}

// ListRunning 返回仍处于 Running 状态的 operation 列表。
func (s *mysqlStore) ListRunning(limit int) ([]*Operation, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.Query(
		`SELECT id, status, pve_node, pve_upid, resource_location, error_code, error_message, creator_request_id, created_at, updated_at
		 FROM operations
		 WHERE status = 'Running'
		 ORDER BY created_at ASC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ops []*Operation
	for rows.Next() {
		op, err := scanOperation(rows)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ops, nil
}

// CompleteOperation 仅在 operation 仍为 Running 时将其推进到终态，并同步写入 outbox 事件。
func (s *mysqlStore) CompleteOperation(op *Operation, event *OutboxEvent) (bool, error) {
	if op == nil {
		return false, nil
	}
	if event == nil {
		return false, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	now := timeutil.NowShanghai()
	result, err := tx.Exec(
		`UPDATE operations
		 SET status = ?, error_code = ?, error_message = ?, updated_at = ?
		 WHERE id = ? AND status = 'Running'`,
		op.Status, op.ErrorCode, op.ErrorMessage, now, op.ID,
	)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowsAffected == 0 {
		return false, nil
	}

	_, err = tx.Exec(
		`INSERT INTO operation_events_outbox (
			id, operation_id, request_id, topic, payload, attempts, last_error, published_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID,
		event.OperationID,
		nullableString(event.RequestID),
		event.Topic,
		string(event.Payload),
		event.Attempts,
		event.LastError,
		nullableTime(event.PublishedAt),
		now,
		now,
	)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	op.UpdatedAt = now
	event.CreatedAt = now
	event.UpdatedAt = now
	return true, nil
}

// ListPendingEvents 返回尚未成功发布到 Kafka 的 outbox 事件。
func (s *mysqlStore) ListPendingEvents(limit int) ([]*OutboxEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.Query(
		`SELECT id, operation_id, request_id, topic, payload, attempts, last_error, published_at, created_at, updated_at
		 FROM operation_events_outbox
		 WHERE published_at IS NULL
		 ORDER BY created_at ASC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*OutboxEvent
	for rows.Next() {
		event, err := scanOutboxEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// MarkEventPublished 将 outbox 事件标记为已发布。
func (s *mysqlStore) MarkEventPublished(id string) error {
	now := timeutil.NowShanghai()
	_, err := s.db.Exec(
		`UPDATE operation_events_outbox
		 SET published_at = ?, last_error = '', updated_at = ?
		 WHERE id = ?`,
		now, now, id,
	)
	return err
}

// MarkEventPublishFailed 记录一次 outbox 事件发布失败。
func (s *mysqlStore) MarkEventPublishFailed(id, lastError string) error {
	now := timeutil.NowShanghai()
	_, err := s.db.Exec(
		`UPDATE operation_events_outbox
		 SET attempts = attempts + 1, last_error = ?, updated_at = ?
		 WHERE id = ?`,
		lastError, now, id,
	)
	return err
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func nullableTime(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}

func scanOperation(scanner interface{ Scan(dest ...any) error }) (*Operation, error) {
	op := &Operation{}
	var creatorRequestID sql.NullString
	err := scanner.Scan(
		&op.ID,
		&op.Status,
		&op.PVENode,
		&op.PVEUpid,
		&op.ResourceLocation,
		&op.ErrorCode,
		&op.ErrorMessage,
		&creatorRequestID,
		&op.CreatedAt,
		&op.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	op.CreatorRequestID = creatorRequestID.String
	op.CreatedAt = timeutil.InShanghai(op.CreatedAt)
	op.UpdatedAt = timeutil.InShanghai(op.UpdatedAt)
	return op, nil
}

func scanOutboxEvent(scanner interface{ Scan(dest ...any) error }) (*OutboxEvent, error) {
	event := &OutboxEvent{}
	var requestID sql.NullString
	var publishedAt sql.NullTime
	var payload string
	err := scanner.Scan(
		&event.ID,
		&event.OperationID,
		&requestID,
		&event.Topic,
		&payload,
		&event.Attempts,
		&event.LastError,
		&publishedAt,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	event.RequestID = requestID.String
	event.Payload = []byte(payload)
	event.CreatedAt = timeutil.InShanghai(event.CreatedAt)
	event.UpdatedAt = timeutil.InShanghai(event.UpdatedAt)
	if publishedAt.Valid {
		ts := timeutil.InShanghai(publishedAt.Time)
		event.PublishedAt = &ts
	}
	return event, nil
}
