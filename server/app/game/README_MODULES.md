# Game 服务模块化架构

## 概述

Game 服务采用模块化架构设计，将不同的业务逻辑（道具、装备、技能等）封装为独立的模块，通过统一的模块管理器进行管理和调用。

## 目录结构

```
app/game/
├── actor/                    # Actor 层（消息处理）
│   └── player/              # 玩家 Actor
│       ├── actor_player.go  # 玩家 Actor 主逻辑
│       └── actor_players.go # 玩家管理 Actor
├── module/                   # 业务模块层（核心业务逻辑）
│   ├── manager.go           # 模块管理器（统一管理所有模块）
│   ├── item/                # 道具模块
│   │   ├── item.go         # 道具接口定义
│   │   └── item_impl.go    # 道具实现
│   ├── equipment/           # 装备模块
│   │   ├── equipment.go
│   │   └── equipment_impl.go
│   ├── skill/               # 技能模块（示例）
│   │   ├── skill.go
│   │   └── skill_impl.go
│   └── online.go            # 在线玩家管理（共享模块）
├── db/                       # 数据访问层
│   ├── player_table.go
│   ├── item_table.go
│   └── equipment_table.go
└── pkg/                      # 共享包
```

## 架构设计原则

### 1. 分层清晰

- **Actor 层**: 只负责消息接收、参数验证、调用模块、构建响应
- **Module 层**: 负责核心业务逻辑
- **DB 层**: 负责数据访问

### 2. 接口驱动

每个模块都定义接口，实现与使用分离：

```go
// 接口定义
type IItemModule interface {
    AddItem(playerId int64, itemId int32, count int64) error
    // ...
}

// 实现
type ItemModule struct{}
func (m *ItemModule) AddItem(...) error { ... }
```

### 3. 统一管理

通过 `ModuleManager` 统一管理所有模块：

```go
// 获取模块
itemModule := module.GetModules().ItemModule
itemModule.AddItem(...)
```

### 4. 模块间调用

模块之间通过接口调用，避免直接依赖：

```go
// 装备模块调用道具模块
itemModule := module.GetModules().ItemModule
itemModule.DeductItem(...)
```

## 使用示例

### 1. 在 Actor 中使用模块

```go
// actor/player/actor_player.go
func (p *actorPlayer) buyItem(session *cproto.Session, req *pb.BuyItemRequest) {
    // 1. 参数验证
    if req.ItemId <= 0 || req.Count <= 0 {
        p.ResponseCode(session, code.ShopItemInvalidParam)
        return
    }
    
    // 2. 获取模块
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

### 2. 模块间调用

```go
// module/equipment/equipment_impl.go
func (m *EquipmentModule) EquipItem(playerId int64, position int32, itemId int32) error {
    // 1. 调用道具模块检查道具
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

## 新增模块步骤

### 1. 创建模块目录和文件

```bash
mkdir -p app/game/module/skill
touch app/game/module/skill/skill.go
touch app/game/module/skill/skill_impl.go
```

### 2. 定义接口

```go
// module/skill/skill.go
package skill

type ISkillModule interface {
    LearnSkill(playerId int64, skillId int32) error
    UpgradeSkill(playerId int64, skillId int32) error
}
```

### 3. 实现接口

```go
// module/skill/skill_impl.go
package skill

type SkillModule struct{}

func NewSkillModule() ISkillModule {
    return &SkillModule{}
}

func (m *SkillModule) LearnSkill(playerId int64, skillId int32) error {
    // 实现逻辑
}
```

### 4. 注册到管理器

```go
// module/manager.go
type ModuleManager struct {
    ItemModule      item.IItemModule
    EquipmentModule equipment.IEquipmentModule
    SkillModule     skill.ISkillModule  // 新增
}

func InitModules() {
    manager = &ModuleManager{
        ItemModule:      item.NewItemModule(),
        EquipmentModule: equipment.NewEquipmentModule(),
        SkillModule:     skill.NewSkillModule(),  // 新增
    }
}
```

### 5. 在 Actor 中使用

```go
// actor/player/actor_player.go
func (p *actorPlayer) learnSkill(session *cproto.Session, req *pb.LearnSkillRequest) {
    skillModule := module.GetModules().SkillModule
    err := skillModule.LearnSkill(p.playerId, req.SkillId)
    // ...
}
```

## 优势

1. **分层清晰**: Actor 层只负责消息处理，业务逻辑在 Module 层
2. **易扩展**: 新增模块只需实现接口，注册到管理器
3. **模块解耦**: 通过接口调用，模块间依赖清晰
4. **类型安全**: 接口定义明确，编译期检查
5. **易测试**: 模块可以独立测试，Mock 接口简单
6. **易维护**: 业务逻辑集中，修改影响范围小

## 注意事项

1. **模块初始化**: 在 `cmd/game/main.go` 中调用 `module.InitModules()`
2. **线程安全**: `ModuleManager` 是只读的，初始化后不再修改
3. **循环依赖**: 避免模块间循环调用，如必须，使用事件机制
4. **错误处理**: 模块方法返回 error，由调用方处理
5. **日志记录**: 在模块实现中记录关键操作日志

## 最佳实践

1. **接口设计**: 接口方法应该职责单一，参数明确
2. **错误处理**: 使用明确的错误类型，便于调用方处理
3. **缓存策略**: 在模块实现中使用缓存，减少数据库访问
4. **事件机制**: 关键操作发送事件，其他模块可以订阅
5. **批量操作**: 提供批量操作方法，减少调用次数



