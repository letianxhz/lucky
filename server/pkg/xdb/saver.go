package xdb

import (
	"context"
	"sync"
	"time"
)

const BatchSize = int32(256)
const RetryInterval = 100 * time.Millisecond

// Saver 保存器
type Saver struct {
	workers      []SaveWorker
	timeout      time.Duration
	SyncInterval time.Duration
	src          *Source
}

// NewSaver 创建保存器
func NewSaver(src *Source, concurrence uint32, timeout time.Duration, syncInterval time.Duration) *Saver {
	s := &Saver{
		src:          src,
		timeout:      timeout,
		SyncInterval: syncInterval,
		workers:      make([]SaveWorker, concurrence),
	}

	for i := range s.workers {
		s.workers[i].Init(s)
	}

	return s
}

// Run 运行保存器
func (s *Saver) Run(ctx context.Context, wg *sync.WaitGroup) {
	for i := range s.workers {
		s.workers[i].Run(ctx, wg)
	}
}

// Put 放入提交对象
func (s *Saver) Put(ctx context.Context, c Commitment, i int32) int32 {
	worker := s.getWorker(PKOf(c))
	return worker.Put(ctx, c, i)
}

// Sync 同步
func (s *Saver) Sync(pk PK) {
	if pk != nil {
		s.getWorker(pk).Sync()
		return
	}

	wg := sync.WaitGroup{}
	for i := range s.workers {
		worker := &s.workers[i]
		wg.Add(1)
		go func() {
			worker.Sync()
			wg.Done()
		}()
	}

	wg.Wait()
}

// OnGoingCount 获取进行中的数量
func (s *Saver) OnGoingCount() int32 {
	ongoing := int32(0)
	for i := range s.workers {
		ongoing += s.workers[i].Ongoing()
	}
	return ongoing
}

// Close 关闭保存器
func (s *Saver) Close() {
	for i := range s.workers {
		s.workers[i].Close()
	}
}

// Recover 恢复
func (s *Saver) Recover(ctx context.Context, wg *sync.WaitGroup) {
	// TODO: 实现重做日志恢复
	// 这里可以读取重做日志文件并恢复未完成的提交
}

func (s *Saver) getWorker(pk PK) *SaveWorker {
	i := HashGroupPK(pk) % cap(s.workers)
	return &s.workers[i]
}

// SaveWorker 保存工作器
type SaveWorker struct {
	mu        sync.RWMutex
	owner     *Saver
	condProd  sync.Cond
	condCons  sync.Cond
	receiving *CommitmentBatch
	syncTimer *time.Timer
	running   bool
}

// Init 初始化工作器
func (sw *SaveWorker) Init(owner *Saver) {
	sw.condCons.L = &sw.mu
	sw.condProd.L = &sw.mu
	sw.owner = owner
}

// Run 运行工作器
func (sw *SaveWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	sw.receiving = getCommitmentBatch(ctx, sw)
	sw.running = true
	wg.Add(1)
	go func() {
		defer wg.Done()
		sw.consume(ctx)
	}()
}

// Put 放入提交对象
func (sw *SaveWorker) Put(ctx context.Context, c Commitment, index int32) int32 {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	for sw.running {
		cb := sw.receiving
		index = cb.Put(ctx, c, index)
		if index < 0 {
			sw.condCons.Signal()
			sw.condProd.Wait()
			continue
		}

		if sw.syncTimer == nil {
			sw.syncTimer = time.AfterFunc(sw.owner.SyncInterval, func() {
				sw.mu.Lock()
				defer sw.mu.Unlock()
				if sw.receiving != nil {
					sw.receiving.overtime = true
					sw.condCons.Signal()
				}
			})
		}

		return index
	}

	return -1
}

// Close 关闭工作器
func (sw *SaveWorker) Close() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.running = false
	sw.condCons.Signal()
	sw.condProd.Broadcast()
}

func (sw *SaveWorker) consume(ctx context.Context) {
	for {
		batch, running := sw.poll(ctx)
		if batch.Len() > 0 {
			if !sw.owner.src.Table().Save(ctx, batch.entries, sw.owner.timeout, RetryInterval, sw.Running) {
				break
			}
		}

		putCommitmentBatch(batch)
		if !running {
			break
		}
	}
}

