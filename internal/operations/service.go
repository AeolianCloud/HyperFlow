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
	store     Store
	querier   TaskStatusQuerier
	logWriter logger.Logger
}

// NewService 创建 Service
func NewService(store Store, querier TaskStatusQuerier, logWriter logger.Logger) *Service {
	return &Service{store: store, querier: querier, logWriter: logWriter}
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

// GetOperation 懒查询：若操作仍为 Running，则查询 PVE 任务状态并按需更新 DB。
// 返回 nil, nil 表示 ID 不存在。
func (s *Service) GetOperation(ctx context.Context, id string) (*Operation, error) {
	op, err := s.store.GetByID(id)
	if err != nil {
		return nil, err
	}
	if op == nil {
		return nil, nil
	}
	if op.Status != "Running" {
		return op, nil
	}
	// 懒更新：查询 PVE 底层任务状态
	pveStatus, pveExitStatus, err := s.querier.GetTaskStatus(contextOrBackground(ctx), op.PVENode, op.PVEUpid)
	if err != nil {
		// PVE 不可达时返回当前缓存状态，不报错
		return op, nil
	}
	if pveStatus == "stopped" {
		var newStatus, errCode, errMsg string
		if pveExitStatus == "OK" {
			newStatus = "Succeeded"
		} else {
			newStatus = "Failed"
			errCode = "TaskFailed"
			errMsg = pveExitStatus
		}
		if err := s.store.UpdateStatus(id, newStatus, errCode, errMsg); err != nil {
			return nil, fmt.Errorf("failed to update operation status: %w", err)
		}
		op.Status = newStatus
		op.ErrorCode = errCode
		op.ErrorMessage = errMsg
		s.logStatusChange(op)
	}
	return op, nil
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
