package xdb

import (
	"context"
)

// CacheContext 缓存上下文
type CacheContext struct {
	cache map[PK]any
}

func newCacheContext() *CacheContext {
	return &CacheContext{
		cache: make(map[PK]any),
	}
}

func getCacheContext(ctx context.Context) *CacheContext {
	// TODO: 从 context 中获取缓存上下文
	// 这里简化实现，实际应该从 context.Value 中获取
	return nil
}

func cacheInvalidate(ctx context.Context, src *Source, m Model) {
	cc := getCacheContext(ctx)
	if cc != nil {
		delete(cc.cache, src.PKOf(m))
	}
}

// BufferContext 缓冲上下文
type BufferContext struct {
	safeMode bool
	rs       []MutableRecord
	ms       map[MutableRecord]Model
}

func newBufferContext() *BufferContext {
	return &BufferContext{
		ms: make(map[MutableRecord]Model),
	}
}

func getBufferContext(ctx context.Context) *BufferContext {
	// TODO: 从 context 中获取缓冲上下文
	// 这里简化实现，实际应该从 context.Value 中获取
	return nil
}

// IsWithORM 检查是否使用 ORM
func IsWithORM(ctx context.Context) bool {
	return getBufferContext(ctx) != nil
}

// Stash 暂存记录
func Stash(ctx context.Context, r MutableRecord) {
	if r.GetHeader().IsExpired() {
		return
	}

	b := getBufferContext(ctx)
	if b == nil {
		return
	}

	b.Stash(r)
}

// EnableSafeMode 开启安全模式
func EnableSafeMode(ctx context.Context) {
	b := getBufferContext(ctx)
	if b != nil {
		b.EnableSafeMode()
	}
}

// SafeMode 是否开启了安全模式
func SafeMode(ctx context.Context) bool {
	b := getBufferContext(ctx)
	if b == nil {
		return false
	}
	return b.SafeMode()
}

func (b *BufferContext) EnableSafeMode() {
	b.safeMode = true
}

func (b *BufferContext) SafeMode() bool {
	return b.safeMode
}

func (b *BufferContext) Stash(r MutableRecord) {
	if r.Committing() || b.ms[r] != nil {
		return
	}

	m := WrappingModel(r)
	b.ms[r] = m
	b.rs = append(b.rs, r)
}

func (b *BufferContext) Flush(ctx context.Context) {
	for len(b.rs) != 0 {
		r := b.rs[0]
		b.rs = b.rs[1:]

		m := b.ms[r]
		delete(b.ms, r)

		if m == nil {
			Save(ctx, r)
		} else {
			Save(ctx, m.(MutableRecord))
		}
	}
}

// LockContext 锁上下文接口
type LockContext interface {
	Lock(m Model, lt lockType, onlyAlive bool, simulate bool) bool
	UnLock(m Model, lt lockType)
}

var lockContextKey = struct{}{}

func getLockContext() LockContext {
	// TODO: 从 context 中获取锁上下文
	// 这里简化实现
	return nil
}

func setLockContext(envReady func() bool) func() {
	// TODO: 设置锁上下文
	return func() {}
}
