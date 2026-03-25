package operations

import (
	"context"
	"encoding/json"
	"testing"
)

func TestServiceReconcileRunningOperationsCreatesTerminalEvent(t *testing.T) {
	store := newFakeStore()
	store.ops["op-1"] = &Operation{
		ID:               "op-1",
		Status:           "Running",
		PVENode:          "node-a",
		PVEUpid:          "UPID:node-a:1",
		ResourceLocation: "/api/pve/nodes/node-a/vms/100",
		CreatorRequestID: "req-1",
	}
	logWriter := &captureLogger{}
	service := NewService(store, fakeQuerier{status: "stopped", exitStatus: "OK"}, logWriter, "hyperflow.operation-events")

	if err := service.ReconcileRunningOperations(context.Background(), 10); err != nil {
		t.Fatalf("ReconcileRunningOperations returned error: %v", err)
	}

	op, err := store.GetByID("op-1")
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if op == nil || op.Status != "Succeeded" {
		t.Fatalf("expected operation to be succeeded, got %#v", op)
	}

	pending, err := store.ListPendingEvents(10)
	if err != nil {
		t.Fatalf("ListPendingEvents returned error: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending event, got %d", len(pending))
	}
	if pending[0].Topic != "hyperflow.operation-events" {
		t.Fatalf("expected topic hyperflow.operation-events, got %q", pending[0].Topic)
	}

	var payload OperationEvent
	if err := json.Unmarshal(pending[0].Payload, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload.OperationID != "op-1" || payload.Status != "Succeeded" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.ProviderTaskRef != "UPID:node-a:1" {
		t.Fatalf("expected providerTaskRef to match UPID, got %q", payload.ProviderTaskRef)
	}

	entries := logWriter.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}
	if entries[0].Event != "operation.change" || entries[0].RequestID != "req-1" {
		t.Fatalf("unexpected log entry: %#v", entries[0])
	}
}

func TestServiceReconcileRunningOperationsHandlesFailedTask(t *testing.T) {
	store := newFakeStore()
	store.ops["op-2"] = &Operation{
		ID:               "op-2",
		Status:           "Running",
		PVENode:          "node-b",
		PVEUpid:          "UPID:node-b:2",
		ResourceLocation: "/api/pve/nodes/node-b/vms/200",
		CreatorRequestID: "req-2",
	}
	service := NewService(store, fakeQuerier{status: "stopped", exitStatus: "disk full"}, &captureLogger{}, "hyperflow.operation-events")

	if err := service.ReconcileRunningOperations(context.Background(), 10); err != nil {
		t.Fatalf("ReconcileRunningOperations returned error: %v", err)
	}

	op, err := store.GetByID("op-2")
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if op == nil || op.Status != "Failed" {
		t.Fatalf("expected operation to be failed, got %#v", op)
	}
	if op.ErrorCode != "TaskFailed" || op.ErrorMessage != "disk full" {
		t.Fatalf("unexpected operation error fields: %#v", op)
	}
}

func TestServiceGetOperationReconcilesRunningOperationOnRead(t *testing.T) {
	store := newFakeStore()
	store.ops["op-3"] = &Operation{
		ID:               "op-3",
		Status:           "Running",
		PVENode:          "node-c",
		PVEUpid:          "UPID:node-c:3",
		ResourceLocation: "/api/pve/nodes/node-c/vms/300",
		CreatorRequestID: "req-3",
	}
	service := NewService(store, fakeQuerier{status: "stopped", exitStatus: "OK"}, &captureLogger{}, "hyperflow.operation-events")

	op, err := service.GetOperation(context.Background(), "op-3")
	if err != nil {
		t.Fatalf("GetOperation returned error: %v", err)
	}
	if op == nil || op.Status != "Succeeded" {
		t.Fatalf("expected reconciled operation to be succeeded, got %#v", op)
	}

	pending, err := store.ListPendingEvents(10)
	if err != nil {
		t.Fatalf("ListPendingEvents returned error: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending event after read-time reconcile, got %d", len(pending))
	}
}
