package eventbus

import (
	"context"
	"errors"
	"hash/fnv"
	"sync"
	"sync/atomic"
)

var (
	ErrPartitionOutOfRange = errors.New("partition index out of range")
	ErrQueueFull           = errors.New("partition queue buffer full")
)

// PartitionedBus extends Bus with key-based deterministic partition routing.
type PartitionedBus interface {
	Bus
	PublishKey(ctx context.Context, subject string, key string, data []byte) error
	SubscribePartition(ctx context.Context, subject string, partition int, handler Handler) (Subscription, error)
	NumPartitions() int
}

// PartitionedStreamBus provides high-throughput partitioned event distribution with in-order processing per key.
type PartitionedStreamBus struct {
	mu            sync.RWMutex
	numPartitions int
	queueDepth    int
	closed        atomic.Bool
	partitions    []*partitionWorker
	subs          map[string]map[int][]*partSub
	nextSubID     atomic.Int64
}

type partitionWorker struct {
	id     int
	queue  chan partitionItem
	cancel context.CancelFunc
}

type partitionItem struct {
	ctx     context.Context
	msg     Message
	key     string
	partIdx int
}

type partSub struct {
	id        int64
	subject   string
	partition int
	handler   Handler
	bus       *PartitionedStreamBus
}

func (s *partSub) Unsubscribe() error {
	return s.bus.unsubscribePart(s.subject, s.partition, s.id)
}

// NewPartitionedStreamBus creates a partitioned event streaming bus with dedicated partition workers.
func NewPartitionedStreamBus(numPartitions int, queueDepth int) *PartitionedStreamBus {
	if numPartitions <= 0 {
		numPartitions = 16
	}
	if queueDepth <= 0 {
		queueDepth = 4096
	}

	bus := &PartitionedStreamBus{
		numPartitions: numPartitions,
		queueDepth:    queueDepth,
		partitions:    make([]*partitionWorker, numPartitions),
		subs:          make(map[string]map[int][]*partSub),
	}

	for i := 0; i < numPartitions; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		w := &partitionWorker{
			id:     i,
			queue:  make(chan partitionItem, queueDepth),
			cancel: cancel,
		}
		bus.partitions[i] = w
		go bus.runWorker(ctx, w)
	}

	return bus
}

func (b *PartitionedStreamBus) runWorker(ctx context.Context, w *partitionWorker) {
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-w.queue:
			if !ok {
				return
			}
			b.dispatchItem(item)
		}
	}
}

func (b *PartitionedStreamBus) dispatchItem(item partitionItem) {
	b.mu.RLock()
	partMap := b.subs[item.msg.Subject]
	var handlers []Handler
	if partMap != nil {
		for _, s := range partMap[item.partIdx] {
			handlers = append(handlers, s.handler)
		}
		// Also dispatch to wildcard partition listeners (-1)
		for _, s := range partMap[-1] {
			handlers = append(handlers, s.handler)
		}
	}
	b.mu.RUnlock()

	for _, h := range handlers {
		_ = h(item.ctx, item.msg)
	}
}

// ComputePartition hashes key deterministically to a partition index.
func (b *PartitionedStreamBus) ComputePartition(key string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum64() % uint64(b.numPartitions))
}

// PublishKey publishes a message routed strictly by key to its designated partition.
func (b *PartitionedStreamBus) PublishKey(ctx context.Context, subject string, key string, data []byte) error {
	if b.closed.Load() {
		return ErrClosed
	}
	if subject == "" {
		return ErrInvalidTopic
	}

	partIdx := b.ComputePartition(key)
	w := b.partitions[partIdx]

	item := partitionItem{
		ctx: ctx,
		msg: Message{
			Subject: subject,
			Data:    data,
			Headers: map[string]string{
				"x-partition-key": key,
			},
		},
		key:     key,
		partIdx: partIdx,
	}

	select {
	case w.queue <- item:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrQueueFull
	}
}

// Publish implements the generic Bus interface by distributing to partition 0.
func (b *PartitionedStreamBus) Publish(ctx context.Context, subject string, data []byte) error {
	return b.PublishKey(ctx, subject, "", data)
}

// SubscribePartition subscribes to a specific partition or all partitions (-1).
func (b *PartitionedStreamBus) SubscribePartition(ctx context.Context, subject string, partition int, handler Handler) (Subscription, error) {
	if b.closed.Load() {
		return nil, ErrClosed
	}
	if subject == "" {
		return nil, ErrInvalidTopic
	}
	if handler == nil {
		return nil, ErrNilHandler
	}
	if partition != -1 && (partition < 0 || partition >= b.numPartitions) {
		return nil, ErrPartitionOutOfRange
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.subs[subject] == nil {
		b.subs[subject] = make(map[int][]*partSub)
	}

	sub := &partSub{
		id:        b.nextSubID.Add(1),
		subject:   subject,
		partition: partition,
		handler:   handler,
		bus:       b,
	}

	b.subs[subject][partition] = append(b.subs[subject][partition], sub)
	return sub, nil
}

// Subscribe subscribes to all partitions (-1) for a subject.
func (b *PartitionedStreamBus) Subscribe(ctx context.Context, subject string, handler Handler) (Subscription, error) {
	return b.SubscribePartition(ctx, subject, -1, handler)
}

func (b *PartitionedStreamBus) unsubscribePart(subject string, partition int, id int64) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	partMap := b.subs[subject]
	if partMap == nil {
		return nil
	}

	list := partMap[partition]
	for i, s := range list {
		if s.id == id {
			partMap[partition] = append(list[:i], list[i+1:]...)
			break
		}
	}
	return nil
}

// NumPartitions returns the total number of worker partitions.
func (b *PartitionedStreamBus) NumPartitions() int {
	return b.numPartitions
}

// Close gracefully terminates all partition worker channels.
func (b *PartitionedStreamBus) Close() error {
	if b.closed.Swap(true) {
		return nil
	}

	for _, w := range b.partitions {
		w.cancel()
	}
	return nil
}
