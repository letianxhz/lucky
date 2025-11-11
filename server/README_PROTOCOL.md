# 通信协议文档

本文档描述了服务器与客户端之间的通信协议定义。

## 协议文件位置

所有协议定义文件位于 `pkg/protocol/` 目录下，使用 Protocol Buffers (protobuf) 格式定义。

## 已实现的协议

### 1. 登录认证协议

**文件**: `pkg/protocol/login.proto`

#### 登录请求 (LoginRequest)
- **路由**: `gate.user.login`
- **请求参数**:
  - `serverId` (int32): 当前登录的服务器ID
  - `token` (string): 登录token（web login api生成的base64字符串）
  - `params` (map<int32, string>): 登录时上传的参数

#### 登录响应 (LoginResponse)
- **响应参数**:
  - `uid` (int64): 游戏内的用户唯一ID
  - `pid` (int32): 平台ID
  - `openId` (string): 平台openId（平台的账号唯一ID）
  - `params` (map<int32, string>): 登录后的扩展参数

**处理逻辑**: `app/gate/actor/agent_actor.go` 中的 `login` 方法

### 2. 购买道具协议

**文件**: `pkg/protocol/shop.proto`

#### 购买道具请求 (BuyItemRequest)
- **路由**: `game.player.buyItem`
- **请求参数**:
  - `shopId` (int32): 商店ID
  - `itemId` (int32): 道具ID
  - `count` (int32): 购买数量
  - `payType` (int32): 支付类型（1:金币, 2:钻石, 3:仙玉等）

#### 购买道具响应 (BuyItemResponse)
- **响应参数**:
  - `itemId` (int32): 道具ID
  - `count` (int32): 购买数量
  - `payType` (int32): 支付类型
  - `costAmount` (int64): 消耗金额
  - `items` (map<int32, int64>): 获得的道具列表 (itemId -> count)

**处理逻辑**: `app/game/actor/player/actor_player.go` 中的 `buyItem` 方法

**错误码**:
- `401`: 商店道具不存在 (ShopItemNotFound)
- `402`: 货币不足 (ShopItemNotEnoughMoney)
- `403`: 购买参数错误 (ShopItemInvalidParam)
- `404`: 购买失败 (ShopItemBuyFail)

## 协议编译

### 编译 Go 代码

```bash
# 进入协议目录
cd pkg/protocol

# 编译所有 proto 文件生成 Go 代码
protoc --go_out=../pb --go_opt=paths=source_relative *.proto
```

### 编译 JavaScript 代码（用于前端）

```bash
# 使用 build_js_protocol.bat 脚本
./build_js_protocol.bat
```

## 路由规则

### Gate 节点路由

- `gate.user.login`: 用户登录（必须在建立连接后第一条消息）

### Game 节点路由

- `game.player.select`: 查询玩家角色
- `game.player.create`: 创建玩家角色
- `game.player.enter`: 玩家角色进入游戏
- `game.player.buyItem`: 购买道具

## 使用示例

### 登录流程

1. 客户端建立连接（TCP/WebSocket）
2. 发送登录请求：`gate.user.login`，携带 `LoginRequest`
3. 服务端验证 token，返回 `LoginResponse`
4. 客户端收到响应后，可以继续后续操作

### 购买道具流程

1. 客户端必须先完成登录和角色进入游戏
2. 发送购买请求：`game.player.buyItem`，携带 `BuyItemRequest`
3. 服务端验证参数，计算价格，返回 `BuyItemResponse`
4. 客户端收到响应，更新本地道具数据

## 注意事项

1. **登录顺序**: 必须先完成用户登录（`gate.user.login`），然后完成角色登录（`game.player.enter`），才能进行游戏内操作
2. **参数验证**: 所有请求都会进行参数验证，无效参数会返回相应的错误码
3. **DB 操作**: 当前购买道具的实现不涉及数据库操作，仅做流程演示。实际使用时需要：
   - 查询商店配置表获取道具价格
   - 检查玩家货币是否足够
   - 扣除玩家货币
   - 添加道具到玩家背包





