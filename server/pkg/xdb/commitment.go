package xdb

import (
	"time"
)

// Commitment 提交对象接口
type Commitment interface {
	SourceInterface
	Merge(Commitment) bool
	Changes() FieldSet
	PrepareWrite() (interface{}, interface{})
	Lifecycle() Lifecycle
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

// RedoOptions 重做日志选项
type RedoOptions struct {
	Dir          string
	Enabled      bool
	SyncInterval time.Duration // if < 0: depend on os, if == 0: sync on write, if > 0: sync with period
}

var redoOptions *RedoOptions

func isRedoEnabled() bool {
	return redoOptions != nil && redoOptions.Enabled
}
