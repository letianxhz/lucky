# 消息处理器使用示例

本文档提供不同 Actor 类型的消息处理器使用示例。

## Player Actor 示例

### Item 模块

```go
// module/item/handler.go
package item

import (
    "lucky/server/app/game/module/handler"
    "lucky/server/pkg/di"
    "lucky/server/pkg/pb"
)

func init() {
    // 默认注册到 Player Actor
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

### Login 模块

```go
// module/login/handler.go
package login

func init() {
    // 默认注册到 Player Actor
    handler.RegisterMsg("select", OnPlayerSelect)
    handler.RegisterMsg("create", OnPlayerCreate)
    handler.RegisterMsg("enter", OnPlayerEnter)
}
```

## Alliance Actor 示例

### Alliance 模块

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
    handler.RegisterMsgForActor(handler.ActorTypeAlliance, "leaveAlliance", OnLeaveAlliance)
    handler.RegisterMsgForActor(handler.ActorTypeAlliance, "getAllianceInfo", OnGetAllianceInfo)
}

func OnCreateAlliance(param *handler.MsgParam) {
    allianceModuleInstance, _ := di.GetByType((*IAllianceModule)(nil))
    allianceModule := allianceModuleInstance.(IAllianceModule)
    
    req := param.GetMsg().(*pb.CreateAllianceRequest)
    // 处理逻辑...
    
    param.GetActor().Response(param.GetSession(), response)
}
```

### Alliance Actor 实现

```go
// actor/alliance/actor_alliance.go
package alliance

import (
    "lucky/server/app/game/module/handler"
    "github.com/cherry-game/cherry/net/parser/pomelo"
)

type ActorAlliance struct {
    pomelo.ActorBase
    allianceId int64
}

func (a *ActorAlliance) OnInit() {
    // 注册所有 Alliance Actor 的消息处理器
    handler.RegisterAllToActorByType(handler.ActorTypeAlliance, &a.ActorBase)
}
```

## Room Actor 示例

### Room 模块

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
    handler.RegisterMsgForActor(handler.ActorTypeRoom, "leaveRoom", OnLeaveRoom)
    handler.RegisterMsgForActor(handler.ActorTypeRoom, "getRoomInfo", OnGetRoomInfo)
    handler.RegisterMsgForActor(handler.ActorTypeRoom, "broadcast", OnBroadcast)
}

func OnCreateRoom(param *handler.MsgParam) {
    roomModuleInstance, _ := di.GetByType((*IRoomModule)(nil))
    roomModule := roomModuleInstance.(IRoomModule)
    
    req := param.GetMsg().(*pb.CreateRoomRequest)
    // 处理逻辑...
    
    param.GetActor().Response(param.GetSession(), response)
}
```

### Room Actor 实现

```go
// actor/room/actor_room.go
package room

import (
    "lucky/server/app/game/module/handler"
    "github.com/cherry-game/cherry/net/parser/pomelo"
)

type ActorRoom struct {
    pomelo.ActorBase
    roomId string
}

func (r *ActorRoom) OnInit() {
    // 注册所有 Room Actor 的消息处理器
    handler.RegisterAllToActorByType(handler.ActorTypeRoom, &r.ActorBase)
}
```

## 扩展新 Actor 类型

### 步骤 1: 定义 Actor 类型

```go
// handler/actor_type.go
const (
    ActorTypeGuild ActorType = "guild"  // 新增公会类型
)
```

### 步骤 2: 创建 Actor 实现

```go
// actor/guild/actor_guild.go
package guild

import (
    "lucky/server/app/game/module/handler"
    "github.com/cherry-game/cherry/net/parser/pomelo"
)

type ActorGuild struct {
    pomelo.ActorBase
    guildId int64
}

func (g *ActorGuild) OnInit() {
    handler.RegisterAllToActorByType(handler.ActorTypeGuild, &g.ActorBase)
}
```

### 步骤 3: 创建模块和处理器

```go
// module/guild/handler.go
package guild

func init() {
    handler.RegisterMsgForActor(handler.ActorTypeGuild, "createGuild", OnCreateGuild)
    handler.RegisterMsgForActor(handler.ActorTypeGuild, "joinGuild", OnJoinGuild)
}

func OnCreateGuild(param *handler.MsgParam) {
    // 处理逻辑...
}
```

## 混合使用示例

一个模块可以为多个 Actor 类型注册处理器：

```go
// module/social/handler.go
package social

func init() {
    // 为 Player Actor 注册
    handler.RegisterMsg("sendMessage", OnSendMessage)
    
    // 为 Alliance Actor 注册
    handler.RegisterMsgForActor(handler.ActorTypeAlliance, "allianceChat", OnAllianceChat)
    
    // 为 Room Actor 注册
    handler.RegisterMsgForActor(handler.ActorTypeRoom, "roomChat", OnRoomChat)
}
```

## 测试示例

```go
// handler_test.go
func TestGetHandlerForActor(t *testing.T) {
    // 测试获取 Player Actor 的处理器
    handler, ok := handler.GetHandlerForActor(handler.ActorTypePlayer, "buyItem")
    if !ok {
        t.Error("handler not found")
    }
    
    // 测试获取 Alliance Actor 的处理器
    handler, ok = handler.GetHandlerForActor(handler.ActorTypeAlliance, "createAlliance")
    if !ok {
        t.Error("handler not found")
    }
}

func TestGetHandlersByActorType(t *testing.T) {
    // 获取所有 Alliance Actor 的处理器
    handlers := handler.GetHandlersByActorType(handler.ActorTypeAlliance)
    if len(handlers) == 0 {
        t.Error("no handlers found")
    }
    
    // 验证处理器
    _, ok := handlers["createAlliance"]
    if !ok {
        t.Error("createAlliance handler not found")
    }
}
```

