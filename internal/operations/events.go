package operations

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"hyperflow/internal/timeutil"
)

// OperationEvent 表示发送给 Kafka 的 operation 终态事件。
type OperationEvent struct {
	EventID          string               `json:"eventId"`
	OperationID      string               `json:"operationId"`
	Status           string               `json:"status"`
	ResourceLocation string               `json:"resourceLocation,omitempty"`
	Error            *OperationEventError `json:"error,omitempty"`
	ProviderTaskRef  string               `json:"providerTaskRef"`
	OccurredAt       time.Time            `json:"occurredAt"`
}

// OperationEventError 表示 operation 失败时的错误详情。
type OperationEventError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OutboxEvent 表示持久化在数据库中的待发布事件。
type OutboxEvent struct {
	ID          string
	OperationID string
	RequestID   string
	Topic       string
	Payload     []byte
	Attempts    int
	LastError   string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewOutboxEvent 基于终态 operation 构造待发布的 Kafka 事件。
func NewOutboxEvent(op *Operation, topic string) (*OutboxEvent, error) {
	if op == nil {
		return nil, fmt.Errorf("operation is required")
	}
	if topic == "" {
		return nil, fmt.Errorf("event topic is required")
	}

	eventID, err := generateEventID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate event id: %w", err)
	}

	payload := OperationEvent{
		EventID:          eventID,
		OperationID:      op.ID,
		Status:           op.Status,
		ResourceLocation: op.ResourceLocation,
		ProviderTaskRef:  op.PVEUpid,
		OccurredAt:       timeutil.NowShanghai(),
	}
	if op.Status == "Failed" {
		payload.Error = &OperationEventError{
			Code:    op.ErrorCode,
			Message: op.ErrorMessage,
		}
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal operation event: %w", err)
	}

	return &OutboxEvent{
		ID:          eventID,
		OperationID: op.ID,
		RequestID:   op.CreatorRequestID,
		Topic:       topic,
		Payload:     rawPayload,
	}, nil
}

func generateEventID() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
