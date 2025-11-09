# 服务职责说明文档

本文档详细说明 Lucky Server 项目中各个服务的职责和作用。

## 目录

- [整体架构](#整体架构)
- [服务列表](#服务列表)
  - [Gate 服务（网关服务）](#gate-服务网关服务)
  - [Game 服务（游戏服务）](#game-服务游戏服务)
  - [Center 服务（中心服务）](#center-服务中心服务)
  - [Master 服务（主节点服务）](#master-服务主节点服务)
  - [Web 服务（Web管理服务）](#web-服务web管理服务)
- [Game 服务模块说明](#game-服务模块说明)
  - [Player 模块组](#player-模块组)
  - [Room 模块组](#room-模块组)
  - [Alliance 模块组](#alliance-模块组)
  - [Shared 模块组](#shared-模块组)
- [服务间通信](#服务间通信)
- [依赖关系](#依赖关系)

---

## 整体架构

Lucky Server 采用微服务架构，基于 `cherry` 框架构建，使用 Actor 模型处理并发。整体架构如下：

```
┌─────────────────────────────────────────────────────────────┐
│                      Client (客户端)                          │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ WebSocket/TCP
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                    Gate Service (网关服务)                    │
│  - 客户端连接管理                                             │
│  - 消息路由转发                                               │
│  - 负载均衡                                                   │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ NATS (消息总线)
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
┌───────▼──────┐  ┌───────▼──────┐  ┌───────▼──────┐
│ Game Service │  │Center Service │  │Master Service│
│  (游戏服务)   │  │  (中心服务)   │  │  (主节点)    │
└──────────────┘  └───────────────┘  └──────────────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                           │
                  ┌────────▼────────┐
                  │  Redis/MongoDB  │
                  │   (数据存储)    │
                  └─────────────────┘
```

---

## 服务列表

### Gate 服务（网关服务）

**路径**: `cmd/gate/main.go`, `app/gate/`

**职责**:
- **客户端连接管理**: 管理所有客户端 WebSocket/TCP 连接
- **消息路由**: 接收客户端消息，根据路由规则转发到对应的 Game 服务节点
- **负载均衡**: 将客户端请求分发到不同的 Game 服务实例
- **会话管理**: 维护客户端会话状态，处理连接建立和断开
- **协议转换**: 处理客户端协议与内部服务协议的转换

**关键组件**:
- `actor/agent_actor.go`: 处理客户端连接的 Actor
- `service/router/`: 消息路由逻辑
- `service/middleware/`: 中间件（认证、限流等）

**启动方式**:
```bash
cd cmd/gate
go run main.go
# 或
NODE_ID=20001 go run main.go
```

---

### Game 服务（游戏服务）

**路径**: `cmd/game/main.go`, `app/game/`

**职责**:
- **游戏逻辑处理**: 处理所有游戏相关的业务逻辑
- **玩家管理**: 管理玩家 Actor，处理玩家登录、创建、进入游戏等
- **房间管理**: 管理游戏房间，处理房间创建、加入、离开等
- **联盟管理**: 处理联盟相关的业务逻辑
- **道具系统**: 处理道具购买、使用、扣除等
- **装备系统**: 处理装备穿戴、卸下等
- **在线状态管理**: 维护玩家在线状态

**关键组件**:
- `actor/`: Actor 层，包括 Player Actor、Room Actor、Alliance Actor 等
- `module/`: 业务模块层，包括 Item、Equipment、Login、Room、Alliance 等模块
- `db/`: 数据库访问层
- `config/`: 游戏配置

**Actor 类型**:
- **Player Actor**: 每个玩家一个实例，处理玩家相关的所有消息
- **Room Actor**: 每个房间一个实例，处理房间内的游戏逻辑
- **Alliance Actor**: 处理联盟相关的逻辑
- **Actor Players**: 管理所有 Player Actor 的父 Actor
- **Actor Rooms**: 管理所有 Room Actor 的父 Actor

**启动方式**:
```bash
cd cmd/game
go run main.go
# 或
NODE_ID=10001 go run main.go
```

---

### Center 服务（中心服务）

**路径**: `cmd/center/main.go`, `app/center/`

**职责**:
- **账号管理**: 处理用户注册、登录、账号绑定等
- **用户认证**: 生成和管理用户 Token
- **开发账号管理**: 管理开发测试账号
- **用户绑定管理**: 处理第三方账号绑定（如微信、QQ等）
- **运维功能**: 提供运维相关的接口和功能

**关键组件**:
- `actor/account/`: 账号管理 Actor
- `actor/ops/`: 运维功能 Actor
- `db/`: 数据库访问层（用户表、绑定表等）

**启动方式**:
```bash
cd cmd/center
go run main.go
# 或
NODE_ID=30001 go run main.go
```

---

### Master 服务（主节点服务）

**路径**: `cmd/master/main.go`, `app/master/`

**职责**:
- **集群管理**: 管理整个服务集群的节点信息
- **服务发现**: 提供服务发现和注册功能
- **配置管理**: 管理集群级别的配置
- **监控和统计**: 收集和统计各服务节点的运行状态

**启动方式**:
```bash
cd cmd/master
go run main.go
```

---

### Web 服务（Web管理服务）

**路径**: `cmd/web/main.go`, `app/web/`

**职责**:
- **Web 管理界面**: 提供 Web 管理后台
- **API 接口**: 提供 HTTP API 接口
- **静态资源**: 提供前端静态文件服务
- **管理功能**: 提供游戏数据查询、配置管理等管理功能

**关键组件**:
- `controller/`: HTTP 控制器
- `view/`: 视图模板
- `static/`: 静态资源（CSS、JS等）
- `sdk/`: SDK 封装

**启动方式**:
```bash
cd cmd/web
go run main.go
```

---

## Game 服务模块说明

Game 服务采用模块化设计，每个模块负责特定的业务功能。模块通过依赖注入（DI）容器管理，支持自动依赖注入。

### Player 模块组

**路径**: `app/game/module/player/`

#### Login 模块 (`player/login/`)

**职责**:
- **玩家登录**: 处理玩家登录逻辑，验证 Token
- **角色创建**: 处理新角色创建，验证角色名唯一性
- **角色选择**: 处理角色选择，返回角色列表
- **进入游戏**: 处理玩家进入游戏，初始化玩家数据

**关键文件**:
- `login.go`: 接口定义 `ILoginModule`
- `login_impl.go`: 实现类 `LoginModule`
- `handler.go`: 消息处理器，注册 `create`、`select`、`enter` 等路由

**消息路由**:
- `game.player.create`: 创建角色
- `game.player.select`: 选择角色
- `game.player.enter`: 进入游戏

#### Item 模块 (`player/item/`)

**职责**:
- **道具管理**: 管理玩家道具的增删改查
- **道具购买**: 处理道具购买逻辑（目前为占位实现）
- **道具检查**: 检查玩家是否拥有足够的道具
- **批量操作**: 支持批量添加和扣除道具

**关键文件**:
- `item.go`: 接口定义 `IItemModule`
- `item_impl.go`: 实现类 `ItemModule`（当前使用内存存储，用于测试）
- `handler.go`: 消息处理器，注册 `buyItem` 路由

**消息路由**:
- `game.player.buyItem`: 购买道具

**依赖关系**:
- 被 `EquipmentModule` 依赖（装备道具时需要扣除道具）

#### Equipment 模块 (`player/equipment/`)

**职责**:
- **装备管理**: 管理玩家装备的穿戴和卸下
- **装备查询**: 查询玩家当前装备信息
- **装备逻辑**: 处理装备道具时的道具扣除逻辑

**关键文件**:
- `equipment.go`: 接口定义 `IEquipmentModule`
- `equipment_impl.go`: 实现类 `EquipmentModule`

**依赖关系**:
- 依赖 `ItemModule`（通过 `di:"auto"` 自动注入）

#### Player Handler (`player/handler.go`)

**职责**:
- **房间相关消息**: 处理玩家与房间的交互
- **Actor 通信**: 作为 Player Actor 与 Room Actor 之间的桥梁

**消息路由**:
- `game.player.joinRoom`: 加入房间
- `game.player.leaveRoom`: 离开房间
- `game.player.getRoomInfo`: 获取房间信息

---

### Room 模块组

**路径**: `app/game/module/room/room/`

#### Room 模块 (`room/room/`)

**职责**:
- **房间管理**: 管理游戏房间的创建、加入、离开
- **房间信息**: 维护房间状态（玩家列表、房间配置等）
- **房间广播**: 处理房间内的消息广播

**关键文件**:
- `room.go`: 接口定义 `IRoomModule`
- `room_impl.go`: 实现类 `RoomModule`
- `handler.go`: 消息处理器（目前房间消息通过 Player Actor 间接调用）

**Actor 通信**:
- Room Actor 通过 Remote 方法接收来自 Player Actor 的调用
- 方法包括: `joinRoom`、`leaveRoom`、`getRoomInfo`、`broadcast`

---

### Alliance 模块组

**路径**: `app/game/module/alliance/alliance/`

#### Alliance 模块 (`alliance/alliance/`)

**职责**:
- **联盟管理**: 处理联盟的创建、加入、离开
- **联盟信息**: 维护联盟状态和信息
- **联盟操作**: 提供联盟相关的业务操作接口

**关键文件**:
- `alliance.go`: 接口定义 `IAllianceModule`
- `alliance_impl.go`: 实现类 `AllianceModule`（当前为占位实现）
- `handler.go`: 消息处理器

**消息路由**:
- `game.alliance.createAlliance`: 创建联盟
- `game.alliance.joinAlliance`: 加入联盟
- `game.alliance.leaveAlliance`: 离开联盟
- `game.alliance.getAllianceInfo`: 获取联盟信息

---

### Shared 模块组

**路径**: `app/game/module/shared/`

#### Handler 模块 (`shared/handler/`)

**职责**:
- **消息处理器注册**: 提供统一的消息处理器注册机制
- **类型安全**: 提供类型安全的消息处理器注册（V3 泛型版本）
- **路由管理**: 管理消息路由到对应的 Actor 和处理器

**关键文件**:
- `receiver_v3.go`: V3 版本的消息处理器注册（泛型，类型安全，无反射）
- `receiver.go`: 统一入口，调用 V3 版本
- `handler.go`: Handler 相关的基础定义
- `actor_type.go`: Actor 类型定义

**特性**:
- 支持多种处理器签名（带返回值、只返回错误、无返回值等）
- 自动类型转换，无需手动类型断言
- 支持错误码返回（`ErrorWithCode`）
- 每个 Actor 实例只注册一次，避免重复注册

#### Online 模块 (`shared/online/`)

**职责**:
- **在线状态管理**: 维护玩家在线状态
- **会话绑定**: 管理玩家 ID 与 Session 的绑定关系
- **在线统计**: 提供在线玩家数量统计

**关键函数**:
- `BindPlayer(playerId int64, session *cproto.Session)`: 绑定玩家
- `UnBindPlayer(playerId int64)`: 解绑玩家
- `Count()`: 获取在线玩家数量

---

## 服务间通信

### 客户端到服务

```
Client → Gate Service → Game Service
```

1. 客户端通过 WebSocket/TCP 连接到 Gate 服务
2. Gate 服务接收消息，根据路由规则转发到对应的 Game 服务节点
3. Game 服务处理消息，返回响应

### 服务间通信

服务间通过 **NATS** 消息总线进行通信：

- **Gate ↔ Game**: Gate 服务将客户端消息转发到 Game 服务
- **Game ↔ Center**: Game 服务调用 Center 服务进行账号验证等操作
- **Game ↔ Game**: 不同 Game 服务节点之间可以通过 NATS 进行通信

### Actor 间通信

Game 服务内部，Actor 之间通过 `actor.Call` 和 `actor.CallWait` 进行通信：

```go
// Player Actor 调用 Room Actor
roomActorPath := cfacade.NewChildPath("rooms", "room_001")
reply, err := actor.CallWait(roomActorPath, "joinRoom", joinReq)
```

---

## 依赖关系

### 服务依赖

```
Gate Service
  └─→ Game Service (通过 NATS)
  └─→ Center Service (通过 NATS)

Game Service
  └─→ Center Service (账号验证)
  └─→ Redis (缓存、在线状态)
  └─→ MongoDB (数据存储)

Center Service
  └─→ Redis (缓存)
  └─→ MongoDB (数据存储)
```

### 模块依赖（Game 服务内部）

```
EquipmentModule
  └─→ ItemModule (通过 DI 注入)

Handler (各模块的 handler)
  └─→ 对应的 Module (通过 DI 注入)
    ├─→ itemHandler → ItemModule
    ├─→ loginHandler → LoginModule
    ├─→ allianceHandler → AllianceModule
    └─→ roomHandler → RoomModule
```

### DI 容器初始化流程

1. **注册阶段**（`init()` 函数）:
   - 各模块在 `init()` 中调用 `di.Register(v)` 注册实例
   - Handler 在 `init()` 中调用 `handler.RegisterHandler()` 注册消息处理器

2. **初始化阶段**（`di.MustInitialize()`）:
   - 遍历所有已注册的实例
   - 检查每个实例是否有需要注入的字段（通过 `di:"auto"` 标签）
   - 自动查找依赖并注入

3. **运行阶段**:
   - Actor 启动时，调用 `handler.RegisterAllToActorByType()` 注册消息处理器
   - 消息到达时，通过注册的处理器路由到对应的 Module 处理

---

## 总结

Lucky Server 采用微服务架构，各服务职责明确：

- **Gate**: 客户端连接和消息路由
- **Game**: 游戏逻辑处理（核心服务）
- **Center**: 账号和用户管理
- **Master**: 集群管理
- **Web**: Web 管理界面

Game 服务内部采用模块化设计，通过 DI 容器管理依赖，支持自动注入，便于扩展和维护。

