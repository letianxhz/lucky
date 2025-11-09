# Game 服务模块层次结构设计

## 设计目标

1. **清晰的层次结构**：Actor 层、Module 层、Handler 层职责分明
2. **易于扩展**：新增 Actor 或 Module 只需遵循约定，无需修改核心代码
3. **依赖管理**：通过 DI 容器管理模块依赖，避免循环依赖
4. **统一注册**：Actor 和 Handler 的注册机制统一且自动化

## 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Game Service                            │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Actor 层   │  │   Module 层   │  │  Handler 层   │     │
│  │              │  │              │  │              │     │
│  │ - Player     │  │ - Player/    │  │ - 消息注册    │     │
│  │ - Room       │  │   Login      │  │ - 路由分发    │     │
│  │ - Alliance   │  │ - Player/    │  │ - 类型转换    │     │
│  │ - ...        │  │   Item       │  │              │     │
│  │              │  │ - Room/      │  │              │     │
│  │              │  │   Room       │  │              │     │
│  │              │  │ - Shared/    │  │              │     │
│  │              │  │   Handler    │  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              DI 容器 (依赖注入)                      │   │
│  │  - 模块注册                                          │   │
│  │  - 依赖解析                                          │   │
│  │  - 生命周期管理                                      │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## 层次结构详解

### 1. Actor 层 (`app/game/actor/`)

**职责**：
- 消息路由和分发
- Actor 生命周期管理
- 子 Actor 创建和管理
- 跨 Actor 通信

**结构**：
```
actor/
├── base/              # Actor 基类和通用功能
│   ├── base.go        # Actor 基类，提供统一初始化
│   └── registry.go    # Actor 注册机制
├── player/            # 玩家 Actor
│   ├── actor_player.go      # 玩家子 Actor
│   └── actor_players.go     # 玩家管理 Actor
├── room/              # 房间 Actor
│   ├── actor_room.go        # 房间子 Actor
│   └── actor_rooms.go       # 房间管理 Actor
├── alliance/          # 联盟 Actor
│   └── actor_alliance.go
└── registry.go        # Actor 注册表（统一注册所有 Actor）
```

**设计原则**：
1. 每个 Actor 类型有独立的目录
2. Manager Actor（如 `ActorPlayers`）负责创建和管理子 Actor
3. 子 Actor（如 `actorPlayer`）处理具体的业务消息
4. 所有 Actor 通过 `OnInit()` 自动注册 Handler

### 2. Module 层 (`app/game/module/`)

**职责**：
- 业务逻辑实现
- 数据访问和缓存
- 业务规则验证
- 跨模块协作

**结构**（按 Actor 划分）：
```
module/
├── player/            # 玩家相关的所有模块
│   ├── login/         # 登录模块
│   │   ├── login.go   # 接口定义
│   │   ├── login_impl.go # 实现（自动注册到 DI）
│   │   └── handler.go # 消息处理器（自动注册）
│   ├── item/          # 道具模块
│   │   ├── item.go
│   │   ├── item_impl.go
│   │   └── handler.go
│   ├── equipment/     # 装备模块
│   │   ├── equipment.go
│   │   └── equipment_impl.go
│   └── handler.go     # 玩家相关的 handler（跨模块协调）
├── room/              # 房间相关的所有模块
│   └── room/          # 房间模块
│       ├── room.go
│       ├── room_impl.go
│       └── handler.go
├── alliance/          # 联盟相关的所有模块
│   └── alliance/      # 联盟模块
│       ├── alliance.go
│       └── handler.go
├── shared/            # 共享模块（跨 actor）
│   ├── online/        # 在线状态模块
│   │   └── online.go
│   └── handler/       # Handler 基础设施
│       ├── actor_type.go  # Actor 类型定义
│       ├── receiver_v3.go # Handler 注册机制（V3 泛型版本）
│       └── ...
└── manager.go         # 模块管理器（触发 init）
```

**设计原则**：
1. 每个模块有独立的目录
2. 模块通过接口定义契约（如 `ILoginModule`）
3. 实现类通过 `init()` 自动注册到 DI 容器
4. Handler 通过 `init()` 自动注册到 Handler 注册表
5. 模块间依赖通过 DI 容器注入

### 3. Handler 层 (`app/game/module/handler/`)

**职责**：
- 消息处理器注册
- 消息路由分发
- 类型转换和验证
- 错误处理

**设计原则**：
1. 支持多种 Actor 类型
2. 类型安全的泛型注册机制
3. 自动注册到对应的 Actor
4. 统一的错误处理机制

## 模块依赖关系

### 依赖规则

1. **Actor → Module**：Actor 可以依赖 Module（通过 DI）
2. **Module → Module**：Module 可以依赖其他 Module（通过 DI，避免循环依赖）
3. **Handler → Module**：Handler 可以依赖 Module（通过 DI）
4. **禁止**：Module 不能直接依赖 Actor（通过接口解耦）

