# Game 服务模块化架构设计

## 设计目标

1. **分层合理**: 清晰的层次结构，职责分明
2. **易扩展**: 新增模块简单，不影响现有代码
3. **模块间调用**: 简洁、类型安全、易维护

## 目录结构

```
app/game/
├── actor/                    # Actor 层（消息处理）
│   ├── player/              # 玩家 Actor
│   │   ├── actor_player.go  # 玩家 Actor 主逻辑
│   │   └── actor_players.go # 玩家管理 Actor
│   └── ...
├── module/                   # 业务模块层（核心业务逻辑）
│   ├── item/                # 道具模块
│   │   ├── item.go         # 道具接口定义
│   │   ├── item_impl.go    # 道具实现
│   │   └── item_cache.go   # 道具缓存
│   ├── equipment/           # 装备模块
│   │   ├── equipment.go
│   │   ├── equipment_impl.go
│   │   └── equipment_cache.go
│   ├── skill/               # 技能模块
│   │   ├── skill.go
│   │   ├── skill_impl.go
│   │   └── skill_cache.go
│   └── online.go            # 在线玩家管理（共享模块）
├── service/                  # 服务层（跨模块服务）
│   ├── player_service.go    # 玩家服务（聚合各模块）
│   └── ...
├── db/                       # 数据访问层
│   ├── player_table.go
│   ├── item_table.go
│   ├── equipment_table.go
│   └── ...
└── pkg/                      # 共享包
    ├── event/               # 事件定义
    └── ...
```

## 架构设计

### 1. 模块接口设计

每个业务模块都实现统一的接口，便于管理和扩展：

```go
// module/item/item.go
package item

import (
    "lucky/server/pkg/pb"
)

// IItemModule 道具模块接口
type IItemModule interface {
    // 获取道具列表
    GetItems(playerId int64) (map[int32]int64, error)
    
    // 添加道具
    AddItem(playerId int64, itemId int32, count int64) error
    
    // 扣除道具
    DeductItem(playerId int64, itemId int32, count int64) error
    
    // 检查道具数量
    CheckItem(playerId int64, itemId int32, count int64) bool
    
    // 批量操作
    BatchAddItems(playerId int64, items map[int32]int64) error
}
```

### 2. 模块实现

```go
// module/item/item_impl.go
package item

import (
    "lucky/server/app/game/db"
    "lucky/server/pkg/code"
    clog "github.com/cherry-game/cherry/logger"
)

type ItemModule struct {
    // 可以注入依赖，如缓存、数据库等
}

func NewItemModule() IItemModule {
    return &ItemModule{}
}

func (m *ItemModule) GetItems(playerId int64) (map[int32]int64, error) {
    // 1. 从缓存获取
    // 2. 缓存未命中，从数据库加载
    // 3. 更新缓存
    // 4. 返回结果
}

func (m *ItemModule) AddItem(playerId int64, itemId int32, count int64) error {
    // 1. 参数验证
    // 2. 更新数据库
    // 3. 更新缓存
    // 4. 发送事件（可选）
}
```

### 3. 模块管理器

统一管理所有模块，提供模块获取和初始化：

```go
// module/manager.go
package module

import (
    "lucky/server/app/game/module/item"
    "lucky/server/app/game/module/equipment"
    "lucky/server/app/game/module/skill"
)

// ModuleManager 模块管理器
type ModuleManager struct {
    ItemModule      item.IItemModule
    EquipmentModule equipment.IEquipmentModule
    SkillModule     skill.ISkillModule
}

var manager *ModuleManager

// InitModules 初始化所有模块
func InitModules() {
    manager = &ModuleManager{
        ItemModule:      item.NewItemModule(),
        EquipmentModule: equipment.NewEquipmentModule(),
        SkillModule:     skill.NewSkillModule(),
    }
}

// GetModules 获取模块管理器
func GetModules() *ModuleManager {
    return manager
}
```

### 4. Actor 中使用模块

在 Actor 中通过模块管理器获取模块，调用业务逻辑：

