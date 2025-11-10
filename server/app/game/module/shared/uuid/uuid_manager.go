package uuid

import (
	"sync"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
)

// Manager UUID 池管理器，管理多个 UUID 池
type Manager struct {
	mu    sync.RWMutex
	pools map[string]*UUIDPool
	app   cfacade.IApplication
}

var (
	globalManager *Manager
	once          sync.Once
)

// InitManager 初始化全局 UUID 管理器
func InitManager(app cfacade.IApplication) {
	once.Do(func() {
		globalManager = &Manager{
			pools: make(map[string]*UUIDPool),
			app:   app,
		}
		clog.Info("[UUIDManager] initialized")
	})
}

// GetManager 获取全局 UUID 管理器
func GetManager() *Manager {
	if globalManager == nil {
		clog.Warn("[UUIDManager] manager not initialized, using default app")
	}
	return globalManager
}

// GetPool 获取指定名称的 UUID 池
func (m *Manager) GetPool(name string) *UUIDPool {
	if name == "" {
		name = DefaultUUIDName
	}

	m.mu.RLock()
	pool, exists := m.pools[name]
	m.mu.RUnlock()

	if exists {
		return pool
	}

	// 创建新的池
	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if pool, exists := m.pools[name]; exists {
		return pool
	}

	pool = NewUUIDPool(m.app, name)
	m.pools[name] = pool

	clog.Infof("[UUIDManager] created new UUID pool: name=%s", name)
	return pool
}

// GetUUID 获取一个 UUID（使用默认池）
func (m *Manager) GetUUID() (int64, error) {
	return m.GetPool(DefaultUUIDName).Get()
}

// GetUUIDByName 获取指定名称的 UUID
func (m *Manager) GetUUIDByName(name string) (int64, error) {
	return m.GetPool(name).Get()
}

// GetUUID 全局函数：获取一个 UUID（使用默认池）
func GetUUID() (int64, error) {
	return GetManager().GetUUID()
}

// GetUUIDByName 全局函数：获取指定名称的 UUID
func GetUUIDByName(name string) (int64, error) {
	return GetManager().GetUUIDByName(name)
}
