package operations

import (
	"context"
	"time"
)

// Reconciler 周期性推进 Running operation 到终态。
type Reconciler struct {
	service   *Service
	interval  time.Duration
	batchSize int
	done      chan struct{}
}

// NewReconciler 创建后台 operation 状态推进器。
func NewReconciler(service *Service, interval time.Duration, batchSize int) *Reconciler {
	if interval <= 0 {
		interval = time.Second
	}
	if batchSize <= 0 {
		batchSize = 100
	}

	return &Reconciler{
		service:   service,
		interval:  interval,
		batchSize: batchSize,
		done:      make(chan struct{}),
	}
}

// Start 启动后台轮询。
func (r *Reconciler) Start(ctx context.Context) {
	go func() {
		defer close(r.done)

		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		for {
			_ = r.service.ReconcileRunningOperations(ctx, r.batchSize)

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// Shutdown 等待后台轮询停止。
func (r *Reconciler) Shutdown(ctx context.Context) {
	select {
	case <-r.done:
	case <-ctx.Done():
	}
}
