# 消息处理器注册机制

本文档说明如何使用注解式消息处理器注册机制，参考 Java 的 `@MsgReceiver` 设计。

## 设计理念

参考 Java 代码：
```java
@MsgReceiver(MsgIds.CSMallQuery)
public static void onCSMallQuery(MsgParam param) {
    HumanService humanService = param.getOwnerObject();
    humanService.getHumanModule().getModMall().query();
}
```

在 Go 中实现类似的注解式注册：
```go
func init() {
    handler.RegisterMsg("buyItem", OnBuyItem)
}

func OnBuyItem(param *handler.MsgParam) {
    itemModule := param.GetModule().(*ItemModule)
    // 处理逻辑
}
```

## 两种注册方式

### 方式 1: 旧的 Register/RegisterAll 方式（兼容）

适用于需要复杂逻辑或需要访问 actor 的场景：

```go
// handler.go
func init() {
    handler.Register(func(actor *pomelo.ActorBase) {
        actor.Local().Register("buyItem", func(session *cproto.Session, req *pb.BuyItemRequest) {
            // 处理逻辑
        })
    })
}
```

### 方式 2: 新的 RegisterMsg/RegisterAllToActor 方式（推荐）

更简洁，类似 Java 的注解方式：

```go
// handler.go
func init() {
    handler.RegisterMsg("buyItem", OnBuyItem)
}

// OnBuyItem 购买道具消息处理器
// 类似 Java: public static void onCSMallBuy(MsgParam param)
func OnBuyItem(param *handler.MsgParam) {
    // 从 di 容器获取模块实例
    itemModuleInstance, err := di.GetByType((*IItemModule)(nil))
    if err != nil {
        param.GetActor().ResponseCode(param.GetSession(), code.Error)
        return
    }
    
    itemModule := itemModuleInstance.(IItemModule)
    
    // 获取消息对象
    req, ok := param.GetMsg().(*pb.BuyItemRequest)
    if !ok {
        param.GetActor().ResponseCode(param.GetSession(), code.ShopItemInvalidParam)
        return
    }
    
    // 处理业务逻辑
    // ...
    
    // 返回响应
    param.GetActor().Response(param.GetSession(), response)
}
```

## MsgParam API

`MsgParam` 提供了类似 Java `MsgParam` 的接口：

```go
type MsgParam struct {
    Session *cproto.Session
    Actor   *pomelo.ActorBase
    Msg     interface{} // 消息对象
}

// GetSession 获取 session
func (p *MsgParam) GetSession() *cproto.Session

// GetActor 获取 actor
func (p *MsgParam) GetActor() *pomelo.ActorBase

// GetMsg 获取消息对象
func (p *MsgParam) GetMsg() interface{}
```

## 完整示例

### Item 模块示例

```go
package item

import (
    "lucky/server/app/game/module/handler"
    "lucky/server/pkg/di"
    "lucky/server/pkg/pb"
)

func init() {
    handler.RegisterMsg("buyItem", OnBuyItem)
}

func OnBuyItem(param *handler.MsgParam) {
    // 获取模块实例
    itemModuleInstance, err := di.GetByType((*IItemModule)(nil))
    if err != nil {
        param.GetActor().ResponseCode(param.GetSession(), code.Error)
        return
    }
    
    itemModule := itemModuleInstance.(IItemModule)
    
    // 获取消息
    req, ok := param.GetMsg().(*pb.BuyItemRequest)
    if !ok {
        param.GetActor().ResponseCode(param.GetSession(), code.ShopItemInvalidParam)
        return
    }
    
    // 业务逻辑
    playerId := db.GetPlayerIdWithUID(param.GetSession().Uid)
    err = itemModule.AddItem(playerId, req.ItemId, int64(req.Count))
    if err != nil {
        param.GetActor().ResponseCode(param.GetSession(), code.ShopItemBuyFail)
        return
    }
    
    // 返回响应
    response := &pb.BuyItemResponse{
        ItemId: req.ItemId,
        Count:  req.Count,
    }
    param.GetActor().Response(param.GetSession(), response)
}
```

### Login 模块示例

```go
package login

import (
    "lucky/server/app/game/module/handler"
    "lucky/server/pkg/di"
    "lucky/server/pkg/pb"
)

func init() {
    handler.RegisterMsg("select", OnPlayerSelect)
    handler.RegisterMsg("create", OnPlayerCreate)
    handler.RegisterMsg("enter", OnPlayerEnter)
}

func OnPlayerSelect(param *handler.MsgParam) {
    loginModuleInstance, _ := di.GetByType((*ILoginModule)(nil))
    loginModule := loginModuleInstance.(ILoginModule)
    
    response, err := loginModule.SelectPlayer(param.GetSession())
    if err != nil {
        param.GetActor().ResponseCode(param.GetSession(), code.Error)
        return
    }
    
    param.GetActor().Response(param.GetSession(), response)
}

func OnPlayerCreate(param *handler.MsgParam) {
    loginModuleInstance, _ := di.GetByType((*ILoginModule)(nil))
    loginModule := loginModuleInstance.(ILoginModule)
    
    req := param.GetMsg().(*pb.PlayerCreateRequest)
    response, err := loginModule.CreatePlayer(param.GetSession(), req, param.GetActor())
    if err != nil {
        param.GetActor().ResponseCode(param.GetSession(), code.PlayerCreateFail)
        return
    }
    
    param.GetActor().Response(param.GetSession(), response)
}

func OnPlayerEnter(param *handler.MsgParam) {
    loginModuleInstance, _ := di.GetByType((*ILoginModule)(nil))
    loginModule := loginModuleInstance.(ILoginModule)
    
    req := param.GetMsg().(*pb.Int64)
    response, err := loginModule.EnterPlayer(param.GetSession(), req, param.GetActor())
    if err != nil {
        param.GetActor().ResponseCode(param.GetSession(), code.PlayerIDError)
        return
    }
    
    param.GetActor().Response(param.GetSession(), response)
}
```

## 优势

1. **更简洁**: 每个消息处理器都是独立的函数，代码更清晰
2. **类型安全**: 通过类型断言确保消息类型正确
3. **易于测试**: 每个处理器函数都可以独立测试
4. **类似 Java**: 与 Java 的 `@MsgReceiver` 注解方式类似，便于理解

## 迁移指南

从旧方式迁移到新方式：

1. 创建 `handler_new.go` 文件（或重命名 `handler.go`）
2. 将每个消息处理器提取为独立函数
3. 在 `init()` 中使用 `handler.RegisterMsg()` 注册
4. 测试验证后，删除旧的 `handler.go` 文件

## 注意事项

1. **消息类型**: 需要在处理器内部进行类型断言
2. **错误处理**: 使用 `param.GetActor().ResponseCode()` 返回错误码
3. **成功响应**: 使用 `param.GetActor().Response()` 返回成功响应
4. **模块获取**: 通过 `di.GetByType()` 获取模块实例

