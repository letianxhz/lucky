package item

import (
	"lucky/server/pkg/di"
	"testing"
)

func TestItemModule_AddItem(t *testing.T) {
	// 重置并初始化 di 容器
	di.Reset()
	// 手动注册模块（需要初始化 items map）
	var itemModule = &ItemModule{
		items: make(map[int64]map[int32]int64),
	}
	di.Register(itemModule)
	// 注册接口实现关系
	di.RegisterImplementation((*IItemModule)(nil), itemModule)
	di.MustInitialize()

	// 获取 ItemModule 实例
	moduleInstance, err := di.GetByType((*IItemModule)(nil))
	if err != nil {
		t.Fatalf("Failed to get ItemModule: %v", err)
	}
	itemModuleInstance := moduleInstance.(IItemModule)

	// 测试添加道具
	playerId := int64(1001)
	itemId := int32(1001)
	count := int64(10)

	err = itemModuleInstance.AddItem(playerId, itemId, count)
	if err != nil {
		t.Errorf("AddItem failed: %v", err)
	}

	// 验证道具已添加
	items, err := itemModuleInstance.GetItems(playerId)
	if err != nil {
		t.Errorf("GetItems failed: %v", err)
	}

	itemCount := items[itemId]
	if itemCount < count {
		t.Errorf("Expected item count >= %d, got %d", count, itemCount)
	}

	t.Logf("AddItem success: playerId=%d, itemId=%d, count=%d, currentCount=%d",
		playerId, itemId, count, itemCount)
}

func TestItemModule_DeductItem(t *testing.T) {
	// 重置并初始化 di 容器
	di.Reset()
	// 手动注册模块（需要初始化 items map）
	var itemModule = &ItemModule{
		items: make(map[int64]map[int32]int64),
	}
	di.Register(itemModule)
	// 注册接口实现关系
	di.RegisterImplementation((*IItemModule)(nil), itemModule)
	di.MustInitialize()

	// 获取 ItemModule 实例
	moduleInstance, err := di.GetByType((*IItemModule)(nil))
	if err != nil {
		t.Fatalf("Failed to get ItemModule: %v", err)
	}
	itemModuleInstance := moduleInstance.(IItemModule)

	// 先添加道具
	playerId := int64(1001)
	itemId := int32(1001)
	addCount := int64(20)
	deductCount := int64(5)

	err = itemModuleInstance.AddItem(playerId, itemId, addCount)
	if err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}

	// 测试扣除道具
	err = itemModuleInstance.DeductItem(playerId, itemId, deductCount)
	if err != nil {
		t.Errorf("DeductItem failed: %v", err)
	}

	// 验证道具数量
	items, err := itemModuleInstance.GetItems(playerId)
	if err != nil {
		t.Errorf("GetItems failed: %v", err)
	}

	itemCount := items[itemId]
	expectedCount := addCount - deductCount
	if itemCount != expectedCount {
		t.Errorf("Expected item count %d, got %d", expectedCount, itemCount)
	}

	t.Logf("DeductItem success: playerId=%d, itemId=%d, deducted=%d, remaining=%d",
		playerId, itemId, deductCount, itemCount)
}

func TestItemModule_CheckItem(t *testing.T) {
	// 重置并初始化 di 容器
	di.Reset()
	// 手动注册模块（需要初始化 items map）
	var itemModule = &ItemModule{
		items: make(map[int64]map[int32]int64),
	}
	di.Register(itemModule)
	// 注册接口实现关系
	di.RegisterImplementation((*IItemModule)(nil), itemModule)
	di.MustInitialize()

	// 获取 ItemModule 实例
	moduleInstance, err := di.GetByType((*IItemModule)(nil))
	if err != nil {
		t.Fatalf("Failed to get ItemModule: %v", err)
	}
	itemModuleInstance := moduleInstance.(IItemModule)

	// 先添加道具
	playerId := int64(1001)
	itemId := int32(1001)
	count := int64(10)

	err = itemModuleInstance.AddItem(playerId, itemId, count)
	if err != nil {
		t.Fatalf("AddItem failed: %v", err)
	}

	// 测试检查道具 - 足够
	hasEnough := itemModuleInstance.CheckItem(playerId, itemId, 5)
	if !hasEnough {
		t.Error("CheckItem should return true for sufficient items")
	}

	// 测试检查道具 - 不足
	hasEnough = itemModuleInstance.CheckItem(playerId, itemId, 20)
	if hasEnough {
		t.Error("CheckItem should return false for insufficient items")
	}

	t.Logf("CheckItem test passed")
}

func TestItemModule_RegisterHandlers(t *testing.T) {
	// 测试消息处理器注册（通过 init() 方式）
	// handler.go 中的 init() 函数会在包导入时自动执行
	// 这里只验证模块可以正常获取
	di.Reset()
	// 手动注册模块（需要初始化 items map）
	var itemModule = &ItemModule{
		items: make(map[int64]map[int32]int64),
	}
	di.Register(itemModule)
	// 注册接口实现关系
	di.RegisterImplementation((*IItemModule)(nil), itemModule)
	di.MustInitialize()

	// 获取 ItemModule 实例
	moduleInstance, err := di.GetByType((*IItemModule)(nil))
	if err != nil {
		t.Fatalf("Failed to get ItemModule: %v", err)
	}
	if moduleInstance == nil {
		t.Error("ItemModule instance is nil")
	}

	t.Logf("ItemModule handler registration test passed (handler.go init() will register handlers)")
}

func TestItemModule_BatchOperations(t *testing.T) {
	// 重置并初始化 di 容器
	di.Reset()
	// 手动注册模块（需要初始化 items map）
	var itemModule = &ItemModule{
		items: make(map[int64]map[int32]int64),
	}
	di.Register(itemModule)
	// 注册接口实现关系
	di.RegisterImplementation((*IItemModule)(nil), itemModule)
	di.MustInitialize()

	// 获取 ItemModule 实例
	moduleInstance, err := di.GetByType((*IItemModule)(nil))
	if err != nil {
		t.Fatalf("Failed to get ItemModule: %v", err)
	}
	itemModuleInstance := moduleInstance.(IItemModule)

	playerId := int64(1001)
	items := map[int32]int64{
		1001: 10,
		1002: 20,
	}

	// 测试批量添加
	err = itemModuleInstance.BatchAddItems(playerId, items)
	if err != nil {
		t.Errorf("BatchAddItems failed: %v", err)
	}

	// 验证道具已添加
	allItems, err := itemModuleInstance.GetItems(playerId)
	if err != nil {
		t.Errorf("GetItems failed: %v", err)
	}

	for itemId, count := range items {
		itemCount := allItems[itemId]
		if itemCount < count {
			t.Errorf("Expected item %d count >= %d, got %d", itemId, count, itemCount)
		}
	}

	// 测试批量扣除
	err = itemModuleInstance.BatchDeductItems(playerId, items)
	if err != nil {
		t.Errorf("BatchDeductItems failed: %v", err)
	}

	t.Logf("BatchOperations test passed")
}
