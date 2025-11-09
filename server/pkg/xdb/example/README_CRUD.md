# xdb CRUD 操作测试

## 概述

这是一个完整的 xdb CRUD（Create、Read、Update、Delete）操作测试示例，演示如何使用 xdb 进行数据库操作。

## 文件说明

- `crud_main.go` - CRUD 测试主程序
- `mysql_config.go` - MySQL 配置器
- `build_crud.sh` - 构建脚本
- `run_crud.sh` - 运行脚本
- `pb/player.sql` - 数据库表结构 SQL

## 使用前准备

### 1. 确保 MySQL 服务运行

```bash
# 检查 MySQL 服务状态
mysql --version

# 启动 MySQL（如果未运行）
# macOS: brew services start mysql
# Linux: sudo systemctl start mysql
```

### 2. 创建数据库和表

```bash
# 创建数据库
mysql -uroot -p -e 'CREATE DATABASE IF NOT EXISTS test;'

# 创建表（使用生成的 SQL 脚本）
mysql -uroot -p test < pb/player.sql
```

### 3. 配置环境变量（可选）

```bash
export MYSQL_DB=test
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=your_password
```

## 运行测试

### 方式一：使用运行脚本

```bash
./run_crud.sh
```

### 方式二：直接运行

```bash
./crud_test
```

## CRUD 操作说明

### 1. CREATE - 创建记录

```go
player, err := xdb.Create[*pb.PlayerRecord](ctx, &pb.Player{
    PlayerId: 10001,
    Name:     "测试玩家",
    Level:    1,
    Exp:      0,
    Ctime:    time.Now().Unix(),
    Mtime:    time.Now().Unix(),
})
```

### 2. READ - 读取记录

```go
player, err := xdb.Get[*pb.PlayerRecord](ctx, playerId)
```

### 3. UPDATE - 更新记录

```go
// 修改字段
player.Name = "更新后的玩家"
player.Level = 10
player.Exp = 1000

// 标记变更的字段
player.GetHeader().SetChanged(
    pb.PlayerFieldName,
    pb.PlayerFieldLevel,
    pb.PlayerFieldExp,
)

// 保存变更
xdb.Save(ctx, player)

// 同步到数据库
xdb.Sync(ctx, player)
```

### 4. DELETE - 删除记录

```go
// 标记为删除
deleted := player.Delete(ctx)

// 保存删除操作
xdb.Save(ctx, player)

// 同步到数据库
xdb.Sync(ctx, player)
```

## 测试流程

1. **初始化 xdb** - 连接 MySQL 数据库
2. **CREATE** - 创建玩家记录（ID: 10001）
3. **READ** - 读取刚创建的玩家记录
4. **UPDATE** - 更新玩家信息（名称、等级、经验）
5. **验证更新** - 再次读取验证更新是否成功
6. **DELETE** - 删除玩家记录
7. **验证删除** - 尝试读取已删除的记录（应该失败）
8. **清理资源** - 关闭连接

## 预期输出

```
=== xdb CRUD 操作测试 ===
MySQL 配置: root@localhost:3306/test

步骤 1: 初始化 xdb...
   ✓ xdb 初始化成功

步骤 2: CREATE - 创建玩家记录...
   ✓ 创建成功: ID=10001, Name=测试玩家, Level=1
   同步保存到数据库...
   ✓ 同步成功

步骤 3: READ - 获取玩家记录...
   ✓ 获取成功: ID=10001, Name=测试玩家, Level=1, Exp=0

步骤 4: UPDATE - 更新玩家记录...
   ✓ 更新成功: Name=更新后的玩家, Level=10, Exp=1000
   ✓ 同步成功

步骤 5: 验证更新 - 再次读取玩家记录...
   ✓ 验证成功: Name=更新后的玩家, Level=10, Exp=1000
   ✓ 数据更新正确

步骤 6: DELETE - 删除玩家记录...
   ✓ 标记删除成功
   ✓ 删除同步成功

步骤 7: 验证删除 - 尝试再次获取玩家记录...
   ✓ 获取失败（预期）: ...
   ✓ 删除验证成功：记录已不存在

步骤 8: 清理资源...
   ✓ 清理完成

=== CRUD 测试完成 ===
```

## 故障排除

### 问题 1: 连接失败

**错误**: `failed to connect to MySQL`

**解决**:
- 检查 MySQL 服务是否运行
- 检查连接参数（主机、端口、用户名、密码）
- 检查防火墙设置

### 问题 2: 表不存在

**错误**: `Table 'test.player' doesn't exist`

**解决**:
- 执行 SQL 脚本创建表: `mysql -uroot -p test < pb/player.sql`

### 问题 3: 权限错误

**错误**: `Access denied for user`

**解决**:
- 检查 MySQL 用户权限
- 确保用户有 CREATE、INSERT、UPDATE、DELETE 权限

## 代码说明

### 字段变更标记

在更新记录时，必须使用 `SetChanged` 标记变更的字段：

```go
player.GetHeader().SetChanged(
    pb.PlayerFieldName,    // 标记 name 字段已变更
    pb.PlayerFieldLevel,   // 标记 level 字段已变更
    pb.PlayerFieldExp,     // 标记 exp 字段已变更
)
```

### 同步保存

`xdb.Save` 是异步保存，`xdb.Sync` 会等待数据真正写入数据库：

```go
xdb.Save(ctx, player)      // 异步保存
xdb.Sync(ctx, player)       // 同步等待保存完成
```

## 扩展

可以基于此示例扩展更多功能：

- 批量操作（GetMulti、SaveMulti）
- 查询操作（Find、GetAll）
- 事务处理
- 错误处理和重试
- 性能测试

