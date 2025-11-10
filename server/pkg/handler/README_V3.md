# 消息处理器注册机制 V3（泛型版本，无反射，高性能）

## 概述

V3 版本使用 Go 泛型实现，完全避免反射，提供最佳性能：

1. **零反射开销**：使用泛型和闭包，编译时确定类型，运行时直接调用
2. **类型安全**：编译时类型检查，避免运行时类型错误
3. **代码简洁**：处理器方法直接接收具体类型，无需类型断言

## 使用方式

### 1. 定义处理器结构体

```go
type roomHandler struct {
    room IRoomModule `di:"auto"`  // 使用 di:"auto" 自动注入依赖
}
```

### 2. 实现处理器方法

处理器方法签名支持以下三种形式：

```go
// 形式 1: 返回 error（最常用）
func (r *roomHandler) OnLeaveRoom(session *cproto.Session, req *pb.LeaveRoomRequest) error {
    return r.room.LeaveRoom(session, req)
}

// 形式 2: 返回 (response, error)（推荐，用于需要返回数据的场景）
func (r *roomHandler) OnCreateRoom(session *cproto.Session, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
    return r.room.CreateRoom(session, req)
}

// 形式 3: 无返回值（不推荐，无法返回错误）
func (r *roomHandler) OnBroadcast(session *cproto.Session, req *pb.RoomBroadcastRequest) {
    r.room.Broadcast(session, req)
}
```

### 3. 注册处理器

在 `init()` 函数中注册：

```go
func init() {
    var h = &roomHandler{}
    di.Register(h)  // 注册到 DI 容器
    
    // 注册消息处理器（泛型版本，无反射）
    handler.RegisterHandlerWithResponse(handler.ActorTypeRoom, "createRoom", h.OnCreateRoom)
    handler.RegisterHandler(handler.ActorTypeRoom, "leaveRoom", h.OnLeaveRoom)
    handler.RegisterHandlerNoReturn(handler.ActorTypeRoom, "broadcast", h.OnBroadcast)
}
```

## 性能优势

### 1. 零反射开销

**旧版本（使用 MsgParam）**：
- 注册时：手动注册，需要手动类型断言
- 调用时：函数调用，需要手动类型转换

**V3 版本（使用泛型）**：
- 注册时：泛型在编译时确定类型，生成具体函数
- 调用时：直接函数调用，无反射开销

### 2. 类型转换优化

- 大多数情况下，cherry 框架已经将 `[]byte` 转换为具体类型
- 类型断言 `msg.(*T)` 是 O(1) 操作，性能极高
- 只有在类型不匹配时才进行序列化/反序列化（很少发生）

### 3. 内存分配优化

- 使用闭包保存类型信息，避免运行时类型查询
- 直接函数调用，无反射相关的内存分配

## 完整示例

```go
package room

import (
    "lucky/server/app/game/module/handler"
    "lucky/server/pkg/di"
    "lucky/server/pkg/pb"
    
    cproto "github.com/cherry-game/cherry/net/proto"
)

func init() {
    var h = &roomHandler{}
    di.Register(h)
    
    // 注册处理器（泛型版本，无反射）
    handler.RegisterHandlerWithResponse(handler.ActorTypeRoom, "createRoom", h.OnCreateRoom)
    handler.RegisterHandlerWithResponse(handler.ActorTypeRoom, "joinRoom", h.OnJoinRoom)
    handler.RegisterHandler(handler.ActorTypeRoom, "leaveRoom", h.OnLeaveRoom)
    handler.RegisterHandlerWithResponse(handler.ActorTypeRoom, "getRoomInfo", h.OnGetRoomInfo)
    handler.RegisterHandler(handler.ActorTypeRoom, "broadcast", h.OnBroadcast)
}

type roomHandler struct {
    room IRoomModule `di:"auto"`
}

func (r *roomHandler) OnCreateRoom(session *cproto.Session, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
    return r.room.CreateRoom(session, req)
}

func (r *roomHandler) OnJoinRoom(session *cproto.Session, req *pb.JoinRoomRequest) (*pb.JoinRoomResponse, error) {
    return r.room.JoinRoom(session, req)
}

func (r *roomHandler) OnLeaveRoom(session *cproto.Session, req *pb.LeaveRoomRequest) error {
    return r.room.LeaveRoom(session, req)
}

func (r *roomHandler) OnGetRoomInfo(session *cproto.Session, req *pb.GetRoomInfoRequest) (*pb.GetRoomInfoResponse, error) {
    return r.room.GetRoomInfo(session, req)
}

func (r *roomHandler) OnBroadcast(session *cproto.Session, req *pb.RoomBroadcastRequest) error {
    return r.room.Broadcast(session, req)
}
```

## API 说明

### RegisterHandler[T]

注册返回 `error` 的处理器：

```go
handler.RegisterHandler(handler.ActorTypeRoom, "leaveRoom", h.OnLeaveRoom)
```

处理器签名：`func(session *cproto.Session, req *T) error`

### RegisterHandlerWithResponse[TReq, TResp]

注册返回 `(response, error)` 的处理器：

```go
handler.RegisterHandlerWithResponse(handler.ActorTypeRoom, "createRoom", h.OnCreateRoom)
```

处理器签名：`func(session *cproto.Session, req *TReq) (*TResp, error)`

### RegisterHandlerNoReturn[T]

注册无返回值的处理器：

```go
handler.RegisterHandlerNoReturn(handler.ActorTypeRoom, "broadcast", h.OnBroadcast)
```

处理器签名：`func(session *cproto.Session, req *T)`

## 性能对比

| 版本 | 注册时开销 | 调用时开销 | 类型转换 |
|------|-----------|-----------|---------|
| 旧版本 (MsgParam) | 低（手动类型断言） | 中（函数调用） | 手动 |
| V3 (泛型版本) | 低（编译时） | 极低（直接调用） | 自动 |

## 兼容性

- V3 版本与旧版本（MsgParam）可以共存
- `RegisterAllToActorByType` 会同时注册所有版本的处理器
- 建议新代码使用 V3 版本，旧代码可以逐步迁移

## 注意事项

1. **Go 版本要求**：需要 Go 1.18+ 支持泛型
2. **类型匹配**：处理器方法的第二个参数类型必须与注册时的泛型类型匹配
3. **性能优化**：cherry 框架已经处理了大部分类型转换，V3 版本主要做类型断言，性能极高

