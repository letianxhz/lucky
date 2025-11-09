package equipment

import (
	"errors"
	"lucky/server/app/game/module/player/item"
	"lucky/server/pkg/di"

	clog "github.com/cherry-game/cherry/logger"
)

// EquipmentModule 装备模块实现
// 类似 Java 项目的装备模块，通过依赖注入获取其他模块
type EquipmentModule struct {
	// 通过依赖注入获取其他模块，避免循环依赖
	// 使用接口类型，支持自动注入
	itemModule item.IItemModule `di:"auto"` // 自动注入道具模块
}

// init 初始化装备模块并注册到 di 容器
// 参考 claim ioc 的实现，简化注册逻辑
// 依赖通过 di.Resolve 在需要时自动注入
func init() {
	var v = &EquipmentModule{}
	di.Register(v)
}

// GetEquipments 获取玩家所有装备
func (m *EquipmentModule) GetEquipments(playerId int64) (map[int32]int32, error) {
	// TODO: 从数据库或缓存获取
	equipments := make(map[int32]int32)
	return equipments, nil
}

// EquipItem 装备道具（需要先扣除道具）
func (m *EquipmentModule) EquipItem(playerId int64, position int32, itemId int32) error {
	if playerId <= 0 || position <= 0 || itemId <= 0 {
		return errors.New("invalid param")
	}

	// 1. 检查道具是否存在（调用道具模块）
	if !m.itemModule.CheckItem(playerId, itemId, 1) {
		return errors.New("item not found")
	}

	// 2. 扣除道具
	err := m.itemModule.DeductItem(playerId, itemId, 1)
	if err != nil {
		return err
	}

	// 3. 装备道具
	// TODO: 更新数据库
	clog.Debugf("[EquipmentModule] EquipItem success. playerId=%d, position=%d, itemId=%d",
		playerId, position, itemId)

	return nil
}

// UnEquipItem 卸下装备（返回道具）
func (m *EquipmentModule) UnEquipItem(playerId int64, position int32) error {
	if playerId <= 0 || position <= 0 {
		return errors.New("invalid param")
	}

	// 1. 获取装备ID
	itemId, err := m.GetEquipmentByPosition(playerId, position)
	if err != nil {
		return err
	}

	// 2. 卸下装备
	// TODO: 更新数据库

	// 3. 返回道具（调用道具模块）
	err = m.itemModule.AddItem(playerId, itemId, 1)
	if err != nil {
		return err
	}

	clog.Debugf("[EquipmentModule] UnEquipItem success. playerId=%d, position=%d, itemId=%d",
		playerId, position, itemId)

	return nil
}

// GetEquipmentByPosition 获取指定位置的装备
func (m *EquipmentModule) GetEquipmentByPosition(playerId int64, position int32) (int32, error) {
	// TODO: 从数据库或缓存获取
	return 0, errors.New("equipment not found")
}
