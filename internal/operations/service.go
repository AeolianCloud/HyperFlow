package operations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"hyperflow/internal/logger"
)

// TaskStatusQuerier 查询底层 PVE 任务状态的接口，由 pve.VmsService 实现
type TaskStatusQuerier interface {
	GetTaskStatus(ctx context.Context, node, upid string) (status string, exitStatus string, err error)
}

// Service 提供 LRO operation 的创建与查询逻辑
type Service struct {
	store      Store
	querier    TaskStatusQuerier
	logWriter  logger.Logger
	eventTopic string
}

// NewService 创建 Service
func NewService(store Store, querier TaskStatusQuerier, logWriter logger.Logger, eventTopic string) *Service {
	return &Service{store: store, querier: querier, logWriter: logWriter, eventTopic: eventTopic}
}

// CreateOperation 生成随机 ID，写入 DB，返回新建的 Operation
func (s *Service) CreateOperation(ctx context.Context, node, upid, resourceLocation string) (*Operation, error) {
	id, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate operation id: %w", err)
	}
	op := &Operation{
		ID:               id,
		Status:           "Running",
		PVENode:          node,
		PVEUpid:          upid,
		ResourceLocation: resourceLocation,
		CreatorRequestID: logger.RequestIDFromContext(contextOrBackground(ctx)),
	}
	if err := s.store.Insert(op); err != nil {
		return nil, fmt.Errorf("failed to persist operation: %w", err)
	}
	return op, nil
}

// GetOperation 返回持久化的 operation 状态；返回 nil, nil 表示 ID 不存在。
func (s *Service) GetOperation(ctx context.Context, id string) (*Operation, error) {
	op, err := s.store.GetByID(id)
	if err != nil {
		return nil, err
	}
	if op == nil || op.Status != "Running" || s.querier == nil {
		return op, nil
	}

	// 读时兜底推进一次状态，避免后台 reconciler 临时失效时长期停留在 Running。
	if err := s.reconcileOperation(contextOrBackground(ctx), op); err != nil {
		return nil, err
	}

	return s.store.GetByID(id)
}

// ReconcileRunningOperations 扫描 Running operations，并在底层任务进入终态时推进状态并写入 outbox。
func (s *Service) ReconcileRunningOperations(ctx context.Context, limit int) error {
	if s.querier == nil {
		return fmt.Errorf("task status querier is not configured")
	}

	ops, err := s.store.ListRunning(limit)
	if err != nil {
		return fmt.Errorf("failed to list running operations: %w", err)
	}

	var firstErr error
	for _, op := range ops {
		if err := s.reconcileOperation(contextOrBackground(ctx), op); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (s *Service) reconcileOperation(ctx context.Context, op *Operation) error {
	pveStatus, pveExitStatus, err := s.querier.GetTaskStatus(ctx, op.PVENode, op.PVEUpid)
	if err != nil {
		return nil
	}
	if pveStatus != "stopped" {
		return nil
	}

	terminal := *op
	if pveExitStatus == "OK" {
		terminal.Status = "Succeeded"
		terminal.ErrorCode = ""
		terminal.ErrorMessage = ""
	} else {
		terminal.Status = "Failed"
		terminal.ErrorCode = "TaskFailed"
		terminal.ErrorMessage = pveExitStatus
	}

	event, err := NewOutboxEvent(&terminal, s.eventTopic)
	if err != nil {
		return fmt.Errorf("failed to build outbox event for operation %s: %w", op.ID, err)
	}

	applied, err := s.store.CompleteOperation(&terminal, event)
	if err != nil {
		return fmt.Errorf("failed to complete operation %s: %w", op.ID, err)
	}
	if !applied {
		return nil
	}

	s.logStatusChange(&terminal)
	return nil
}

func generateID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// logStatusChange 统一记录 Operation 终态日志，并始终归属到创建该 Operation 的 request_id。
func (s *Service) logStatusChange(op *Operation) {
	if s.logWriter == nil {
		return
	}

	level := "INFO"
	message := "status=" + op.Status
	if op.Status == "Failed" {
		level = "ERROR"
		message = fmt.Sprintf("status=%s error_code=%s error_message=%s", op.Status, op.ErrorCode, op.ErrorMessage)
	}

	s.logWriter.Log(logger.Entry{
		RequestID:   op.CreatorRequestID,
		Level:       level,
		Event:       "operation.change",
		OperationID: op.ID,
		Node:        op.PVENode,
		Message:     message,
	})
}

// contextOrBackground 兜底 nil context，避免底层调用使用空 context 触发 panic。
func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
