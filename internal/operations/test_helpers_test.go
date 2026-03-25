package operations

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"hyperflow/internal/logger"
)

type fakeStore struct {
	mu          sync.Mutex
	ops         map[string]*Operation
	events      map[string]*OutboxEvent
	pendingErr  error
	runningErr  error
	getErr      error
	insertErr   error
	completeErr error
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		ops:    make(map[string]*Operation),
		events: make(map[string]*OutboxEvent),
	}
}

func (s *fakeStore) CreateTable() error {
	return nil
}

func (s *fakeStore) Insert(op *Operation) error {
	if s.insertErr != nil {
		return s.insertErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ops[op.ID] = cloneOperation(op)
	return nil
}

func (s *fakeStore) GetByID(id string) (*Operation, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.ops[id]
	if !ok {
		return nil, nil
	}
	return cloneOperation(op), nil
}

func (s *fakeStore) ListRunning(limit int) ([]*Operation, error) {
	if s.runningErr != nil {
		return nil, s.runningErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := make([]string, 0, len(s.ops))
	for id, op := range s.ops {
		if op.Status == "Running" {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	if limit > 0 && len(ids) > limit {
		ids = ids[:limit]
	}

	ops := make([]*Operation, 0, len(ids))
	for _, id := range ids {
		ops = append(ops, cloneOperation(s.ops[id]))
	}
	return ops, nil
}

func (s *fakeStore) CompleteOperation(op *Operation, event *OutboxEvent) (bool, error) {
	if s.completeErr != nil {
		return false, s.completeErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.ops[op.ID]
	if !ok || current.Status != "Running" {
		return false, nil
	}

	s.ops[op.ID] = cloneOperation(op)
	s.events[event.ID] = cloneOutboxEvent(event)
	return true, nil
}

func (s *fakeStore) ListPendingEvents(limit int) ([]*OutboxEvent, error) {
	if s.pendingErr != nil {
		return nil, s.pendingErr
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := make([]string, 0, len(s.events))
	for id, event := range s.events {
		if event.PublishedAt == nil {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	if limit > 0 && len(ids) > limit {
		ids = ids[:limit]
	}

	events := make([]*OutboxEvent, 0, len(ids))
	for _, id := range ids {
		events = append(events, cloneOutboxEvent(s.events[id]))
	}
	return events, nil
}

func (s *fakeStore) MarkEventPublished(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	event, ok := s.events[id]
	if !ok {
		return errors.New("event not found")
	}
	now := time.Now()
	event.PublishedAt = &now
	event.LastError = ""
	return nil
}

func (s *fakeStore) MarkEventPublishFailed(id, lastError string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	event, ok := s.events[id]
	if !ok {
		return errors.New("event not found")
	}
	event.Attempts++
	event.LastError = lastError
	return nil
}

type fakeQuerier struct {
	status     string
	exitStatus string
	err        error
}

func (q fakeQuerier) GetTaskStatus(ctx context.Context, node, upid string) (string, string, error) {
	return q.status, q.exitStatus, q.err
}

type captureLogger struct {
	mu      sync.Mutex
	entries []logger.Entry
}

func (l *captureLogger) Log(entry logger.Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, entry)
}

func (l *captureLogger) Shutdown(ctx context.Context) {}

func (l *captureLogger) Entries() []logger.Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]logger.Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

type fakeProducer struct {
	mu        sync.Mutex
	published []publishedMessage
	err       error
	closed    bool
}

type publishedMessage struct {
	topic string
	key   []byte
	value []byte
}

func (p *fakeProducer) Publish(ctx context.Context, topic string, key, value []byte) error {
	if p.err != nil {
		return p.err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.published = append(p.published, publishedMessage{
		topic: topic,
		key:   append([]byte(nil), key...),
		value: append([]byte(nil), value...),
	})
	return nil
}

func (p *fakeProducer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return nil
}

func cloneOperation(op *Operation) *Operation {
	if op == nil {
		return nil
	}
	cloned := *op
	return &cloned
}

func cloneOutboxEvent(event *OutboxEvent) *OutboxEvent {
	if event == nil {
		return nil
	}
	cloned := *event
	cloned.Payload = append([]byte(nil), event.Payload...)
	if event.PublishedAt != nil {
		ts := *event.PublishedAt
		cloned.PublishedAt = &ts
	}
	return &cloned
}
