# Actor 通信测试说明

## 测试目标

测试 Player Actor 和 Room Actor 之间的通信，验证：
1. Player Actor 可以调用 Room Actor 的 Remote 方法
2. Room Actor 可以正确处理来自 Player Actor 的调用
3. 消息处理器注册机制正常工作

## 测试流程

1. **玩家登录**: 获取 token，连接网关，用户登录
2. **角色操作**: 查看角色列表，创建角色（如果没有），进入游戏
3. **Actor 通信测试**:
   - 测试 1: Player Actor 调用 Room Actor 的 `joinRoom` 方法
   - 测试 2: Player Actor 调用 Room Actor 的 `getRoomInfo` 方法
   - 测试 3: Player Actor 调用 Room Actor 的 `leaveRoom` 方法

## 运行测试

### 方式 1: 直接运行

```bash
cd lucky/server
go run ./robot_client
```

确保 `testActorComm = true` 在 `main.go` 中。

### 方式 2: 编译后运行

```bash
cd lucky/server
go build -o bin/robot_test_actor_comm ./robot_client
./bin/robot_test_actor_comm
```

## 测试消息路由

- `game.player.joinRoom` - 玩家加入房间（Player Actor -> Room Actor）
- `game.player.getRoomInfo` - 获取房间信息（Player Actor -> Room Actor）
- `game.player.leaveRoom` - 玩家离开房间（Player Actor -> Room Actor）

## 预期结果

1. **加入房间**: Room Actor 收到 `joinRoom` 调用，返回成功响应
2. **获取房间信息**: Room Actor 收到 `getRoomInfo` 调用，返回房间信息
3. **离开房间**: Room Actor 收到 `leaveRoom` 调用，返回成功响应

## 日志观察

测试过程中会输出以下关键日志：

### Player Actor 侧
- `[PlayerHandler] Player X requesting to join room room_001`
- `[PlayerHandler] Player joined room successfully: ...`
- `[PlayerHandler] Got room info: ...`
- `[PlayerHandler] Player left room successfully`

### Room Actor 侧
- `[ActorRoom] Received joinRoom request from another actor: ...`
- `[ActorRoom] Received getRoomInfo request from another actor: ...`
- `[ActorRoom] Received leaveRoom request from another actor: ...`

## 架构说明

### Actor 路径

- **Player Actor**: `player.{playerId}` (child actor)
- **Room Actor**: `rooms.room_001` (child actor)

### 通信方式

1. **Player Actor 调用 Room Actor**:
   ```go
   roomActorPath := cfacade.NewChildPath("", "rooms", roomId)
   code := param.GetActor().CallWait(roomActorPath, "joinRoom", req, &reply)
   ```

2. **Room Actor 接收调用**:
   ```go
   r.Remote().Register("joinRoom", r.onRemoteJoinRoom)
   ```

### 消息处理器注册

- **Player Actor**: 在 `module/player/handler.go` 的 `init()` 中注册
- **Room Actor**: 在 `actor/room/actor_room.go` 的 `OnInit()` 中注册 Remote 处理器

## 故障排查

1. **Actor 路径错误**: 检查 `NewChildPath` 的参数是否正确
2. **Remote 方法未注册**: 检查 Room Actor 的 `OnInit()` 是否注册了 Remote 处理器
3. **消息路由错误**: 检查客户端发送的路由是否正确（`game.player.joinRoom`）

## 扩展测试

可以添加更多测试场景：
- 多个玩家同时加入房间
- 房间满员时的处理
- 房间不存在时的错误处理
- 跨节点 Actor 通信（如果配置了集群）

