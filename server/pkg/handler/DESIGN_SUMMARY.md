# 消息处理器架构设计总结

## 设计目标

设计一个优雅且易扩展的消息处理器注册机制，支持多种 Actor 类型（Player、Alliance、Room、Guild 等），参考 Java 的 `@MsgReceiver` 注解方式。

## 核心设计

### 1. Actor 类型系统

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

**优势**:
- 类型安全：通过常量定义，避免字符串拼写错误
- 易于扩展：新增 Actor 类型只需添加常量
- 类型隔离：不同 Actor 类型的消息处理器完全隔离

### 2. 消息处理器注册

#### 方式 1: 默认注册到 Player Actor

```go
// 适用于 Player Actor 的消息处理器
handler.RegisterMsg("buyItem", OnBuyItem)
```

#### 方式 2: 为指定 Actor 类型注册

```go
// 适用于 Alliance Actor 的消息处理器
handler.RegisterMsgForActor(handler.ActorTypeAlliance, "createAlliance", OnCreateAlliance)
```

### 3. Actor 初始化

```go
// Player Actor
func (p *actorPlayer) OnInit() {
    handler.RegisterAllToActorByType(handler.ActorTypePlayer, &p.ActorBase)
}

// Alliance Actor
func (a *ActorAlliance) OnInit() {
    handler.RegisterAllToActorByType(handler.ActorTypeAlliance, &a.ActorBase)
}
```

## 架构优势

### 1. 类型隔离

不同 Actor 类型的消息处理器完全隔离，不会冲突：

```go
// Player Actor 的 buyItem
handler.RegisterMsg("buyItem", OnPlayerBuyItem)

// Alliance Actor 的 buyItem（不会冲突）
handler.RegisterMsgForActor(handler.ActorTypeAlliance, "buyItem", OnAllianceBuyItem)
```

### 2. 易于扩展

新增 Actor 类型只需 3 步：

1. **定义 Actor 类型常量**
   ```go
   const ActorTypeNewType ActorType = "newtype"
   ```

2. **创建 Actor 实现**
   ```go
   func (a *ActorNewType) OnInit() {
       handler.RegisterAllToActorByType(handler.ActorTypeNewType, &a.ActorBase)
   }
   ```

3. **创建模块和处理器**
   ```go
   func init() {
       handler.RegisterMsgForActor(handler.ActorTypeNewType, "doSomething", OnDoSomething)
   }
   ```

### 3. 模块化设计

每个业务模块独立管理自己的消息处理器：

```
module/
├── item/
│   ├── handler.go          # Item 模块的消息处理器（Player Actor）
│   ├── item.go             # 接口定义
│   └── item_impl.go         # 实现
├── alliance/
│   ├── handler.go           # Alliance 模块的消息处理器（Alliance Actor）
│   ├── alliance.go          # 接口定义
│   └── alliance_impl.go     # 实现
└── room/
    ├── handler.go           # Room 模块的消息处理器（Room Actor）
    ├── room.go              # 接口定义
    └── room_impl.go         # 实现
```

### 4. 统一接口

所有消息处理器使用相同的 `MsgParam` 接口：

```go
type MsgParam struct {
    Session *cproto.Session
    Actor   *pomelo.ActorBase
    Msg     interface{}
}

// 提供统一的访问方法
func (p *MsgParam) GetSession() *cproto.Session
func (p *MsgParam) GetActor() *pomelo.ActorBase
func (p *MsgParam) GetMsg() interface{}
```

## 使用示例

### Player Actor 消息处理器

```go
// module/item/handler.go
func init() {
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

### Alliance Actor 消息处理器

```go
// module/alliance/handler.go
func init() {
    handler.RegisterMsgForActor(handler.ActorTypeAlliance, "createAlliance", OnCreateAlliance)
}

func OnCreateAlliance(param *handler.MsgParam) {
    allianceModuleInstance, _ := di.GetByType((*IAllianceModule)(nil))
    allianceModule := allianceModuleInstance.(IAllianceModule)
    
    req := param.GetMsg().(*pb.CreateAllianceRequest)
    // 处理逻辑...
    
    param.GetActor().Response(param.GetSession(), response)
}
```

### Room Actor 消息处理器

```go
// module/room/handler.go
func init() {
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

## 文件结构

```
app/game/
├── actor/
│   ├── player/
│   │   └── actor_player.go      # Player Actor 实现
│   ├── alliance/
│   │   └── actor_alliance.go     # Alliance Actor 实现
│   └── room/
│       └── actor_room.go         # Room Actor 实现
├── module/
│   ├── handler/
│   │   ├── actor_type.go         # Actor 类型定义
│   │   ├── receiver.go            # 消息处理器注册机制
│   │   ├── handler.go             # 旧的注册方式（兼容）
│   │   ├── ARCHITECTURE.md        # 架构设计文档
│   │   └── USAGE_EXAMPLES.md     # 使用示例
│   ├── item/
│   │   └── handler.go            # Item 模块消息处理器
│   ├── alliance/
│   │   └── handler.go            # Alliance 模块消息处理器
│   └── room/
│       └── handler.go             # Room 模块消息处理器
```

## 最佳实践

1. **模块组织**: 每个业务模块在各自的 `handler.go` 中注册处理器
2. **命名规范**: 处理器函数使用 `On` 前缀，如 `OnBuyItem`、`OnCreateAlliance`
3. **错误处理**: 统一使用 `param.GetActor().ResponseCode()` 返回错误码
4. **类型断言**: 始终检查类型断言结果，避免 panic
5. **日志记录**: 在关键步骤记录日志，便于调试

## 扩展指南

### 添加新的 Actor 类型

1. 在 `handler/actor_type.go` 中添加常量
2. 创建 `actor/newtype/actor_newtype.go`
3. 在 `OnInit()` 中调用 `handler.RegisterAllToActorByType()`
4. 创建对应的模块和 `handler.go`

### 添加新的消息处理器

1. 在模块的 `handler.go` 中定义处理器函数
2. 在 `init()` 中使用 `handler.RegisterMsg()` 或 `handler.RegisterMsgForActor()` 注册
3. 实现业务逻辑

## 总结

这个架构设计实现了：

✅ **类型隔离**: 不同 Actor 类型的消息处理器完全隔离  
✅ **易于扩展**: 新增 Actor 类型只需 3 步  
✅ **模块化**: 每个模块独立管理自己的消息处理器  
✅ **统一接口**: 所有消息处理器使用相同的 `MsgParam` 接口  
✅ **自动注册**: 通过 `init()` 自动注册，无需手动调用  
✅ **类型安全**: 通过类型断言确保消息类型正确  
✅ **类似 Java**: 与 `@MsgReceiver` 注解方式类似，便于理解

