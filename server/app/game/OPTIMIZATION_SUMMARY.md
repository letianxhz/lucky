# Game 服务模块层次结构优化总结

## 优化目标

优化 game 服务中多个 actor 的模块层次结构，使其清晰合理、易于扩展。

## 优化内容

### 1. 创建统一的 Actor 注册机制

**文件**：`app/game/actor/registry.go`

**优化前**：
- Actor 注册分散在 `main.go` 中
- 新增 Actor 需要修改 `main.go`

**优化后**：
- 所有 Actor 统一在 `registry.go` 中注册
- 新增 Actor 只需在 `RegisterActors()` 中添加
- `main.go` 只需调用 `actor.RegisterActors()`

**示例**：
```go
// actor/registry.go
func RegisterActors() []cfacade.IActorHandler {
    return []cfacade.IActorHandler{
        NewActorPlayers(),
        NewActorRooms(),
        // 新增 Actor 只需在这里添加
    }
}

// cmd/game/main.go
app.AddActors(actor.RegisterActors()...)
```

### 2. 创建 Actor 基类

**文件**：`app/game/actor/base/base.go`

**优化前**：
- 每个 Actor 都需要重复实现 `OnInit()` 中的 Handler 注册逻辑

**优化后**：
- 提供 `BaseActor` 基类，统一处理 Handler 注册
- 子类可以继承基类，减少重复代码

**示例**：
```go
// 使用基类（可选，如果 Actor 逻辑简单）
type ActorNewActor struct {
    base.BaseActor
}

func NewActorNewActor() *ActorNewActor {
    return &ActorNewActor{
        BaseActor: *base.NewBaseActor(handler.ActorTypeNewActor),
    }
}
```

### 3. 创建架构设计文档

**文件**：`app/game/ARCHITECTURE.md`

**内容**：
- 整体架构设计
- 层次结构说明
- 模块依赖关系
- 扩展指南
- 最佳实践

### 4. 创建 Module 层设计文档

**文件**：`app/game/module/README.md`

**内容**：
- Module 层设计原则
- 模块职责划分
- 添加新模块的步骤
- 依赖关系说明

## 优化后的目录结构

```
app/game/
├── ARCHITECTURE.md          # 整体架构设计文档
├── OPTIMIZATION_SUMMARY.md  # 优化总结（本文档）
│
├── actor/                   # Actor 层
│   ├── base/                # Actor 基类（新增）
│   │   └── base.go
│   ├── registry.go          # Actor 注册表（新增）
│   ├── player/
│   │   ├── actor_player.go
│   │   └── actor_players.go
│   ├── room/
│   │   ├── actor_room.go
│   │   └── actor_rooms.go
│   ├── alliance/
│   │   └── actor_alliance.go
│   ├── player.go            # 导出函数
│   └── room.go              # 导出函数
│
├── module/                  # Module 层
│   ├── README.md            # Module 层设计文档（新增）
│   ├── manager.go           # 模块管理器
│   ├── handler/             # Handler 基础设施
│   ├── login/               # 登录模块
│   ├── item/                # 道具模块
│   ├── room/                # 房间模块
│   ├── alliance/            # 联盟模块
│   ├── player/              # 玩家模块
│   ├── equipment/           # 装备模块
│   └── online/              # 在线状态模块
│
└── cmd/game/
    └── main.go              # 使用统一的注册机制
```

## 层次结构说明

### 1. Actor 层

**职责**：
- 消息路由和分发
- Actor 生命周期管理
- 子 Actor 创建和管理
- 跨 Actor 通信

**特点**：
- 统一的注册机制（`registry.go`）
- 可选的基类支持（`base/base.go`）
- 每个 Actor 类型有独立目录

### 2. Module 层

**职责**：
- 业务逻辑实现
- 数据访问和缓存
- 业务规则验证
- 跨模块协作

**特点**：
- 通过接口定义契约
- 通过 DI 容器管理依赖
- 自动注册机制（`init()` 函数）

