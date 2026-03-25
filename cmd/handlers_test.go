package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"hyperflow/internal/operations"
)

type handlerStore struct {
	op *operations.Operation
}

func (s *handlerStore) CreateTable() error                    { return nil }
func (s *handlerStore) Insert(op *operations.Operation) error { return nil }
func (s *handlerStore) GetByID(id string) (*operations.Operation, error) {
	if s.op == nil || s.op.ID != id {
		return nil, nil
	}
	cloned := *s.op
	return &cloned, nil
}
func (s *handlerStore) ListRunning(limit int) ([]*operations.Operation, error) {
	return nil, nil
}
func (s *handlerStore) CompleteOperation(op *operations.Operation, event *operations.OutboxEvent) (bool, error) {
	return false, nil
}
func (s *handlerStore) ListPendingEvents(limit int) ([]*operations.OutboxEvent, error) {
	return nil, nil
}
func (s *handlerStore) MarkEventPublished(id string) error                { return nil }
func (s *handlerStore) MarkEventPublishFailed(id, lastError string) error { return nil }

func TestGetOperationReturnsOperationResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := &handlerStore{
		op: &operations.Operation{
			ID:               "op-1",
			Status:           "Succeeded",
			ResourceLocation: "/api/pve/nodes/node-a/vms/100",
		},
	}
	operationsSvcGlobal = operations.NewService(store, nil, nil, "hyperflow.operation-events")

	router := gin.New()
	api := router.Group("/api/pve")
	registerOperationsRoutes(api.Group("/operations"), operationsSvcGlobal)

	req := httptest.NewRequest(http.MethodGet, "/api/pve/operations/op-1", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var response OperationResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.ID != "op-1" || response.Status != "Succeeded" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestGetOperationReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	operationsSvcGlobal = operations.NewService(&handlerStore{}, nil, nil, "hyperflow.operation-events")

	router := gin.New()
	api := router.Group("/api/pve")
	registerOperationsRoutes(api.Group("/operations"), operationsSvcGlobal)

	req := httptest.NewRequest(http.MethodGet, "/api/pve/operations/missing", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}
