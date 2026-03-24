package operations

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// TaskStatusQuerier 查询底层 PVE 任务状态的接口，由 pve.VmsService 实现
type TaskStatusQuerier interface {
	GetTaskStatus(node, upid string) (status string, exitStatus string, err error)
}

// Service 提供 LRO operation 的创建与查询逻辑
type Service struct {
	store   Store
	querier TaskStatusQuerier
}

// NewService 创建 Service
func NewService(store Store, querier TaskStatusQuerier) *Service {
	return &Service{store: store, querier: querier}
}

// CreateOperation 生成随机 ID，写入 DB，返回新建的 Operation
func (s *Service) CreateOperation(node, upid, resourceLocation string) (*Operation, error) {
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
	}
	if err := s.store.Insert(op); err != nil {
		return nil, fmt.Errorf("failed to persist operation: %w", err)
	}
	return op, nil
}

// GetOperation 懒查询：若操作仍为 Running，则查询 PVE 任务状态并按需更新 DB。
// 返回 nil, nil 表示 ID 不存在。
func (s *Service) GetOperation(id string) (*Operation, error) {
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
	pveStatus, pveExitStatus, err := s.querier.GetTaskStatus(op.PVENode, op.PVEUpid)
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
		_ = s.store.UpdateStatus(id, newStatus, errCode, errMsg)
		op.Status = newStatus
		op.ErrorCode = errCode
		op.ErrorMessage = errMsg
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
