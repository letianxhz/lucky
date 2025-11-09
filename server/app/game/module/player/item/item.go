package item

import "errors"

// IItemModule 道具模块接口
// 定义道具相关的所有业务操作
type IItemModule interface {
	// GetItems 获取玩家所有道具
	GetItems(playerId int64) (map[int32]int64, error)

	// AddItem 添加道具
	AddItem(playerId int64, itemId int32, count int64) error

	// DeductItem 扣除道具
	DeductItem(playerId int64, itemId int32, count int64) error

	// CheckItem 检查道具数量是否足够
	CheckItem(playerId int64, itemId int32, count int64) bool

	// BatchAddItems 批量添加道具
	BatchAddItems(playerId int64, items map[int32]int64) error

	// BatchDeductItems 批量扣除道具
	BatchDeductItems(playerId int64, items map[int32]int64) error
}

// 错误定义
var (
	ErrItemNotFound  = errors.New("item not found")
	ErrItemNotEnough = errors.New("item not enough")
)
