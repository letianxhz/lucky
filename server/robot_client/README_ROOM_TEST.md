# Room 功能测试文档

本文档描述了如何使用机器人客户端测试 Room（房间）相关的功能。

## 功能概述

Room 测试包括以下功能：
1. **创建房间** - 测试创建房间功能
2. **加入房间** - 测试多个机器人同时加入房间
3. **获取房间信息** - 测试获取房间详细信息
4. **房间广播** - 测试房间内消息广播
5. **离开房间** - 测试机器人离开房间
6. **状态验证** - 验证离开后的房间状态

## 测试场景

### 场景 1: 创建房间
- 第一个机器人创建房间
- 验证房间创建成功

### 场景 2: 多个机器人加入房间
- 所有机器人并发加入房间
- 验证每个机器人都能成功加入
- 验证房间玩家数量正确

### 场景 3: 获取房间信息
- 获取房间详细信息
- 验证房间ID、玩家数量、最大玩家数、玩家ID列表

### 场景 4: 房间广播
- 在房间内发送广播消息
- 验证广播功能正常

### 场景 5: 部分机器人离开房间
- 一半机器人离开房间
- 验证离开功能正常

### 场景 6: 验证离开后的房间状态
- 再次获取房间信息
- 验证玩家数量减少

### 场景 7: 剩余机器人全部离开
- 所有剩余机器人离开房间
- 验证房间清空

## 使用方法

### 方法 1: 使用 Shell 脚本（推荐）

```bash
cd lucky/server
./robot_client/test_room.sh
```

### 方法 2: 手动编译运行

```bash
cd lucky/server
go build -o bin/robot_test_room -ldflags "-X main.testRoom=true -X main.printLog=true" ./robot_client
./bin/robot_test_room
```

### 方法 3: 修改 main.go 配置

在 `robot_client/main.go` 中设置：

```go
testRoom = true  // 设置为 true
```

然后运行：

```bash
cd lucky/server
go run ./robot_client
```

## 测试账号

测试使用以下账号（会自动注册）：
- `test_room_1` / `test_room_1`
- `test_room_2` / `test_room_2`
- `test_room_3` / `test_room_3`
- `test_room_4` / `test_room_4`

## 预期结果

测试成功时，应该看到：
- ✓ 所有机器人初始化成功
- ✓ 创建房间成功
- ✓ 所有机器人加入房间成功
- ✓ 获取房间信息成功，显示正确的玩家数量
- ✓ 房间广播成功
- ✓ 机器人离开房间成功
- ✓ 房间状态更新正确

## 日志输出

测试日志会输出到：
- 控制台（实时）
- `/tmp/robot_test_room.log`（完整日志）

## 注意事项

1. 确保游戏服务器已启动
2. 确保网关服务器已启动
3. 确保 Web 节点已启动
4. 测试账号会自动注册，如果已存在会跳过注册
5. 每个机器人会创建唯一的角色名（使用时间戳）

## 故障排查

如果测试失败，检查：
1. 服务器是否正常运行
2. 网络连接是否正常
3. 日志中的错误信息
4. 房间模块是否正确注册到 DI 容器
5. Room Actor 是否正确注册

## 相关文件

- `robot_client/test_room.go` - Room 测试实现
- `robot_client/test_room.sh` - 测试脚本
- `robot_client/main.go` - 主入口，包含测试标志
- `app/game/module/room/` - Room 模块实现
- `app/game/actor/room/` - Room Actor 实现
- `pkg/protocol/room.proto` - Room 协议定义