```go
// actor/player/actor_player.go
package player

import (
    "lucky/server/app/game/module"
    "lucky/server/pkg/pb"
)

func (p *actorPlayer) buyItem(session *cproto.Session, req *pb.BuyItemRequest) {
    // 1. 参数验证
    // 2. 获取道具模块
    itemModule := module.GetModules().ItemModule
    
    // 3. 调用模块方法
    err := itemModule.AddItem(p.playerId, req.ItemId, int64(req.Count))
    if err != nil {
        p.ResponseCode(session, code.ShopItemBuyFail)
        return
    }
    
    // 4. 构建响应
    response := &pb.BuyItemResponse{...}
    p.Response(session, response)
}
```

### 5. 模块间调用

模块之间通过接口调用，避免直接依赖：

```go
// module/equipment/equipment_impl.go
package equipment

import (
    "lucky/server/app/game/module"
)

func (m *EquipmentModule) EquipItem(playerId int64, itemId int32) error {
    // 1. 检查道具是否存在（调用道具模块）
    itemModule := module.GetModules().ItemModule
    if !itemModule.CheckItem(playerId, itemId, 1) {
        return errors.New("item not found")
    }
    
    // 2. 扣除道具
    err := itemModule.DeductItem(playerId, itemId, 1)
    if err != nil {
        return err
    }
    
    // 3. 装备道具
    // ...
}
```

## 完整示例

### 道具模块完整实现

```go
// module/item/item.go
package item

type IItemModule interface {
    GetItems(playerId int64) (map[int32]int64, error)
    AddItem(playerId int64, itemId int32, count int64) error
    DeductItem(playerId int64, itemId int32, count int64) error
    CheckItem(playerId int64, itemId int32, count int64) bool
}

// module/item/item_impl.go
package item

import (
    "lucky/server/app/game/db"
    "lucky/server/pkg/code"
)

type ItemModule struct{}

func NewItemModule() IItemModule {
    return &ItemModule{}
}

func (m *ItemModule) GetItems(playerId int64) (map[int32]int64, error) {
    // 实现逻辑
    return db.GetPlayerItems(playerId)
}

func (m *ItemModule) AddItem(playerId int64, itemId int32, count int64) error {
    return db.AddPlayerItem(playerId, itemId, count)
}

// ... 其他方法
```

### Actor 中使用

```go
// actor/player/actor_player.go
func (p *actorPlayer) buyItem(session *cproto.Session, req *pb.BuyItemRequest) {
    // 参数验证
    if req.ItemId <= 0 || req.Count <= 0 {
        p.ResponseCode(session, code.ShopItemInvalidParam)
        return
    }
    
    // 获取模块
    itemModule := module.GetModules().ItemModule
    
    // 调用模块方法
    err := itemModule.AddItem(p.playerId, req.ItemId, int64(req.Count))
    if err != nil {
        p.ResponseCode(session, code.ShopItemBuyFail)
        return
    }
    
    // 构建响应
    response := &pb.BuyItemResponse{
        ItemId:  req.ItemId,
        Count:   req.Count,
        Items:   map[int32]int64{req.ItemId: int64(req.Count)},
    }
    p.Response(session, response)
}
```

## 优势

1. **分层清晰**: Actor 层只负责消息处理，业务逻辑在 Module 层
2. **易扩展**: 新增模块只需实现接口，注册到管理器
3. **模块解耦**: 通过接口调用，模块间依赖清晰
4. **类型安全**: 接口定义明确，编译期检查
5. **易测试**: 模块可以独立测试，Mock 接口简单

## 扩展性

### 新增模块步骤

1. 在 `module/` 下创建新模块目录
2. 定义接口 `IXXXModule`
3. 实现接口 `XXXModule`
4. 在 `ModuleManager` 中注册
5. 在 Actor 中使用

### 示例：新增技能模块

```go
// 1. 创建 module/skill/skill.go
package skill

type ISkillModule interface {
    LearnSkill(playerId int64, skillId int32) error
    UpgradeSkill(playerId int64, skillId int32) error
}

// 2. 实现 module/skill/skill_impl.go
// 3. 在 ModuleManager 中注册
// 4. 在 Actor 中使用
```

