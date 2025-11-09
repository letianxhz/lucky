package item

import (
	"errors"
	"lucky/server/app/game/db"
	"lucky/server/pkg/di"
	"sync"

	clog "github.com/cherry-game/cherry/logger"
)

// ItemModule 道具模块实现
// 类似 Java 项目的 ItemManager，负责道具相关的业务逻辑
type ItemModule struct {
	// 可以注入依赖，如缓存、数据库等
	// cache item.IItemCache `di:"auto"`
	// db    *db.DB          `di:"auto"`

	// 临时内存存储（用于测试，实际应该使用数据库/缓存）
	mu    sync.RWMutex
	items map[int64]map[int32]int64 // key: playerId, value: map[itemId]count
}

// init 初始化道具模块并注册到 di 容器
// 参考 claim ioc 的实现，简化注册逻辑
func init() {
	var v = &ItemModule{
		items: make(map[int64]map[int32]int64),
	}
	di.Register(v)
}

// GetItems 获取玩家所有道具
func (m *ItemModule) GetItems(playerId int64) (map[int32]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 如果内存中有数据，返回内存数据
	if playerItems, found := m.items[playerId]; found {
		// 返回副本，避免外部修改
		result := make(map[int32]int64, len(playerItems))
		for k, v := range playerItems {
			result[k] = v
		}
		return result, nil
	}

	// TODO: 从数据库或缓存获取
	// 如果没有内存数据，返回模拟数据（用于兼容现有逻辑）
	items := make(map[int32]int64)
	items[1001] = 100 // 模拟数据
	items[1002] = 50  // 模拟数据
	return items, nil
}

// GetItemCount 获取指定道具数量
func (m *ItemModule) GetItemCount(playerId int64, itemId int32) (int64, error) {
	items, err := m.GetItems(playerId)
	if err != nil {
		return 0, err
	}
	count, ok := items[itemId]
	if !ok {
		return 0, ErrItemNotFound
	}
	return count, nil
}

// CheckItem 检查道具数量是否足够
func (m *ItemModule) CheckItem(playerId int64, itemId int32, count int64) bool {
	currentCount, err := m.GetItemCount(playerId, itemId)
	if err != nil {
		return false
	}
	return currentCount >= count
}

// AddItem 添加道具
func (m *ItemModule) AddItem(playerId int64, itemId int32, count int64) error {
	if playerId <= 0 || itemId <= 0 || count <= 0 {
		return errors.New("invalid param")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 初始化玩家道具map
	if m.items[playerId] == nil {
		m.items[playerId] = make(map[int32]int64)
	}

	// 添加到内存存储（用于测试）
	m.items[playerId][itemId] += count

	// TODO: 1. 添加到数据库
	// TODO: 2. 更新缓存
	// TODO: 3. 发送事件（可选）
	clog.Debugf("[ItemModule] AddItem success. playerId=%d, itemId=%d, count=%d",
		playerId, itemId, count)
	return nil
}

// DeductItem 扣除道具
func (m *ItemModule) DeductItem(playerId int64, itemId int32, count int64) error {
	if playerId <= 0 || itemId <= 0 || count <= 0 {
		return errors.New("invalid param")
	}

	// 检查道具数量
	if !m.CheckItem(playerId, itemId, count) {
		return ErrItemNotEnough
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 从内存存储扣除（用于测试）
	if playerItems, found := m.items[playerId]; found {
		if currentCount, exists := playerItems[itemId]; exists {
			playerItems[itemId] = currentCount - count
			if playerItems[itemId] <= 0 {
				delete(playerItems, itemId)
			}
		}
	}

	// TODO: 1. 扣除数据库
	// TODO: 2. 更新缓存
	// TODO: 3. 发送事件（可选）
	clog.Debugf("[ItemModule] DeductItem success. playerId=%d, itemId=%d, count=%d",
		playerId, itemId, count)
	return nil
}

// BatchAddItems 批量添加道具
func (m *ItemModule) BatchAddItems(playerId int64, items map[int32]int64) error {
	if playerId <= 0 || len(items) == 0 {
		return errors.New("invalid param")
	}

	// TODO: 批量操作优化
	for itemId, count := range items {
		if err := m.AddItem(playerId, itemId, count); err != nil {
			return err
		}
	}
	return nil
}

// BatchDeductItems 批量扣除道具
func (m *ItemModule) BatchDeductItems(playerId int64, items map[int32]int64) error {
	if playerId <= 0 || len(items) == 0 {
		return errors.New("invalid param")
	}

	// 先检查所有道具是否足够
	for itemId, count := range items {
		if !m.CheckItem(playerId, itemId, count) {
			return ErrItemNotEnough
		}
	}

	// TODO: 批量操作优化
	for itemId, count := range items {
		if err := m.DeductItem(playerId, itemId, count); err != nil {
			return err
		}
	}
	return nil
}

// 避免未使用的导入
var _ = db.ItemTable{}
