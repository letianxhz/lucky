package xdb

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	LifecycleUnavailable Lifecycle = iota
	LifecycleNew
	LifecycleNormal
	LifecycleDeleted
)

type Lifecycle int8

func (l Lifecycle) String() string {
	switch l {
	case LifecycleNew:
		return "new"
	case LifecycleNormal:
		return "normal"
	case LifecycleDeleted:
		return "deleted"
	default:
		return "unavailable"
	}
}

// Record 表示一个数据库记录
type Record interface {
	json.Marshaler
	json.Unmarshaler
	fmt.Stringer
	SourceInterface
	XId() string // 用于日志
	Lifecycle() Lifecycle
	Snapshoot() interface{} // 快照，返回 proto.Message 或类似结构
	XVersion() int64
	GetHeader() *Header
}

// MutableRecord 表示可变的记录
type MutableRecord interface {
	Record
	Init(ctx context.Context, data interface{}) error
	Update(ctx context.Context, changes interface{}, fs FieldSet) error
	Delete(ctx context.Context) bool
	Commit(ctx context.Context) (Commitment, FieldSet)
	Committing() bool
	Dirty() bool
	SavingIndex() int32
	SetSavingIndex(int32)
}

// Model 表示一个模型，包含监听器
// 注意：在 actor 模型中，不需要锁，因为每个 actor 在单线程环境中运行
type Model interface {
	Record
	Listener
	ValidateAffinity() bool
}

// Loggable 可记录日志的接口
type Loggable interface {
	Record
	Loggable(lif Lifecycle, changes FieldSet) bool
}

// Volatile 易失性接口，表示数据可能很快过期
type Volatile interface {
	Model
	IsVolatile() bool
}

var volatileType = reflect.TypeOf((*Volatile)(nil)).Elem()

// Listener 监听器接口，用于模型生命周期事件
type Listener interface {
	OnCreate(ctx context.Context)
	OnLoad(ctx context.Context)
	OnUpdate(ctx context.Context, fs FieldSet)
	OnDelete(ctx context.Context)
	OnReload(ctx context.Context)
	OnRefresh(ctx context.Context)
}

// SourceInterface 源接口
type SourceInterface interface {
	Source() *Source
}

// PKOf 获取记录的主键
func PKOf(r SourceInterface) PK {
	return r.Source().PKOf(r)
}

// WrappingModel 从 Record 获取 Model
func WrappingModel(record Record) Model {
	if record == nil {
		return nil
	}

	m, isModel := record.(Model)
	if isModel {
		return m
	}

	src := record.Source()
	if src == nil || src.ModelType == nil {
		return nil
	}

	// 使用反射获取嵌入的 Model
	val := reflect.ValueOf(record)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			if model, ok := field.Interface().(Model); ok {
				return model
			}
		}
	}

	return nil
}

// IsDeleted 检查记录是否已删除（需要持有对象锁）
func IsDeleted(record Record) bool {
	return record.Lifecycle() == LifecycleDeleted || record.Lifecycle() == LifecycleUnavailable
}

// IsExpired 检查记录是否已过期
func IsExpired(record Record) bool {
	return record.GetHeader().IsExpired()
}
