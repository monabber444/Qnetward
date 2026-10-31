package eventbus

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPartitionedStreamBus_KeyRouting(t *testing.T) {
	bus := NewPartitionedStreamBus(8, 1024)
	defer bus.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := "tenant-alpha:device-42"
	targetPart := bus.ComputePartition(key)

	var wg sync.WaitGroup
	var receivedTargetPart int
	var receivedWildcard int
	var mu sync.Mutex

	wg.Add(20) // 10 messages * 2 subscribers (target partition sub + wildcard sub)

	// Subscribe to specific target partition
	_, err := bus.SubscribePartition(ctx, "telemetry.hot", targetPart, func(ctx context.Context, msg Message) error {
		mu.Lock()
		receivedTargetPart++
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("SubscribePartition failed: %v", err)
	}

	// Subscribe to all partitions (wildcard -1)
	_, err = bus.SubscribePartition(ctx, "telemetry.hot", -1, func(ctx context.Context, msg Message) error {
		mu.Lock()
		receivedWildcard++
		mu.Unlock()
		wg.Done()
		return nil
	})
	if err != nil {
		t.Fatalf("SubscribePartition wildcard failed: %v", err)
	}

	// Publish 10 messages with same key
	for i := 0; i < 10; i++ {
		payload := []byte(fmt.Sprintf(`{"seq":%d}`, i))
		if err := bus.PublishKey(ctx, "telemetry.hot", key, payload); err != nil {
			t.Fatalf("PublishKey failed: %v", err)
		}
	}

	wg.Wait()

	mu.Lock()
	if receivedTargetPart != 10 {
		t.Errorf("Expected 10 messages in target partition, got %d", receivedTargetPart)
	}
	if receivedWildcard != 10 {
		t.Errorf("Expected 10 messages in wildcard subscriber, got %d", receivedWildcard)
	}
	mu.Unlock()
}

func TestAdaptiveBatcher_SizeAndTimeFlushes(t *testing.T) {
	var mu sync.Mutex
	var flushedBatches [][][]byte

	flushFn := func(ctx context.Context, batch [][]byte) error {
		mu.Lock()
		flushedBatches = append(flushedBatches, batch)
		mu.Unlock()
		return nil
	}

	// Batcher with maxBatchSize = 10, flushInterval = 60ms
	batcher := NewAdaptiveBatcher(10, 60*time.Millisecond, flushFn)

	// 1. Push 25 items: should trigger 2 full batches (size 10 each) immediately
	for i := 0; i < 25; i++ {
		_ = batcher.Push([]byte(fmt.Sprintf("item-%d", i)))
	}

	// Wait brief interval to let workers process size flushes
	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	initialFlushed := len(flushedBatches)
	mu.Unlock()

	if initialFlushed < 2 {
		t.Errorf("Expected at least 2 size flushes, got %d", initialFlushed)
	}

	// 2. Wait for timer flush of remaining 5 items (>60ms)
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	finalFlushed := len(flushedBatches)
	var totalItems int
	for _, b := range flushedBatches {
		totalItems += len(b)
	}
	mu.Unlock()

	if finalFlushed < 3 {
		t.Errorf("Expected at least 3 flushes after interval, got %d", finalFlushed)
	}
	if totalItems != 25 {
		t.Errorf("Expected 25 total items across all flushes, got %d", totalItems)
	}

	_ = batcher.Close()
}
