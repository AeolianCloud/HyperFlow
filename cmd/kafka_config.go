package main

import (
	"fmt"
	"os"
	"strings"
)

type kafkaConfig struct {
	Brokers              []string
	OperationEventsTopic string
}

func loadKafkaConfigFromEnv() (kafkaConfig, error) {
	cfg := kafkaConfig{
		OperationEventsTopic: strings.TrimSpace(os.Getenv("KAFKA_OPERATION_EVENTS_TOPIC")),
	}

	for _, broker := range strings.Split(strings.TrimSpace(os.Getenv("KAFKA_BROKERS")), ",") {
		broker = strings.TrimSpace(broker)
		if broker == "" {
			continue
		}
		cfg.Brokers = append(cfg.Brokers, broker)
	}

	if len(cfg.Brokers) == 0 {
		return kafkaConfig{}, fmt.Errorf("KAFKA_BROKERS environment variable is required")
	}
	if cfg.OperationEventsTopic == "" {
		return kafkaConfig{}, fmt.Errorf("KAFKA_OPERATION_EVENTS_TOPIC environment variable is required")
	}

	return cfg, nil
}
