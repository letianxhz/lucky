# Module 层设计说明

## 概述

Module 层负责实现业务逻辑，是游戏服务的核心业务层。每个 Module 代表一个业务领域，通过接口定义契约，通过 DI 容器管理依赖。

## 目录结构

```
module/
├── handler/           # 消息处理器基础设施（不依赖业务模块）
│   ├── actor_type.go  # Actor 类型定义
│   ├── receiver_v3.go # Handler 注册机制（V3 泛型版本）
│   └── ...
├── login/             # 登录模块
│   ├── login.go       # 接口定义 (ILoginModule)
│   ├── login_impl.go  # 实现（自动注册到 DI）
│   └── handler.go     # 消息处理器（自动注册）
├── item/              # 道具模块
│   ├── item.go        # 接口定义 (IItemModule)
│   ├── item_impl.go   # 实现（自动注册到 DI）
│   └── handler.go     # 消息处理器（自动注册）
├── room/              # 房间模块
│   ├── room.go        # 接口定义 (IRoomModule)
│   ├── room_impl.go   # 实现（自动注册到 DI）
│   └── handler.go     # 消息处理器（自动注册）
├── alliance/          # 联盟模块
│   ├── alliance.go    # 接口定义 (IAllianceModule)
│   └── handler.go     # 消息处理器（自动注册）
├── player/            # 玩家模块（跨模块协调）
│   └── handler.go     # 消息处理器（自动注册）
├── equipment/         # 装备模块
│   ├── equipment.go   # 接口定义 (IEquipmentModule)
│   └── equipment_impl.go # 实现（自动注册到 DI）
├── online/            # 在线状态模块
│   └── online.go      # 实现（自动注册到 DI）
└── manager.go         # 模块管理器（触发所有模块的 init）
```

## 模块设计原则

### 1. 接口定义

每个模块都应该定义接口，明确模块的职责和契约：

```go
// module/item/item.go
type IItemModule interface {
    AddItem(playerId int64, itemId int32, count int64) error
    DeductItem(playerId int64, itemId int32, count int64) error
    GetItems(playerId int64) (map[int32]int64, error)
}
```

### 2. 实现自动注册

实现类通过 `init()` 函数自动注册到 DI 容器：

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

### 3. Handler 自动注册

Handler 通过 `init()` 函数自动注册到 Handler 注册表：

```go
// module/item/handler.go
func init() {
    var h = &itemHandler{}
    di.Register(h)
    handler.RegisterHandler(handler.ActorTypePlayer, "buyItem", h.OnBuyItem)
}
```

### 4. 依赖注入

模块间的依赖通过 DI 容器注入，使用 `di:"auto"` 标签：

```go
// module/equipment/equipment_impl.go
type EquipmentModule struct {
    item IItemModule `di:"auto"` // 自动注入 ItemModule
}
```

## 模块依赖关系

### 依赖规则

1. **Module → Module**：允许，通过 DI 注入
2. **Handler → Module**：允许，通过 DI 注入
3. **禁止循环依赖**：如果 A 依赖 B，B 不能依赖 A

### 当前依赖关系

```
LoginModule (独立)
    ↓
ItemModule (独立)
    ↓
EquipmentModule → ItemModule
    ↓
PlayerHandler → LoginModule, ItemModule, RoomModule
    ↓
RoomModule (独立)
    ↓
AllianceModule (独立)
```

## 添加新模块

### 步骤 1: 创建目录和接口

```go
// module/newmodule/newmodule.go
package newmodule

type INewModule interface {
    DoSomething(session *cproto.Session, req *pb.Request) (*pb.Response, error)
}
```

### 步骤 2: 实现接口

```go
// module/newmodule/newmodule_impl.go
package newmodule

import (
    "lucky/server/pkg/di"
    // 其他依赖
)

type NewModule struct {
    // 依赖其他模块（可选）
    item IItemModule `di:"auto"`
}

func init() {
    var v = &NewModule{}
    di.Register(v)
    di.RegisterImplementation((*INewModule)(nil), v)
}

func (m *NewModule) DoSomething(session *cproto.Session, req *pb.Request) (*pb.Response, error) {
    // 实现逻辑
    return &pb.Response{}, nil
}
```

### 步骤 3: 注册 Handler（如果需要）

```go
// module/newmodule/handler.go
package newmodule

import (
    "lucky/server/app/game/module/handler"
    "lucky/server/pkg/di"
    "lucky/server/pkg/pb"
    
    cproto "github.com/cherry-game/cherry/net/proto"
)

func init() {
    var h = &newModuleHandler{}
    di.Register(h)
    handler.RegisterHandler(handler.ActorTypePlayer, "newAction", h.OnNewAction)
}

type newModuleHandler struct {
    newModule INewModule `di:"auto"`
}

func (h *newModuleHandler) OnNewAction(session *cproto.Session, req *pb.Request) (*pb.Response, error) {
    return h.newModule.DoSomething(session, req)
}
```

### 步骤 4: 在 manager.go 中导入（触发 init）

```go
// module/manager.go
package module

import (
    _ "lucky/server/app/game/module/newmodule" // 触发 init 函数
    // 其他模块...
)
```

## 模块职责划分

### LoginModule
- 玩家登录、创建、选择角色
- 玩家进入游戏

### ItemModule
- 道具的增删改查
- 道具数量管理

### EquipmentModule
- 装备的穿戴、卸下
- 装备属性计算
- 依赖 ItemModule

### RoomModule
- 房间的创建、加入、离开
- 房间信息管理

### AllianceModule
- 联盟的创建、加入、离开
- 联盟信息管理

### OnlineModule
- 玩家在线状态管理
- 玩家绑定/解绑

### PlayerHandler
- 跨模块协调
- 玩家相关的复杂业务逻辑

## 最佳实践

1. **单一职责**：每个模块只负责一个业务领域
2. **接口隔离**：通过接口定义模块契约
3. **依赖注入**：所有依赖通过 DI 容器注入
4. **自动注册**：通过 `init()` 函数自动注册
5. **避免循环依赖**：如果出现循环依赖，考虑提取公共模块
6. **Handler 轻量级**：Handler 只负责消息接收和参数验证，业务逻辑委托给 Module

## 测试

每个模块都应该有对应的测试文件：

```go
// module/item/item_test.go
func TestItemModule_AddItem(t *testing.T) {
    di.Reset()
    var itemModule = &ItemModule{
        items: make(map[int64]map[int32]int64),
    }
    di.Register(itemModule)
    di.RegisterImplementation((*IItemModule)(nil), itemModule)
    di.MustInitialize()
    
    // 测试逻辑...
}
```

