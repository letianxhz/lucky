# 购买道具测试用例

本文档说明如何运行购买道具的测试用例。

## 测试用例说明

测试用例 `TestBuyItem` 会完整测试购买道具的流程，包括：

1. **登录流程**
   - 获取 token
   - 连接网关
   - 用户登录
   - 查看/创建角色
   - 角色进入游戏

2. **购买道具 - 成功案例**
   - 购买道具 1001 (金币支付) x1
   - 购买道具 1001 (金币支付) x2
   - 购买道具 1002 (钻石支付) x1

3. **购买道具 - 失败案例**
   - 无效道具ID (应该返回错误码 401)
   - 无效数量 (应该返回错误码 403)
   - 无效支付类型 (应该返回错误码 403)

## 运行测试

### 方法 1: 使用测试脚本

```bash
cd lucky/server
./robot_client/run_buy_item_test.sh
```

### 方法 2: 直接编译运行

```bash
cd lucky/server

# 编译测试程序
go build -o bin/robot_buy_item_test -ldflags "-X main.testBuyItem=true -X main.printLog=true" ./robot_client

# 运行测试
./bin/robot_buy_item_test
```

### 方法 3: 修改代码运行

在 `robot_client/main.go` 中设置：

```go
var (
    testBuyItem = true  // 设置为 true
    printLog    = true  // 设置为 true 查看详细日志
)
```

然后运行：

```bash
cd lucky/server
go run ./robot_client
```

## 前置条件

运行测试前，请确保以下服务已启动：

1. **Web 服务** (端口 8081)
   ```bash
   cd lucky/server
   ./bin/web
   ```

2. **Gate 服务** (端口 10011)
   ```bash
   cd lucky/server
   ./bin/gate
   ```

3. **Game 服务** (节点 10001)
   ```bash
   cd lucky/server
   NODE_ID=10001 ./bin/game
   ```

或者使用启动脚本：

```bash
cd lucky/server
./start_all.sh
```

## 测试账号

测试用例使用以下账号：
- 账号: `test_buy_item`
- 密码: `test_buy_item`

测试程序会自动注册该账号（如果不存在）。

## 预期结果

### 成功案例

每个成功案例应该返回：
- `itemId`: 购买的道具ID
- `count`: 购买数量
- `payType`: 支付类型
- `costAmount`: 消耗金额
- `items`: 获得的道具列表

### 失败案例

失败案例应该返回相应的错误码：
- `401`: 商店道具不存在
- `403`: 购买参数错误

## 日志输出

设置 `printLog = true` 后，会输出详细的测试日志，包括：
- 每个步骤的执行状态
- 请求和响应的详细信息
- 错误信息（如果有）

## 注意事项

1. 当前实现不涉及数据库操作，仅做流程演示
2. 道具价格是硬编码的（道具 1001 = 100，道具 1002 = 200）
3. 实际使用时需要从配置表读取价格，并检查玩家货币是否足够



