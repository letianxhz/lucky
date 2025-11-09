package xdb

import (
	"context"
	"reflect"
	"time"
)

// Driver 驱动接口
type Driver interface {
	Name() string
	Init(ctx context.Context, config interface{}) error
	Validate(src *Source) error
	NewDao(ctx context.Context, option interface{}) (Dao, error)
	ExtendType(base reflect.Type, extension reflect.Type)
}

var drivers = map[string]Driver{}

func init() {
	RegisterDriver(noStorage{})
}

// RegisterDriver 注册驱动
func RegisterDriver(driver Driver) {
	drivers[driver.Name()] = driver
}

// GetDriver 获取驱动
func GetDriver(name string) Driver {
	return drivers[name]
}

// Dao 数据访问对象接口
type Dao interface {
	Table(src *Source) Table
}

// RecordCursor 记录游标接口
type RecordCursor interface {
	Next(ctx context.Context) bool
	Decode(val interface{}) error
	All(ctx context.Context, results interface{}) error
	Close(ctx context.Context) error
}

// Table 表接口
type Table interface {
	Recover(ctx context.Context, commitments []Commitment) error
	// Save 返回false，意味着保存失败，且放弃重试，通常只应发生在系统关闭的过程中，否则都应该返回true
	Save(ctx context.Context, commitments []Commitment, writeTimeout time.Duration, retryInterval time.Duration, running func() bool) bool
	Fetch(ctx context.Context, onlyOne bool, pk PK) (RecordCursor, error)
	FetchMulti(ctx context.Context, pks []PK) (RecordCursor, error)
	Find(ctx context.Context, filter interface{}) (RecordCursor, error)
}

// noStorage 无存储驱动
type noStorage struct{}

func (n noStorage) Name() string {
	return "none"
}

func (n noStorage) Init(context.Context, interface{}) error {
	return nil
}

func (n noStorage) Validate(*Source) error {
	return nil
}

func (n noStorage) NewDao(context.Context, interface{}) (Dao, error) {
	return nil, nil
}

func (n noStorage) ExtendType(base reflect.Type, extension reflect.Type) {
}

// NoStorageTable 无存储表
type NoStorageTable struct {
}

func (t NoStorageTable) Recover(context.Context, []Commitment) error {
	return nil
}

func (t NoStorageTable) Save(context.Context, []Commitment, time.Duration, time.Duration, func() bool) bool {
	return true
}

func (t NoStorageTable) Fetch(context.Context, bool, PK) (RecordCursor, error) {
	return NoStorageRecordCursor{}, nil
}

func (t NoStorageTable) FetchMulti(context.Context, []PK) (RecordCursor, error) {
	return NoStorageRecordCursor{}, nil
}

func (t NoStorageTable) Find(context.Context, interface{}) (RecordCursor, error) {
	return NoStorageRecordCursor{}, nil
}

// NoStorageRecordCursor 无存储记录游标
type NoStorageRecordCursor struct {
}

func (n NoStorageRecordCursor) Next(context.Context) bool {
	return false
}

func (n NoStorageRecordCursor) Decode(interface{}) error {
	return nil
}

func (n NoStorageRecordCursor) All(context.Context, interface{}) error {
	return nil
}

func (n NoStorageRecordCursor) Close(context.Context) error {
	return nil
}
