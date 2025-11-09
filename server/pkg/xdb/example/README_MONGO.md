# MongoDB 测试指南

## 概述

本指南介绍如何使用 MongoDB 驱动测试 xdb 模块。

## 前置要求

1. **MongoDB 服务**
   - 确保 MongoDB 服务正在运行
   - 默认连接: `mongodb://localhost:27017`
   - 可以通过环境变量 `MONGO_URI` 自定义

2. **Go 依赖**
   - `go.mongodb.org/mongo-driver` (MongoDB Go 驱动)

## 文件说明

### MongoDB 驱动实现

- `../storage/mongo/driver.go` - MongoDB 驱动实现
- `../storage/mongo/dao.go` - DAO 实现
- `../storage/mongo/table.go` - Table 实现

### 测试文件

- `mongo_config.go` - MongoDB 配置器
- `mongo_main.go` - MongoDB 测试主程序
- `test_mongo.sh` - 测试脚本

## 快速开始

### 方式一：使用测试脚本（推荐）

```bash
cd lucky/server/pkg/xdb/example
./test_mongo.sh
```

### 方式二：手动运行

```bash
cd lucky/server/pkg/xdb/example

# 1. 生成代码
./generate.sh

# 2. 修改生成的代码（将 DriverName 改为 "mongo"）
# 可以使用 sed 或手动编辑 player_xdb.pb.go

# 3. 运行测试
export MONGO_URI=mongodb://localhost:27017  # 可选
go run mongo_main.go mongo_config.go player.pb.go player_xdb.pb.go
```

## 配置说明

### MongoConfigurator

`MongoConfigurator` 实现了 `xdb.Configurator` 和 `xdb.DatabaseConfig` 接口：

```go
type MongoConfigurator struct {
    MongoURI string // MongoDB 连接 URI
}
```

### 配置项

- **MongoURI**: MongoDB 连接 URI，默认为 `mongodb://localhost:27017`
- **PoolSize**: 连接池大小，默认 10
- **Concurrence**: 并发保存协程数，默认 2
- **SaveTimeout**: 保存超时时间，默认 5 秒

## 测试内容

测试程序会执行以下操作：

1. 初始化 xdb（使用 MongoDB 驱动）
2. 检查 Source 注册
3. 测试 PK 创建
4. 创建 Player 记录
5. 保存到 MongoDB
6. 同步保存
7. 从 MongoDB 获取记录
8. 清理资源

## 故障排除

### 1. 连接失败

**错误**: `failed to connect to MongoDB`

**解决**:
- 检查 MongoDB 服务是否运行: `mongosh` 或 `mongo`
- 检查连接 URI 是否正确
- 检查防火墙设置

### 2. 驱动未注册

**错误**: `invalid driver: mongo`

**解决**:
- 确保导入了 MongoDB 驱动包: `_ "lucky/server/pkg/xdb/storage/mongo"`
- 检查驱动是否正确注册

### 3. 代码生成问题

**错误**: `DriverName: "mysql"` 而不是 `"mongo"`

**解决**:
- 手动修改生成的 `player_xdb.pb.go` 文件
- 将 `DriverName: "none"` 或 `DriverName: "mysql"` 改为 `DriverName: "mongo"`

## 下一步

1. **完善驱动实现**: 添加更多 MongoDB 特性支持
2. **添加索引支持**: 自动创建 MongoDB 索引
3. **添加事务支持**: 支持 MongoDB 事务
4. **性能优化**: 优化批量操作性能

