package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// startConsumer reads snapshots from Kafka and pushes them onto the same queue /ingest uses.
func startConsumer(ctx context.Context) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return // Kafka not configured: HTTP ingest only
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(brokers, ","),
		GroupID:  "nexus-api",
		Topic:    "nexus.telemetry",
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  500 * time.Millisecond,
	})
	go func() {
		defer r.Close()
		for {
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("kafka read: %v", err)
				time.Sleep(time.Second)
				continue
			}
			var s Snapshot
			if err := json.Unmarshal(m.Value, &s); err != nil {
				log.Printf("kafka: bad message: %v", err)
				continue
			}
			// Block when the queue is full instead of dropping: Kafka keeps the backlog for us.
			select {
			case queue <- s:
			case <-ctx.Done():
				return
			}
		}
	}()
	log.Printf("kafka consumer started (group nexus-api, brokers %s)", brokers)
}
