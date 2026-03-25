package operations

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestOutboxPublisherPublishesPersistedPendingEvent(t *testing.T) {
	store := newFakeStore()
	op := &Operation{
		ID:               "op-1",
		Status:           "Succeeded",
		PVEUpid:          "UPID:node-a:1",
		ResourceLocation: "/api/pve/nodes/node-a/vms/100",
		CreatorRequestID: "req-1",
	}
	event, err := NewOutboxEvent(op, "hyperflow.operation-events")
	if err != nil {
		t.Fatalf("NewOutboxEvent returned error: %v", err)
	}
	store.events[event.ID] = cloneOutboxEvent(event)

	producer := &fakeProducer{}
	logWriter := &captureLogger{}
	publisher := NewOutboxPublisher(store, producer, logWriter, 0, 10)

	if err := publisher.PublishPending(context.Background()); err != nil {
		t.Fatalf("PublishPending returned error: %v", err)
	}

	if len(producer.published) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(producer.published))
	}
	if string(producer.published[0].key) != "op-1" {
		t.Fatalf("expected Kafka message key to equal operation id, got %q", string(producer.published[0].key))
	}

	persisted := store.events[event.ID]
	if persisted.PublishedAt == nil {
		t.Fatalf("expected event to be marked published")
	}

	var payload OperationEvent
	if err := json.Unmarshal(producer.published[0].value, &payload); err != nil {
		t.Fatalf("failed to decode published payload: %v", err)
	}
	if payload.EventID == "" || payload.OperationID != "op-1" {
		t.Fatalf("unexpected published payload: %#v", payload)
	}

	entries := logWriter.Entries()
	if len(entries) != 1 || entries[0].Event != "operation.event.publish" || entries[0].Level != "INFO" {
		t.Fatalf("unexpected publish log entries: %#v", entries)
	}
}

func TestOutboxPublisherRecordsPublishFailure(t *testing.T) {
	store := newFakeStore()
	op := &Operation{
		ID:               "op-2",
		Status:           "Failed",
		PVEUpid:          "UPID:node-b:2",
		ResourceLocation: "/api/pve/nodes/node-b/vms/200",
		ErrorCode:        "TaskFailed",
		ErrorMessage:     "disk full",
		CreatorRequestID: "req-2",
	}
	event, err := NewOutboxEvent(op, "hyperflow.operation-events")
	if err != nil {
		t.Fatalf("NewOutboxEvent returned error: %v", err)
	}
	store.events[event.ID] = cloneOutboxEvent(event)

	producer := &fakeProducer{err: errors.New("kafka unavailable")}
	logWriter := &captureLogger{}
	publisher := NewOutboxPublisher(store, producer, logWriter, 0, 10)

	if err := publisher.PublishPending(context.Background()); err == nil {
		t.Fatalf("expected PublishPending to return an error")
	}

	persisted := store.events[event.ID]
	if persisted.PublishedAt != nil {
		t.Fatalf("expected event to remain unpublished")
	}
	if persisted.Attempts != 1 || persisted.LastError != "kafka unavailable" {
		t.Fatalf("unexpected persisted failure state: %#v", persisted)
	}

	entries := logWriter.Entries()
	if len(entries) != 1 || entries[0].Level != "ERROR" {
		t.Fatalf("unexpected publish failure log entries: %#v", entries)
	}
}
