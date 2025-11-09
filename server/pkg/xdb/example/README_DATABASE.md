# xdb 数据库配置初始化

## 概述

参考 `orm.go` 的实现，xdb 现在支持在启动时自动初始化数据库配置。

## 功能说明

### DatabaseConfig 接口

`DatabaseConfig` 是一个可选接口，如果配置器实现了此接口，xdb 会在驱动初始化之前自动调用 `InitializeDatabase()` 方法。

```go
type DatabaseConfig interface {
    InitializeDatabase() error
}
```

### 使用方式

#### 1. 实现 DatabaseConfig 接口

```go
type MyConfigurator struct {
    // 配置字段
}

// 实现 DatabaseConfig 接口
func (c *MyConfigurator) InitializeDatabase() error {
    // 初始化数据库连接
    database.MustInitialize(config.Get().GetDatabase())
    return nil
}

// 实现其他 Configurator 接口方法
func (c *MyConfigurator) RedoOptions() *xdb.RedoOptions { ... }
// ...
```

#### 2. 在 Setup 中自动调用

当调用 `xdb.Setup(ctx, configurator)` 时，如果配置器实现了 `DatabaseConfig` 接口，会自动在驱动初始化之前调用 `InitializeDatabase()`。

#### 3. 使用 MustInitialize（推荐）

类似于 `orm.MustInitialize`，xdb 也提供了 `MustInitialize` 函数：

```go
import "lucky/server/pkg/xdb"

func main() {
    ctx := context.Background()
    configurator := &MyConfigurator{}
    
    // 如果失败会 panic
    xdb.MustInitialize(ctx, configurator)
}
```

## 初始化顺序

xdb 的初始化顺序如下：

1. **数据库初始化**（如果实现了 `DatabaseConfig`）
   - 调用 `InitializeDatabase()`
   - 初始化数据库连接池等

2. **收集驱动和表配置**
   - 遍历所有注册的 Source
   - 收集驱动、DAO、表配置信息

3. **初始化驱动**
   - 调用 `driver.Init(ctx, driverOptions)`

4. **验证 Source**
   - 调用 `driver.Validate(src)`

5. **创建 DAO**
   - 调用 `driver.NewDao(ctx, daoOptions)`

6. **初始化表和 Source**
   - 初始化每个 Source 的表和缓存

7. **恢复数据**
   - 从重做日志恢复数据

8. **启动保存器**
   - 启动异步保存协程

## 示例

### 测试环境配置

```go
type TestConfigurator struct{}

func (c *TestConfigurator) InitializeDatabase() error {
    // 测试环境不需要初始化数据库
    return nil
}
```

### 生产环境配置

```go
type ProductionConfigurator struct {
    config interface{} // 配置对象
}

func (c *ProductionConfigurator) InitializeDatabase() error {
    // 初始化数据库连接
    database.MustInitialize(c.config.GetDatabase())
    return nil
}
```

参考 `production_config.go` 查看完整示例。

## 注意事项

1. **可选实现**：`DatabaseConfig` 接口是可选的，如果不实现，xdb 会跳过数据库初始化步骤

2. **错误处理**：如果 `InitializeDatabase()` 返回错误，`Setup` 会立即返回错误，不会继续初始化

3. **初始化顺序**：数据库初始化在驱动初始化之前执行，确保数据库连接已就绪

4. **幂等性**：建议 `InitializeDatabase()` 实现为幂等的，可以安全地多次调用

## 与 orm 的对比

| 特性 | orm | xdb |
|------|-----|-----|
| 初始化函数 | `orm.MustInitialize(redis, config)` | `xdb.MustInitialize(ctx, configurator)` |
| 数据库初始化 | 在外部调用 `database.MustInitialize` | 在 `Setup` 内部自动调用（可选） |
| 配置接口 | `Config` 接口 | `Configurator` 接口 + `DatabaseConfig` 接口（可选） |

## 迁移指南

如果从 orm 迁移到 xdb：

1. 将 `orm.MustInitialize(redis, config)` 替换为 `xdb.MustInitialize(ctx, configurator)`

2. 将数据库初始化逻辑移到 `InitializeDatabase()` 方法中

3. 实现 `Configurator` 接口的所有方法

4. 可选实现 `DatabaseConfig` 接口

