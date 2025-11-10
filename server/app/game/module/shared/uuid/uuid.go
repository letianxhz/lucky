package uuid

import (
	"sync"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"lucky/server/gen/msg"
	rpcCenter "lucky/server/pkg/rpc/center"
)

const (
	// DefaultUUIDName 默认 UUID 名称
	DefaultUUIDName = "default"
)

// UUIDPool UUID 池，管理本地预分配的 UUID
type UUIDPool struct {
	mu          sync.Mutex
	app         cfacade.IApplication
	name        string
	current     int64 // 当前可用的 UUID 值
	end         int64 // 当前范围的结束值
	initialized bool
}

// NewUUIDPool 创建新的 UUID 池
func NewUUIDPool(app cfacade.IApplication, name string) *UUIDPool {
	if name == "" {
		name = DefaultUUIDName
	}
	return &UUIDPool{
		app:  app,
		name: name,
	}
}

// Get 获取一个 UUID
func (p *UUIDPool) Get() (int64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 如果未初始化或已用完，从 center 申请
	if !p.initialized || p.current > p.end {
		if err := p.allocate(); err != nil {
			return 0, err
		}
	}

	uuid := p.current
	p.current++

	clog.Debugf("[UUIDPool] allocated UUID: name=%s, value=%d", p.name, uuid)
	return uuid, nil
}

// allocate 从 center 申请新的 UUID 范围
func (p *UUIDPool) allocate() error {
	clog.Infof("[UUIDPool] requesting UUID range from center: name=%s", p.name)

	range_, errCode := rpcCenter.AllocateUUID(p.app, p.name)
	if errCode != 0 {
		clog.Errorf("[UUIDPool] allocate UUID failed: name=%s, errCode=%d", p.name, errCode)
		return &AllocateError{Code: errCode}
	}

	if range_ == nil {
		clog.Errorf("[UUIDPool] allocate UUID returned nil: name=%s", p.name)
		return &AllocateError{Code: 1}
	}

	p.current = range_.Start
	p.end = range_.End
	p.initialized = true

	clog.Infof("[UUIDPool] allocated UUID range: name=%s, start=%d, end=%d", p.name, range_.Start, range_.End)
	return nil
}

// AllocateError UUID 分配错误
type AllocateError struct {
	Code int32
}

func (e *AllocateError) Error() string {
	return "failed to allocate UUID"
}
