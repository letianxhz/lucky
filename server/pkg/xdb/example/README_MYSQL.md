# MySQL 驱动使用说明

## 概述

xdb 现在支持 MySQL 数据库驱动。本文档说明如何配置和使用 MySQL 驱动。

## 文件结构

```
example/
├── mysql_config.go      # MySQL 配置器
├── mysql_main.go         # MySQL 测试程序
├── test_mysql.sh         # MySQL 测试脚本
└── ...
```

## 配置步骤

### 1. 在 proto 文件中指定 MySQL 驱动

在 `player.proto` 中设置：

```protobuf
message Player {
  option (xdb.table) = "player";
  option (xdb.driver) = DRIVER_MYSQL;  // 使用 MySQL 驱动
  // ...
}
```

### 2. 生成代码

```bash
./generate.sh
```

### 3. 修复驱动名称（如果需要）

如果代码生成器没有正确识别驱动，可以手动修复：

```bash
./fix_driver.sh mysql
```

或者手动编辑 `pb/player_xdb.pb.go`，将 `DriverName: "none"` 改为 `DriverName: "mysql"`。

### 4. 配置 MySQL 连接

#### 方式 1: 使用环境变量

```bash
export MYSQL_DB=test
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=your_password
```

#### 方式 2: 在代码中配置

```go
configurator := NewMySQLConfigurator(
    "test",      // 数据库名
    "localhost", // 主机
    3306,        // 端口
    "root",      // 用户名
    "password",  // 密码
)
```

### 5. 运行测试

```bash
./test_mysql.sh
```

或者直接运行：

```bash
go run mysql_main.go mysql_config.go pb/player.pb.go pb/player_xdb.pb.go
```

## MySQL 配置选项

### DaoOptions

```go
type DaoOptions struct {
    DBName       string        // 数据库名
    Host         string        // 主机地址
    Port         int32         // 端口（默认 3306）
    Username     string        // 用户名
    Password     string        // 密码
    Charset      string        // 字符集（默认 utf8mb4）
    MaxOpenConns int32         // 最大打开连接数（默认 10）
    QueryTimeout time.Duration // 查询超时时间（默认 5s）
}
```

### TableOptions

```go
type TableOptions struct {
    DaoKey       interface{}   // DAO 键（用于标识不同的数据库连接）
    Concurrence  uint32        // 存储并发协程数
    SaveTimeout  time.Duration // 存储超时时间
    SyncInterval time.Duration // 存储队列刷新间隔
}
```

## 注意事项

1. **数据库表结构**: MySQL 驱动需要数据库表已经存在。可以手动创建表，或者实现自动建表逻辑。

2. **字符集**: 默认使用 `utf8mb4`，支持完整的 UTF-8 字符集。

3. **连接池**: 默认最大打开连接数为 10，可以根据实际需求调整。

4. **事务**: 当前实现是简化版本，完整的 MySQL 驱动可能需要支持事务等高级特性。

## 与 MongoDB 的对比

| 特性 | MySQL | MongoDB |
|------|-------|---------|
| 驱动名称 | `mysql` | `mongo` |
| 连接方式 | DSN (Data Source Name) | URI |
| 数据格式 | 关系型（表） | 文档型（集合） |
| 查询语言 | SQL | MongoDB Query |

## 故障排除

### 问题 1: 连接失败

**错误**: `failed to ping MySQL`

**解决**:
- 检查 MySQL 服务是否运行
- 检查连接参数（主机、端口、用户名、密码）
- 检查防火墙设置
- 检查 MySQL 用户权限

### 问题 2: 表不存在

**错误**: `Table 'xxx' doesn't exist`

**解决**:
- 手动创建数据库表
- 或者实现自动建表逻辑

### 问题 3: 驱动未注册

**错误**: `invalid driver: mysql`

**解决**:
- 确保导入了 MySQL 驱动包：`_ "lucky/server/pkg/xdb/storage/mysql"`
- 检查 `init()` 函数是否正确注册了驱动

## 示例代码

完整示例请参考：
- `mysql_config.go` - 配置器实现
- `mysql_main.go` - 测试程序



