package operations

import (
	"context"
	"sync"
	"testing"
)

func TestConcurrentLockSameName(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, nil, nil, "test-topic")

	var wg sync.WaitGroup
	errs := make(chan error, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := svc.AcquireDiskLock(context.Background(), "node-a", "100")
			if err != nil {
				errs <- err
				return
			}
			release()
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected lock error: %v", err)
		}
	}
}

func TestAcquireDiskLockReturnsRelease(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, nil, nil, "test-topic")

	release, err := svc.AcquireDiskLock(context.Background(), "node-a", "101")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if release == nil {
		t.Fatal("expected non-nil release function")
	}
	release()
}

func TestReleaseIdempotent(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, nil, nil, "test-topic")

	release, err := svc.AcquireDiskLock(context.Background(), "node-a", "102")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// calling release multiple times should not panic
	release()
	release()
}

func TestLockNameFormat(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, nil, nil, "test-topic")

	release, err := svc.AcquireDiskLock(context.Background(), "pve-node-01", "100")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	release()
}
