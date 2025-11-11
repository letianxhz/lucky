# Cherry Actor 消息处理机制详解

## 概述

Cherry 框架基于 skynet 的 Actor 模型，每个 Actor 独立运行在一个 goroutine 中，所有逻辑都是串行处理。Actor 接收三种类型的消息，每种消息都有自己的队列和处理机制。

## 三种消息类型

### 1. Local 消息（本地消息）
- **来源**：游戏客户端发送的消息
- **队列**：`localMail` (本地消息邮箱)
- **处理函数**：`processLocal()`
- **注册方式**：`actor.Local().Register(route, handler)`
- **发送方式**：`PostLocal(m *Message)`

### 2. Remote 消息（远程消息）
- **来源**：Actor 之间调用的消息
- **队列**：`remoteMail` (远程消息邮箱)
- **处理函数**：`processRemote()`
- **注册方式**：`actor.Remote().Register(funcName, handler)`
- **发送方式**：`PostRemote(m *Message)` 或 `Call/CallWait`

### 3. Event 消息（事件消息）
- **来源**：通过订阅/发布机制发送的事件
- **队列**：`event` (事件邮箱)
- **处理函数**：`processEvent()`
- **发送方式**：`PostEvent(data IEventData)`

## 消息路由机制

### 客户端消息 → Local 消息

```
客户端 → Agent → LocalDataRoute → PostLocal → localMail → processLocal → Local().Register 的处理器
```

**代码流程**：
1. 客户端发送消息到 Agent
2. `DefaultDataRoute` 解析路由
3. `LocalDataRoute` 创建 Message 并设置 `Session`
4. 调用 `PostLocal` 将消息放入 `localMail` 队列
5. Actor 的 `processLocal` 从队列取出消息
6. 调用 `Local().Register` 注册的处理器

**关键代码** (`cherry/net/parser/pomelo/pomelo.go`)：
```go
func LocalDataRoute(agent *Agent, session *cproto.Session, route *pmessage.Route, msg *pmessage.Message, targetPath string) {
    message := cfacade.GetMessage()
    message.Source = session.AgentPath
    message.Target = targetPath
    message.FuncName = route.Method()
    message.Session = session  // Local 消息包含 Session
    message.Args = msg.Data

    agent.ActorSystem().PostLocal(&message)  // 发送到 Local 队列
}
```

### Actor 间消息 → Remote 消息

```
Actor A → Call/CallWait → PostRemote → remoteMail → processRemote → Remote().Register 的处理器
```

**代码流程**：
1. Actor A 调用 `Call` 或 `CallWait` 发送消息
2. 如果是跨节点，通过集群组件发送；如果是本地，调用 `PostRemote`
3. `PostRemote` 将消息放入 `remoteMail` 队列
4. Actor B 的 `processRemote` 从队列取出消息
5. 调用 `Remote().Register` 注册的处理器

**关键代码** (`cherry/net/actor/system.go`)：
```go
func (p *System) Call(source, target, funcName string, arg any) int32 {
    // ... 参数验证 ...
    
    if targetPath.NodeID != "" && targetPath.NodeID != p.NodeID() {
        // 跨节点：通过集群发送
        err = p.app.Cluster().PublishRemote(targetPath.NodeID, clusterPacket)
    } else {
        // 本地：直接 PostRemote
        remoteMsg := cfacade.GetMessage()
        remoteMsg.Source = source
        remoteMsg.Target = target
        remoteMsg.FuncName = funcName
        remoteMsg.Args = arg  // Remote 消息没有 Session

        p.PostRemote(&remoteMsg)  // 发送到 Remote 队列
    }
}
```

## 如何区分玩家消息和 Actor 消息？

### 1. 通过注册方式区分

```go
// 注册处理客户端消息的处理器（Local）
actor.Local().Register("buyItem", func(session *cproto.Session, req *pb.BuyItemRequest) {
    // session 包含客户端连接信息
    // 可以调用 actor.Response(session, response) 返回给客户端
})

// 注册处理 Actor 间消息的处理器（Remote）
actor.Remote().Register("joinRoom", func(req interface{}) interface{} {
    // 没有 session，直接返回结果给调用方
    return &pb.JoinRoomResponse{...}
})
```

### 2. 通过消息队列区分

Actor 内部维护三个独立的消息队列：
- `localMail` - 客户端消息队列
- `remoteMail` - Actor 间消息队列
- `event` - 事件消息队列

Actor 的主循环通过 `select` 语句监听这三个队列：

