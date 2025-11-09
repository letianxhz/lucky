package equipment

// IEquipmentModule 装备模块接口
type IEquipmentModule interface {
	// GetEquipments 获取玩家所有装备
	GetEquipments(playerId int64) (map[int32]int32, error) // key:位置, value:装备ID

	// EquipItem 装备道具（需要先扣除道具）
	EquipItem(playerId int64, position int32, itemId int32) error

	// UnEquipItem 卸下装备（返回道具）
	UnEquipItem(playerId int64, position int32) error

	// GetEquipmentByPosition 获取指定位置的装备
	GetEquipmentByPosition(playerId int64, position int32) (int32, error)
}
