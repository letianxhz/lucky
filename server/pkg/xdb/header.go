package xdb

import (
	"context"
	"math"
)

const FieldSetAll = FieldSet(math.MaxUint64)
const FieldSetEmpty = FieldSet(0)

type Field int8

type FieldSet uint64

// MakeFieldSet 创建字段集合
func MakeFieldSet(fields ...Field) FieldSet {
	return FieldSetEmpty.Add(fields...)
}

// Add 添加字段（注意：不会修改原值）
func (fs FieldSet) Add(fields ...Field) FieldSet {
	nfs := fs
	for _, f := range fields {
		nfs = nfs | (1 << f)
	}
	return nfs
}

// Del 删除字段（不会改变原值）
func (fs FieldSet) Del(fields ...Field) FieldSet {
	return fs.Subtract(MakeFieldSet(fields...))
}

// Contains 检查是否包含指定字段
func (fs FieldSet) Contains(fields ...Field) bool {
	for _, f := range fields {
		if fs&(1<<f) == 0 {
			return false
		}
	}
	return true
}

// ContainsAny 检查是否包含任意一个字段
func (fs FieldSet) ContainsAny(fields ...Field) bool {
	for _, f := range fields {
		if fs&(1<<f) != 0 {
			return true
		}
	}
	return false
}

// ContainsExact 检查是否完全匹配字段集合
func (fs FieldSet) ContainsExact(fields ...Field) bool {
	return fs == MakeFieldSet(fields...)
}

// Union 并集
func (fs FieldSet) Union(other FieldSet) FieldSet {
	return fs | other
}

// Intersect 交集
func (fs FieldSet) Intersect(other FieldSet) FieldSet {
	return fs & other
}

// Subtract 差集
func (fs FieldSet) Subtract(other FieldSet) FieldSet {
	return fs & ^other
}

// Debug 调试，返回字段描述
func (fs FieldSet) Debug(o interface{}) []*FieldDesc {
	src := SourceOf(o)
	if src == nil {
		return nil
	}

	var ret []*FieldDesc
	for _, field := range src.Fields {
		if fs.Contains(field.Code) {
			ret = append(ret, field)
		}
	}
	return ret
}

const (
	FlagDirty int16 = 1 << iota
	FlagCommitting
	FlagLoadComplete
	FlagTmp
	FlagExpired
	FlagReadonly
	FlagMirror
)

// Header 记录头信息
type Header struct {
	lifecycle   Lifecycle
	flags       int16
	SavingIndex int32
	changes     FieldSet
}

// Init 初始化头信息
func (h *Header) Init(lifecycle Lifecycle) {
	h.lifecycle = lifecycle
	h.setFlag(FlagDirty, lifecycle == LifecycleNew)
}

func (h *Header) setFlag(flag int16, value bool) {
	if value {
		h.flags = h.flags | flag
	} else {
		h.flags = h.flags & ^flag
	}
}

func (h *Header) getFlag(flag int16) bool {
	return h.flags&flag != 0
}

// SetChanged 设置变更字段
func (h *Header) SetChanged(fields ...Field) {
	switch h.lifecycle {
	case LifecycleNew:
		return
	case LifecycleDeleted:
		panic("record has been deleted")
	case LifecycleUnavailable:
		panic("record is unavailable")
	}

	h.changes = h.changes.Add(fields...)
	h.setFlag(FlagDirty, true)
}

// SetChanges 设置变更集合
func (h *Header) SetChanges(fields FieldSet) {
	h.changes = fields
}

// Dirty 检查是否有未保存的变更
func (h *Header) Dirty() bool {
	return h.getFlag(FlagDirty)
}

// Lifecycle 获取生命周期
func (h *Header) Lifecycle() Lifecycle {
	return h.lifecycle
}

// Changed 检查字段是否变更
func (h *Header) Changed(field Field) bool {
	return h.changes.Contains(field)
}

// Changes 获取变更集合
func (h *Header) Changes() FieldSet {
	return h.changes
}

// MarkAsDeleted 标记为已删除
func (h *Header) MarkAsDeleted(ctx context.Context) bool {
	prev := h.lifecycle

	switch prev {
	case LifecycleNew:
		h.lifecycle = LifecycleUnavailable
		h.setFlag(FlagDirty, false)
		return true
	case LifecycleNormal:
		h.lifecycle = LifecycleDeleted
		h.setFlag(FlagDirty, true)
		return true
	default:
		return false
	}
}

// Merge 合并头信息（当不可合并时，返回false）
func (h *Header) Merge(other *Header) bool {
	switch h.lifecycle {
	case LifecycleNew:
		if other.lifecycle == LifecycleDeleted || other.lifecycle == LifecycleUnavailable {
			h.lifecycle = LifecycleUnavailable
		}
		return true
	case LifecycleNormal:
		if other.lifecycle == LifecycleDeleted || other.lifecycle == LifecycleUnavailable {
			h.lifecycle = LifecycleDeleted
		} else {
			h.changes = h.changes.Union(other.changes)
		}
		return true
	default:
		// LifecycleDeleted, LifecycleUnavailable不可合并
		return false
	}
}

// Committing 检查是否正在提交
func (h *Header) Committing() bool {
	return h.getFlag(FlagCommitting)
}

// Commit 开始提交，返回完成函数
func (h *Header) Commit() func() {
	h.setFlag(FlagCommitting, true)

	return func() {
		switch h.lifecycle {
		case LifecycleDeleted:
			h.lifecycle = LifecycleUnavailable
		case LifecycleNew:
			h.lifecycle = LifecycleNormal
		}

		h.changes = 0
		h.setFlag(FlagDirty, false)
		h.setFlag(FlagCommitting, false)
	}
}

// Loading 检查是否正在加载
func (h *Header) Loading() bool {
	return !h.getFlag(FlagLoadComplete)
}

// LoadComplete 标记加载完成
func (h *Header) LoadComplete() {
	h.setFlag(FlagLoadComplete, true)
}

// IsTmp 检查是否为临时记录
func (h *Header) IsTmp() bool {
	return h.getFlag(FlagTmp)
}

// SetTmp 设置临时标记
func (h *Header) SetTmp(tmp bool) {
	h.setFlag(FlagTmp, tmp)
}

// IsExpired 检查是否已过期
func (h *Header) IsExpired() bool {
	return h.getFlag(FlagExpired)
}

// SetExpired 设置过期标记
func (h *Header) SetExpired(expired bool) {
	h.setFlag(FlagExpired, expired)
}

// EnableReadonly 启用只读
func (h *Header) EnableReadonly() {
	h.setFlag(FlagReadonly, true)
}

// Readonly 检查是否只读
func (h *Header) Readonly() bool {
	return h.getFlag(FlagReadonly)
}

// EnableWrite 启用写入
func (h *Header) EnableWrite() {
	h.setFlag(FlagReadonly, false)
}

// EnableMirror 启用镜像
func (h *Header) EnableMirror() {
	h.setFlag(FlagMirror, true)
}

// IsMirror 检查是否为镜像
func (h *Header) IsMirror() bool {
	return h.getFlag(FlagMirror)
}

// ValidateAffinity 验证亲和性（用于验证是否在同一 goroutine 中访问）
// 默认实现返回 false，表示不进行亲和性验证
// 如果需要亲和性验证，可以在生成的 Record 代码中覆盖此方法
func (h *Header) ValidateAffinity() bool {
	return false
}