```go
func (p *Actor) loop() bool {
    select {
    case <-p.localMail.C:
        p.processLocal()    // 处理客户端消息
    case <-p.remoteMail.C:
        p.processRemote()   // 处理 Actor 间消息
    case <-p.event.C:
        p.processEvent()    // 处理事件消息
    }
}
```

### 3. 通过消息结构区分

**Local 消息**：
- 包含 `Session` 字段（客户端连接信息）
- 可以通过 `Session` 获取客户端 UID、SID 等信息
- 可以调用 `actor.Response(session, data)` 返回给客户端

**Remote 消息**：
- **不包含** `Session` 字段
- 只有 `Args` 字段（参数）
- 返回值直接返回给调用方（通过 `CallWait` 的 `reply` 参数）

### 4. 通过处理函数签名区分

**Local 处理器签名**：
```go
func(session *cproto.Session, msg T)
// session 参数用于标识客户端连接
```

**Remote 处理器签名**：
```go
func(req interface{}) interface{}
// 没有 session，直接返回结果
```

## 实际应用示例

### 示例 1：Player Actor 处理客户端消息

```go
func (p *actorPlayer) OnInit() {
    // 注册客户端消息处理器（Local）
    handler.RegisterAllToActorByType(handler.ActorTypePlayer, &p.ActorBase)
    // 这会调用 actor.Local().Register("buyItem", ...)
    
    // 注册 Actor 间消息处理器（Remote）
    p.Remote().Register("sessionClose", p.sessionClose)
    // 网关断开连接时，会通过 Remote 消息通知
}
```

### 示例 2：Room Actor 同时处理两种消息

```go
func (r *ActorRoom) OnInit() {
    // 注册客户端消息处理器（Local）
    // 如果客户端直接发送消息到 Room Actor（通常不会，因为 Room 是子 Actor）
    handler.RegisterAllToActorByType(handler.ActorTypeRoom, &r.ActorBase)
    
    // 注册 Actor 间消息处理器（Remote）
    // Player Actor 通过 CallWait 调用这些方法
    r.Remote().Register("joinRoom", r.onRemoteJoinRoom)
    r.Remote().Register("leaveRoom", r.onRemoteLeaveRoom)
    r.Remote().Register("getRoomInfo", r.onRemoteGetRoomInfo)
}
```

### 示例 3：Player Actor 调用 Room Actor

```go
// Player Actor 的处理器（Local，来自客户端）
func OnJoinRoom(param *handler.MsgParam) {
    // 获取客户端 session
    session := param.GetSession()
    
    // 调用 Room Actor（Remote 消息）
    roomActorPath := cfacade.NewChildPath("", "rooms", roomId)
    var reply pb.JoinRoomResponse
    code := param.GetActor().CallWait(roomActorPath, "joinRoom", joinReq, &reply)
    // CallWait 会发送 Remote 消息到 Room Actor
    
    // 将结果返回给客户端
    param.GetActor().Response(session, &reply)
}
```

## 消息流程图

```
┌─────────────┐
│   客户端     │
└──────┬──────┘
       │ 发送消息
       ▼
┌─────────────┐
│   Agent     │
└──────┬──────┘
       │ LocalDataRoute
       ▼
┌─────────────┐      PostLocal      ┌──────────────┐
│ ActorSystem │ ──────────────────> │  localMail   │
└─────────────┘                     └──────┬───────┘
                                           │ processLocal
                                           ▼
                                    ┌──────────────┐
                                    │ Local().Register │
                                    │  的处理器      │
                                    └──────────────┘

┌─────────────┐
│  Actor A    │
└──────┬──────┘
       │ Call/CallWait
       ▼
┌─────────────┐      PostRemote     ┌──────────────┐
│ ActorSystem │ ──────────────────> │ remoteMail   │
└─────────────┘                     └──────┬───────┘
                                           │ processRemote
                                           ▼
                                    ┌──────────────┐
                                    │ Remote().Register │
                                    │  的处理器      │
                                    └──────────────┘
```

## 总结

1. **Local 消息**：来自客户端，包含 `Session`，通过 `Local().Register` 注册处理器
2. **Remote 消息**：来自其他 Actor，不包含 `Session`，通过 `Remote().Register` 注册处理器
3. **区分标志**：
   - 注册方式：`Local().Register` vs `Remote().Register`
   - 消息队列：`localMail` vs `remoteMail`
   - 处理函数：`processLocal` vs `processRemote`
   - 消息结构：Local 消息有 `Session`，Remote 消息没有
   - 处理器签名：Local 处理器有 `session` 参数，Remote 处理器没有

这种设计确保了客户端消息和 Actor 间消息的完全隔离，避免了消息混淆和安全性问题。



