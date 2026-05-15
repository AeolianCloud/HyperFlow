package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"hyperflow/internal/operations"
)

type diskHandlerStore struct {
	operations.Store
}

func (s *diskHandlerStore) AcquireLock(ctx context.Context, name string, timeout int) (func(), error) {
	return func() {}, nil
}

func (s *diskHandlerStore) Insert(op *operations.Operation) error {
	op.ID = "test-op-1"
	return nil
}

func TestAttachDiskRequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := &diskHandlerStore{}
	operationsSvcGlobal = operations.NewService(store, nil, nil, "test-topic")

	router := gin.New()
	api := router.Group("/api/pve")
	nodes := api.Group("/nodes/:node")
	registerVmsRoutes(nodes.Group("/vms"), nil)

	tests := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{
			name:       "missing body",
			body:       nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing size and storage",
			body:       map[string]any{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "zero size",
			body:       map[string]any{"size": 0, "storage": "local"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty storage",
			body:       map[string]any{"size": 100, "storage": ""},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/pve/nodes/test-node/vms/100/disks", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d; body: %s", tt.wantStatus, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestCreateVmDataDisksValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := &diskHandlerStore{}
	operationsSvcGlobal = operations.NewService(store, nil, nil, "test-topic")

	router := gin.New()
	api := router.Group("/api/pve")
	nodes := api.Group("/nodes/:node")
	registerVmsRoutes(nodes.Group("/vms"), nil)

	tests := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{
			name: "data disk missing size",
			body: map[string]any{
				"vmid":       200,
				"cores":      2,
				"memory":     2048,
				"diskSource": "local:import/focal.qcow2",
				"storage":    "local-lvm",
				"dataDisks":  []any{map[string]any{"storage": "ceph"}},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "data disk missing storage",
			body: map[string]any{
				"vmid":       200,
				"cores":      2,
				"memory":     2048,
				"diskSource": "local:import/focal.qcow2",
				"storage":    "local-lvm",
				"dataDisks":  []any{map[string]any{"size": 100}},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/pve/nodes/test-node/vms", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d; body: %s", tt.wantStatus, recorder.Code, recorder.Body.String())
			}
		})
	}
}
