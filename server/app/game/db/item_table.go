package db

import (
	clog "github.com/cherry-game/cherry/logger"
)

// ItemTable 道具表
type ItemTable struct {
	PlayerId int64 `gorm:"column:player_id;primary_key;comment:'角色id'" json:"playerId"`
	ItemId   int32 `gorm:"column:item_id;primary_key;comment:'道具id'" json:"itemId"`
	Count    int64 `gorm:"column:count;comment:'道具数量'" json:"count"`
}

func (*ItemTable) TableName() string {
	return "item"
}

// GetPlayerItems 获取玩家所有道具
func GetPlayerItems(playerId int64) (map[int32]int64, error) {
	// TODO: 从数据库查询
	// 临时实现：返回空 map
	items := make(map[int32]int64)
	return items, nil
}

// AddPlayerItem 添加玩家道具
func AddPlayerItem(playerId int64, itemId int32, count int64) error {
	// TODO: 更新数据库
	// 临时实现：只记录日志
	clog.Debugf("[ItemDB] AddPlayerItem. playerId=%d, itemId=%d, count=%d",
		playerId, itemId, count)
	return nil
}

// DeductPlayerItem 扣除玩家道具
func DeductPlayerItem(playerId int64, itemId int32, count int64) error {
	// TODO: 更新数据库
	// 临时实现：只记录日志
	clog.Debugf("[ItemDB] DeductPlayerItem. playerId=%d, itemId=%d, count=%d",
		playerId, itemId, count)
	return nil
}

// GetPlayerItemCount 获取玩家道具数量
func GetPlayerItemCount(playerId int64, itemId int32) (int64, error) {
	// TODO: 从数据库查询
	// 临时实现：返回 0
	return 0, nil
}
