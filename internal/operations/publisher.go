package operations

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"hyperflow/internal/logger"
)

// Producer 定义向 Kafka 发布事件的最小接口。
type Producer interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
	Close() error
}

// KafkaProducer 负责将事件写入 Kafka。
type KafkaProducer struct {
	topic  string
	writer *kafka.Writer
}

// NewKafkaProducer 创建基于 kafka-go 的 producer。
func NewKafkaProducer(brokers []string, topic string) (*KafkaProducer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("at least one Kafka broker is required")
	}
	if topic == "" {
		return nil, fmt.Errorf("Kafka topic is required")
	}

	return &KafkaProducer{
		topic: topic,
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.Hash{},
		},
	}, nil
}

// Publish 将单条 operation 事件写入 Kafka。
func (p *KafkaProducer) Publish(ctx context.Context, topic string, key, value []byte) error {
	if topic != p.topic {
		return fmt.Errorf("unexpected topic %q", topic)
	}

	return p.writer.WriteMessages(contextOrBackground(ctx), kafka.Message{
		Key:   key,
		Value: value,
	})
}

// Close 关闭 Kafka producer。
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

// OutboxPublisher 周期性将待发布事件发送到 Kafka。
type OutboxPublisher struct {
	store     Store
	producer  Producer
	logWriter logger.Logger
	interval  time.Duration
	batchSize int
	done      chan struct{}
}

// NewOutboxPublisher 创建 outbox publisher。
func NewOutboxPublisher(store Store, producer Producer, logWriter logger.Logger, interval time.Duration, batchSize int) *OutboxPublisher {
	if interval <= 0 {
		interval = time.Second
	}
	if batchSize <= 0 {
		batchSize = 100
	}

	return &OutboxPublisher{
		store:     store,
		producer:  producer,
		logWriter: logWriter,
		interval:  interval,
		batchSize: batchSize,
		done:      make(chan struct{}),
	}
}

// Start 启动后台发布循环。
func (p *OutboxPublisher) Start(ctx context.Context) {
	go func() {
		defer close(p.done)

		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			_ = p.PublishPending(contextOrBackground(ctx))

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// Shutdown 停止发布循环并关闭底层 producer。
func (p *OutboxPublisher) Shutdown(ctx context.Context) {
	select {
	case <-p.done:
	case <-ctx.Done():
	}

	if p.producer != nil {
		_ = p.producer.Close()
	}
}

// PublishPending 发布一批待发送事件。
func (p *OutboxPublisher) PublishPending(ctx context.Context) error {
	events, err := p.store.ListPendingEvents(p.batchSize)
	if err != nil {
		return fmt.Errorf("failed to list pending outbox events: %w", err)
	}

	var firstErr error
	for _, event := range events {
		if err := p.producer.Publish(ctx, event.Topic, []byte(event.OperationID), event.Payload); err != nil {
			_ = p.store.MarkEventPublishFailed(event.ID, err.Error())
			p.logPublishResult(event, "ERROR", fmt.Sprintf("topic=%s error=%s", event.Topic, err.Error()))
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		if err := p.store.MarkEventPublished(event.ID); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		p.logPublishResult(event, "INFO", "topic="+event.Topic)
	}

	return firstErr
}

func (p *OutboxPublisher) logPublishResult(event *OutboxEvent, level, message string) {
	if p.logWriter == nil || event == nil {
		return
	}

	p.logWriter.Log(logger.Entry{
		RequestID:   event.RequestID,
		Level:       level,
		Event:       "operation.event.publish",
		OperationID: event.OperationID,
		Message:     message,
	})
}
