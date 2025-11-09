# 消息处理器架构设计

本文档说明如何设计优雅且易扩展的消息处理器注册机制，支持多种 Actor 类型。

## 设计目标

1. **支持多种 Actor 类型**: Player、Alliance、Room、Guild 等
2. **易于扩展**: 新增 Actor 类型只需定义常量，无需修改核心代码
3. **模块化**: 每个模块独立注册自己的消息处理器
4. **类型安全**: 通过类型断言确保消息类型正确
5. **自动注册**: 通过 `init()` 函数自动注册，无需手动调用

## 架构设计

### 1. Actor 类型定义

```go
// handler/actor_type.go
type ActorType string

const (
    ActorTypePlayer   ActorType = "player"
    ActorTypeAlliance ActorType = "alliance"
    ActorTypeRoom     ActorType = "room"
    ActorTypeGuild    ActorType = "guild"
    ActorTypeWorld    ActorType = "world"
)
```

### 2. 消息处理器注册

#### 方式 1: 为 Player Actor 注册（默认）

```go
// item/handler.go
func init() {
    // 默认注册到 Player Actor
    handler.RegisterMsg("buyItem", OnBuyItem)
}
```

#### 方式 2: 为指定 Actor 类型注册

```go
// alliance/handler.go
func init() {
    // 为 Alliance Actor 注册
    handler.RegisterMsgForActor(handler.ActorTypeAlliance, "createAlliance", OnCreateAlliance)
    handler.RegisterMsgForActor(handler.ActorTypeAlliance, "joinAlliance", OnJoinAlliance)
}
```

### 3. Actor 初始化

```go
// actor/player/actor_player.go
func (p *actorPlayer) OnInit() {
    // 注册所有 Player Actor 的消息处理器
    handler.RegisterAllToActorByType(handler.ActorTypePlayer, &p.ActorBase)
}

// actor/alliance/actor_alliance.go
func (a *ActorAlliance) OnInit() {
    // 注册所有 Alliance Actor 的消息处理器
    handler.RegisterAllToActorByType(handler.ActorTypeAlliance, &a.ActorBase)
}
```

## 使用示例

### Player Actor 示例

```go
// module/item/handler.go
package item

import (
    "lucky/server/app/game/module/handler"
    "lucky/server/pkg/di"
    "lucky/server/pkg/pb"
)

func init() {
    // 为 Player Actor 注册消息处理器
    handler.RegisterMsg("buyItem", OnBuyItem)
}

func OnBuyItem(param *handler.MsgParam) {
    itemModuleInstance, _ := di.GetByType((*IItemModule)(nil))
    itemModule := itemModuleInstance.(IItemModule)
    
    req := param.GetMsg().(*pb.BuyItemRequest)
    // 处理逻辑...
    
    param.GetActor().Response(param.GetSession(), response)
}
```

### Alliance Actor 示例

```go
// module/alliance/handler.go
package alliance

import (
    "lucky/server/app/game/module/handler"
    "lucky/server/pkg/di"
    "lucky/server/pkg/pb"
)

func init() {
    // 为 Alliance Actor 注册消息处理器
    handler.RegisterMsgForActor(handler.ActorTypeAlliance, "createAlliance", OnCreateAlliance)
    handler.RegisterMsgForActor(handler.ActorTypeAlliance, "joinAlliance", OnJoinAlliance)
}

func OnCreateAlliance(param *handler.MsgParam) {
    allianceModuleInstance, _ := di.GetByType((*IAllianceModule)(nil))
    allianceModule := allianceModuleInstance.(IAllianceModule)
    
    req := param.GetMsg().(*pb.CreateAllianceRequest)
    // 处理逻辑...
    
    param.GetActor().Response(param.GetSession(), response)
}
```

### Room Actor 示例

```go
// module/room/handler.go
package room

import (
    "lucky/server/app/game/module/handler"
    "lucky/server/pkg/di"
    "lucky/server/pkg/pb"
)

func init() {
    // 为 Room Actor 注册消息处理器
    handler.RegisterMsgForActor(handler.ActorTypeRoom, "createRoom", OnCreateRoom)
    handler.RegisterMsgForActor(handler.ActorTypeRoom, "joinRoom", OnJoinRoom)
}

func OnCreateRoom(param *handler.MsgParam) {
    roomModuleInstance, _ := di.GetByType((*IRoomModule)(nil))
    roomModule := roomModuleInstance.(IRoomModule)
    
    req := param.GetMsg().(*pb.CreateRoomRequest)
    // 处理逻辑...
    
    param.GetActor().Response(param.GetSession(), response)
}
```

## 扩展新 Actor 类型

### 步骤 1: 定义 Actor 类型

```go
// handler/actor_type.go
const (
    ActorTypeNewType ActorType = "newtype"
)
```

### 步骤 2: 创建 Actor 实现

```go
// actor/newtype/actor_newtype.go
package newtype

import (
    "lucky/server/app/game/module/handler"
    "github.com/cherry-game/cherry/net/parser/pomelo"
)

type ActorNewType struct {
    pomelo.ActorBase
}

func (a *ActorNewType) OnInit() {
    handler.RegisterAllToActorByType(handler.ActorTypeNewType, &a.ActorBase)
}
```

### 步骤 3: 创建模块和处理器

```go
// module/newtype/handler.go
package newtype

func init() {
    handler.RegisterMsgForActor(handler.ActorTypeNewType, "doSomething", OnDoSomething)
}

func OnDoSomething(param *handler.MsgParam) {
    // 处理逻辑...
}
```

## 优势

1. **类型隔离**: 不同 Actor 类型的消息处理器完全隔离，不会冲突
2. **易于扩展**: 新增 Actor 类型只需 3 步
3. **模块化**: 每个模块独立管理自己的消息处理器
4. **自动注册**: 通过 `init()` 自动注册，无需手动调用
5. **类型安全**: 通过类型断言确保消息类型正确
6. **统一接口**: 所有消息处理器使用相同的 `MsgParam` 接口

## 最佳实践

1. **模块组织**: 每个业务模块（item、alliance、room）在各自的 `handler.go` 中注册处理器
2. **命名规范**: 处理器函数使用 `On` 前缀，如 `OnBuyItem`、`OnCreateAlliance`
3. **错误处理**: 统一使用 `param.GetActor().ResponseCode()` 返回错误码
4. **日志记录**: 在关键步骤记录日志，便于调试
5. **类型断言**: 始终检查类型断言结果，避免 panic

## 测试

```go
// 测试获取处理器
handler, ok := handler.GetHandlerForActor(handler.ActorTypeAlliance, "createAlliance")
if !ok {
    t.Error("handler not found")
}

// 测试获取所有处理器
handlers := handler.GetHandlersByActorType(handler.ActorTypeAlliance)
if len(handlers) == 0 {
    t.Error("no handlers found")
}
```

