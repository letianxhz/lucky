package item

import (
	"lucky/server/app/game/db"
	"lucky/server/app/game/module/shared/handler"
	"lucky/server/gen/msg"
	"lucky/server/pkg/di"

	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
	"lucky/server/pkg/code"
)

// init 注册道具模块的消息处理器
func init() {
	var h = &itemHandler{}
	di.Register(h)

	// 使用新的类型安全注册机制（V3 泛型版本，无反射，高性能）
	handler.RegisterHandler(handler.ActorTypePlayer, "buyItem", h.OnBuyItem)
}

type itemHandler struct {
	item IItemModule `di:"auto"`
}

// 道具价格表（包级变量，避免每次函数调用都创建 map）
// TODO: 应该从配置表加载
var itemPriceMap = map[int32]int64{
	1001: 100,
	1002: 200,
}

// OnBuyItem 购买道具消息处理器
// 签名: func(session *cproto.Session, req *msg.BuyItemRequest) (*msg.BuyItemResponse, error)
// 注册层会自动进行类型转换，这里直接接收具体类型，无需类型断言
func (h *itemHandler) OnBuyItem(session *cproto.Session, req *msg.BuyItemRequest) (*msg.BuyItemResponse, error) {
	// 参数验证（合并多个判断，减少分支预测失败）
	if req.ItemId <= 0 || req.Count <= 0 || req.PayType <= 0 || req.PayType > 3 {
		return nil, handler.NewErrorWithCode(code.ShopItemInvalidParam)
	}

	// 使用 map 查找价格（性能优化：O(1) 时间复杂度，比 if-else 快）
	itemPrice, ok := itemPriceMap[req.ItemId]
	if !ok {
		return nil, handler.NewErrorWithCode(code.ShopItemNotFound)
	}

	// 计算总价格
	totalCost := itemPrice * int64(req.Count)

	// TODO: 检查玩家货币是否足够
	// TODO: 扣除货币

	// 获取玩家ID
	playerId := db.GetPlayerIdWithUID(session.Uid)
	if playerId <= 0 {
		return nil, handler.NewErrorWithCode(code.PlayerNotLogin)
	}

	// 添加道具
	err := h.item.AddItem(playerId, req.ItemId, int64(req.Count))
	if err != nil {
		clog.Warnf("[ItemHandler] AddItem failed. playerId=%d, itemId=%d, count=%d, err=%v",
			playerId, req.ItemId, req.Count, err)
		return nil, handler.NewErrorWithCode(code.ShopItemBuyFail)
	}

	// 构建响应（优化：直接创建 map，避免额外的内存分配）
	items := make(map[int32]int64, 1)
	items[req.ItemId] = int64(req.Count)

	return &msg.BuyItemResponse{
		ItemId:     req.ItemId,
		Count:      req.Count,
		PayType:    req.PayType,
		CostAmount: totalCost,
		Items:      items,
	}, nil
}
