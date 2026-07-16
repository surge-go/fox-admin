package operlog

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	defaultRecorderQueueSize = 256
	auditWriteTimeout        = time.Second
	auditCloseTimeout        = 2 * time.Second
)

// Recorder 将操作审计写入有界队列，避免数据库写入阻塞业务请求。
type Recorder struct {
	service *Service
	logger  *zap.Logger
	queue   chan RecordInput
	stop    chan struct{}
	done    chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc

	mu        sync.RWMutex
	closed    bool
	closeOnce sync.Once
}

// NewRecorder 创建操作审计异步记录器。
func NewRecorder(service *Service, logger *zap.Logger) *Recorder {
	return newRecorder(service, logger, defaultRecorderQueueSize)
}

func newRecorder(service *Service, logger *zap.Logger, queueSize int) *Recorder {
	if service == nil {
		panic("operation log recorder service is nil")
	}
	if logger == nil {
		panic("operation log recorder logger is nil")
	}
	if queueSize <= 0 {
		panic("operation log recorder queue size must be positive")
	}

	ctx, cancel := context.WithCancel(context.Background())
	recorder := &Recorder{
		service: service,
		logger:  logger,
		queue:   make(chan RecordInput, queueSize),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		ctx:     ctx,
		cancel:  cancel,
	}
	go recorder.run()
	return recorder
}

// Enqueue 非阻塞提交一条操作审计；队列已满或记录器已关闭时返回 false。
func (r *Recorder) Enqueue(input *RecordInput) bool {
	if r == nil || input == nil {
		return false
	}

	item := cloneRecordInput(input)
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.closed {
		return false
	}
	select {
	case r.queue <- item:
		return true
	default:
		r.logger.Warn("操作日志队列已满，丢弃当前审计记录", zap.String("module", item.Module), zap.String("action", item.Action))
		return false
	}
}

// Close 停止接收新记录，并在限定时间内排空队列。
func (r *Recorder) Close() {
	if r == nil {
		return
	}
	r.closeOnce.Do(func() {
		r.mu.Lock()
		r.closed = true
		close(r.stop)
		r.mu.Unlock()

		timer := time.NewTimer(auditCloseTimeout)
		defer timer.Stop()
		select {
		case <-r.done:
		case <-timer.C:
			r.logger.Warn("操作日志队列关闭超时，取消剩余审计写入", zap.Int("pending", len(r.queue)))
			r.cancel()
			<-r.done
		}
		r.cancel()
	})
}

func (r *Recorder) run() {
	defer close(r.done)
	for {
		select {
		case input := <-r.queue:
			r.record(input)
		case <-r.stop:
			r.drain()
			return
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *Recorder) drain() {
	for {
		select {
		case <-r.ctx.Done():
			return
		case input := <-r.queue:
			r.record(input)
		default:
			return
		}
	}
}

func (r *Recorder) record(input RecordInput) {
	ctx, cancel := context.WithTimeout(r.ctx, auditWriteTimeout)
	defer cancel()
	if err := r.service.Record(ctx, &input); err != nil && r.ctx.Err() == nil {
		r.logger.Error("记录系统操作日志失败", zap.String("module", input.Module), zap.String("action", input.Action), zap.Error(err))
	}
}

func cloneRecordInput(input *RecordInput) RecordInput {
	cloned := *input
	if input.UserID != nil {
		userID := *input.UserID
		cloned.UserID = &userID
	}
	return cloned
}