### 3. Handler 层

**职责**：
- 消息处理器注册
- 消息路由分发
- 类型转换和验证
- 错误处理

**特点**：
- 支持多种 Actor 类型
- 类型安全的泛型注册机制
- 自动注册到对应的 Actor

## 扩展指南

### 添加新的 Actor

1. **创建 Actor 实现**：
   ```go
   // actor/newactor/actor_newactor.go
   type ActorNewActor struct {
       pomelo.ActorBase
   }
   
   func (a *ActorNewActor) OnInit() {
       handler.RegisterAllToActorByType(handler.ActorTypeNewActor, &a.ActorBase)
   }
   ```

2. **定义导出函数**（如果需要）：
   ```go
   // actor/newactor.go
   package actor
   
   func NewActorNewActor() *newactor.ActorNewActor {
       return &newactor.ActorNewActor{}
   }
   ```

3. **在 registry.go 中注册**：
   ```go
   func RegisterActors() []cfacade.IActorHandler {
       return []cfacade.IActorHandler{
           NewActorPlayers(),
           NewActorRooms(),
           NewActorNewActor(), // 新增
       }
   }
   ```

### 添加新的 Module

1. **创建模块目录和接口**
2. **实现接口并自动注册到 DI**
3. **注册 Handler（如果需要）**
4. **在 manager.go 中导入（触发 init）**

详细步骤请参考 `module/README.md`。

## 优势

### 1. 清晰的层次结构

- Actor 层、Module 层、Handler 层职责分明
- 每个层次都有明确的职责和边界

### 2. 易于扩展

- 新增 Actor 只需在 `registry.go` 中添加
- 新增 Module 只需遵循约定，无需修改核心代码
- 统一的注册机制，减少重复代码

### 3. 依赖管理

- 通过 DI 容器管理模块依赖
- 避免循环依赖
- 接口隔离，实现可替换

### 4. 统一注册

- Actor 统一在 `registry.go` 中注册
- Module 通过 `init()` 自动注册
- Handler 通过 `init()` 自动注册

### 5. 文档完善

- 架构设计文档（`ARCHITECTURE.md`）
- Module 层设计文档（`module/README.md`）
- 优化总结文档（本文档）

## 使用示例

### 启动服务

```go
// cmd/game/main.go
func main() {
    // ... 配置和组件注册 ...
    
    // 统一初始化 DI 容器
    di.MustInitialize()
    
    // 注册所有 Actor（统一管理）
    app.AddActors(actor.RegisterActors()...)
    
    // 启动服务器
    app.Startup()
}
```

### Actor 初始化

```go
// actor/player/actor_player.go
func (p *actorPlayer) OnInit() {
    // 自动注册所有为 ActorTypePlayer 注册的 Handler
    handler.RegisterAllToActorByType(handler.ActorTypePlayer, &p.ActorBase)
}
```

### Module 注册

```go
// module/item/item_impl.go
func init() {
    var v = &ItemModule{
        items: make(map[int64]map[int32]int64),
    }
    di.Register(v)
    di.RegisterImplementation((*IItemModule)(nil), v)
}
```

### Handler 注册

```go
// module/item/handler.go
func init() {
    var h = &itemHandler{}
    di.Register(h)
    handler.RegisterHandler(handler.ActorTypePlayer, "buyItem", h.OnBuyItem)
}
```

## 总结

通过这次优化，我们实现了：

✅ **清晰的层次结构**：Actor 层、Module 层、Handler 层职责分明  
✅ **统一的注册机制**：Actor 和 Module 都有统一的注册方式  
✅ **易于扩展**：新增 Actor 或 Module 只需遵循约定  
✅ **依赖管理**：通过 DI 容器管理依赖，避免循环依赖  
✅ **文档完善**：提供了详细的架构设计和扩展指南  

这些优化使得代码结构更加清晰合理，易于维护和扩展。



