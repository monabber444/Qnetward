package eventbus

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrBatcherClosed = errors.New("adaptive batcher is closed")
)

// BatchFlushFunc is the consumer callback receiving a compiled micro-batch.
type BatchFlushFunc func(ctx context.Context, batch [][]byte) error

// AdaptiveBatcher accumulates items into micro-batches, flushing on size or timer.
type AdaptiveBatcher struct {
	maxBatchSize  int
	flushInterval time.Duration
	flushFn       BatchFlushFunc
	inputChan     chan []byte
	closed        atomic.Bool
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewAdaptiveBatcher creates a micro-batcher that flushes on maxBatchSize or flushInterval.
func NewAdaptiveBatcher(maxBatchSize int, flushInterval time.Duration, flushFn BatchFlushFunc) *AdaptiveBatcher {
	if maxBatchSize <= 0 {
		maxBatchSize = 1000
	}
	if flushInterval <= 0 {
		flushInterval = 50 * time.Millisecond
	}

	ctx, cancel := context.WithCancel(context.Background())
	b := &AdaptiveBatcher{
		maxBatchSize:  maxBatchSize,
		flushInterval: flushInterval,
		flushFn:       flushFn,
		inputChan:     make(chan []byte, maxBatchSize*4),
		ctx:           ctx,
		cancel:        cancel,
	}

	b.wg.Add(1)
	go b.run()
	return b
}

func (b *AdaptiveBatcher) run() {
	defer b.wg.Done()
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	batch := make([][]byte, 0, b.maxBatchSize)

	flushCurrent := func() {
		if len(batch) == 0 {
			return
		}
		toFlush := make([][]byte, len(batch))
		copy(toFlush, batch)
		batch = batch[:0]
		if b.flushFn != nil {
			_ = b.flushFn(b.ctx, toFlush)
		}
	}

	for {
		select {
		case <-b.ctx.Done():
			flushCurrent()
			return

		case <-ticker.C:
			flushCurrent()

		case item, ok := <-b.inputChan:
			if !ok {
				flushCurrent()
				return
			}
			batch = append(batch, item)
			if len(batch) >= b.maxBatchSize {
				flushCurrent()
			}
		}
	}
}

// Push appends an item to the micro-batch queue.
func (b *AdaptiveBatcher) Push(item []byte) error {
	if b.closed.Load() {
		return ErrBatcherClosed
	}

	select {
	case b.inputChan <- item:
		return nil
	case <-b.ctx.Done():
		return ErrBatcherClosed
	}
}

// Close flushes remaining items and stops the batcher goroutine.
func (b *AdaptiveBatcher) Close() error {
	if b.closed.Swap(true) {
		return nil
	}
	close(b.inputChan)
	b.wg.Wait()
	b.cancel()
	return nil
}