func (sw *SaveWorker) poll(ctx context.Context) (*CommitmentBatch, bool) {
	sw.mu.Lock()

	defer func() {
		if sw.syncTimer != nil {
			sw.syncTimer.Stop()
			sw.syncTimer = nil
		}

		sw.mu.Unlock()
	}()

	for {
		current := sw.receiving

		if !sw.running {
			sw.receiving = nil
			return current, false
		}

		if current.Consumable() {
			current.Retire(ctx)
			sw.receiving = getCommitmentBatch(ctx, sw)
			sw.condProd.Broadcast()
			return current, true
		}

		sw.condCons.Wait()
	}
}

// Running 检查是否运行中
func (sw *SaveWorker) Running() bool {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.running
}

// Ongoing 获取进行中的数量
func (sw *SaveWorker) Ongoing() int32 {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	if sw.receiving == nil {
		return 0
	}
	return sw.receiving.Len()
}

// Sync 同步
func (sw *SaveWorker) Sync() {
	<-(func() <-chan interface{} {
		sw.mu.Lock()
		defer sw.mu.Unlock()
		chSync := sw.receiving.Sync()
		sw.condCons.Signal()
		return chSync
	})()
}

// CommitmentBatch 提交批次
type CommitmentBatch struct {
	entries  []Commitment
	redo     RedoLogFile
	overtime bool
	chSync   chan interface{}
}

// Sync 同步
func (cb *CommitmentBatch) Sync() <-chan interface{} {
	chSync := cb.chSync
	if chSync == nil {
		chSync = make(chan interface{})
		cb.chSync = chSync
	}
	return chSync
}

// Consumable 检查是否可消费
func (cb *CommitmentBatch) Consumable() bool {
	l := cb.Len()
	return l == BatchSize || cb.overtime || cb.chSync != nil
}

// Len 获取长度
func (cb *CommitmentBatch) Len() int32 {
	if cb == nil {
		return 0
	}
	return int32(len(cb.entries))
}

// Retire 退役
func (cb *CommitmentBatch) Retire(ctx context.Context) {
	cb.redo.Retire(ctx)
}

// Put 放入提交对象
func (cb *CommitmentBatch) Put(ctx context.Context, c Commitment, index int32) int32 {
	l := cb.Len()

	if index >= 0 && index < l && cb.entries[index].Merge(c) {
		cb.redo.Log(ctx, c)
		return index
	}

	if l < BatchSize {
		cb.entries = append(cb.entries, c)
		cb.redo.Log(ctx, c)
		return l
	}

	return -1
}

var batchPool = sync.Pool{
	New: func() interface{} {
		return &CommitmentBatch{
			entries: make([]Commitment, 0, BatchSize),
		}
	},
}

func getCommitmentBatch(ctx context.Context, sw *SaveWorker) *CommitmentBatch {
	batch := batchPool.Get().(*CommitmentBatch)
	// 初始化redo（如果未初始化）
	if batch.redo == nil {
		batch.redo = &simpleRedoLog{}
	}
	batch.redo.Serve(ctx, sw)
	return batch
}

func putCommitmentBatch(c *CommitmentBatch) {
	c.redo.Destroy()
	if c.chSync != nil {
		close(c.chSync)
		c.chSync = nil
	}
	c.overtime = false
	c.entries = c.entries[0:0]
	batchPool.Put(c)
}

// RedoLogFile 重做日志文件接口
type RedoLogFile interface {
	Serve(ctx context.Context, sw *SaveWorker)
	Log(ctx context.Context, c Commitment)
	Retire(ctx context.Context)
	Destroy()
}

// 简单的重做日志实现（可以后续扩展）
type simpleRedoLog struct{}

func (s *simpleRedoLog) Serve(ctx context.Context, sw *SaveWorker) {}
func (s *simpleRedoLog) Log(ctx context.Context, c Commitment)     {}
func (s *simpleRedoLog) Retire(ctx context.Context)                {}
func (s *simpleRedoLog) Destroy()                                  {}

// HashGroupPK 获取主键的哈希组
func HashGroupPK(pk PK) int {
	src := pk.Source()
	return src.repo.HashGroup(pk)
}

// HashGroup 获取记录的哈希组
func HashGroup(v Record) int {
	return HashGroupPK(v.Source().PKOf(v))
}