### 依赖示例

```
Player Actor
    ↓ (通过 DI)
LoginModule ← ItemModule ← EquipmentModule
    ↓
Handler (LoginHandler)
```

## 扩展指南

### 添加新的 Actor

1. **定义 Actor 类型**：
   ```go
   // module/handler/actor_type.go
   const ActorTypeNewActor ActorType = "newActor"
   ```

2. **创建 Actor 实现**：
   ```go
   // actor/newactor/actor_newactor.go
   type ActorNewActor struct {
       pomelo.ActorBase
   }
   
   func (a *ActorNewActor) OnInit() {
       handler.RegisterAllToActorByType(handler.ActorTypeNewActor, &a.ActorBase)
   }
   ```

3. **注册 Actor**：
   ```go
   // actor/registry.go
   func RegisterActors() []cfacade.IActorHandler {
       return []cfacade.IActorHandler{
           actor.NewActorPlayers(),
           actor.NewActorRooms(),
           actor.NewActorNewActor(), // 新增
       }
   }
   ```

### 添加新的 Module

1. **定义接口**：
   ```go
   // module/newmodule/newmodule.go
   type INewModule interface {
       DoSomething(session *cproto.Session, req *pb.Request) (*pb.Response, error)
   }
   ```

2. **实现接口**：
   ```go
   // module/newmodule/newmodule_impl.go
   type NewModule struct {
       // 依赖其他模块
       item IItemModule `di:"auto"`
   }
   
   func init() {
       var v = &NewModule{}
       di.Register(v)
       di.RegisterImplementation((*INewModule)(nil), v)
   }
   ```

3. **注册 Handler**（如果需要）：
   ```go
   // module/newmodule/handler.go
   func init() {
       var h = &newModuleHandler{}
       di.Register(h)
       handler.RegisterHandler(handler.ActorTypePlayer, "newAction", h.OnNewAction)
   }
   ```

## 初始化流程

```
1. 应用启动
   ↓
2. 导入模块包（触发 init()）
   ├─ Module 实现注册到 DI
   └─ Handler 注册到 Handler 注册表
   ↓
3. DI.MustInitialize()
   └─ 注入所有模块依赖
   ↓
4. 注册 Actor
   └─ 通过 actor/registry.go 统一注册
   ↓
5. Actor.OnInit()
   └─ 自动注册对应的 Handler
   ↓
6. 服务就绪，开始处理消息
```

## 最佳实践

1. **Actor 职责单一**：每个 Actor 只处理一种类型的实体（Player、Room、Alliance）
2. **Module 业务独立**：每个 Module 负责一个业务领域（Login、Item、Room）
3. **Handler 轻量级**：Handler 只负责消息接收和参数验证，业务逻辑委托给 Module
4. **依赖注入**：所有依赖通过 DI 容器注入，避免直接 import
5. **接口隔离**：通过接口定义模块契约，实现可以替换
6. **自动注册**：通过 `init()` 函数自动注册，减少手动配置

## 目录结构规范

```
app/game/
├── actor/              # Actor 层
│   ├── base/           # 基类和通用功能
│   ├── player/         # 玩家 Actor
│   ├── room/           # 房间 Actor
│   ├── alliance/       # 联盟 Actor
│   └── registry.go     # Actor 注册表
├── module/             # Module 层（按 Actor 划分）
│   ├── player/         # 玩家相关的所有模块
│   │   ├── login/      # 登录模块
│   │   ├── item/       # 道具模块
│   │   ├── equipment/  # 装备模块
│   │   └── handler.go  # 玩家相关的 handler
│   ├── room/           # 房间相关的所有模块
│   │   └── room/       # 房间模块
│   ├── alliance/       # 联盟相关的所有模块
│   │   └── alliance/   # 联盟模块
│   ├── shared/         # 共享模块（跨 actor）
│   │   ├── online/     # 在线状态模块
│   │   └── handler/    # Handler 基础设施
│   └── manager.go      # 模块管理器
├── entity/             # Entity 层（按 Actor 划分）
│   ├── player/         # 玩家相关的实体
│   ├── room/           # 房间相关的实体
│   ├── alliance/       # 联盟相关的实体
│   └── shared/         # 共享实体
├── db/                 # 数据访问层
├── config/             # 配置
└── service/            # 服务层（路由、中间件等）
```

## 总结

这个架构设计提供了：
- ✅ 清晰的层次结构
- ✅ 易于扩展的机制
- ✅ 统一的注册方式
- ✅ 依赖注入管理
- ✅ 类型安全的 Handler

通过遵循这些设计原则和规范，可以轻松地添加新的 Actor 和 Module，同时保持代码的清晰和可维护性。

