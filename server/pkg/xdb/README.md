# xdb - ORM 模块

xdb 是一个参考 `zplus-go/orm` 实现的 ORM 模块，提供了类似的功能和接口。

## 核心概念

### Record（记录）
- `Record`: 表示一个数据库记录，包含基本的生命周期管理
- `MutableRecord`: 可变的记录，支持创建、更新、删除操作
- `Model`: 模型，包含锁和监听器功能

### Source（数据源）
每个数据表对应一个 `Source`，包含：
- 表结构信息（字段、主键等）
- 存储驱动配置
- 缓存仓库（Repo）
- 保存器（Saver）

### 主要功能

1. **CRUD 操作**
   - `Create`: 创建新记录
   - `Get`: 获取单个记录
   - `GetAll`: 获取所有匹配的记录
   - `GetMulti`: 批量获取记录
   - `Save`: 保存记录变更
   - `Sync`: 同步等待数据入库

2. **缓存管理**
   - 内存缓存（Repo）
   - 支持易失性数据（Volatile）
   - 前缀查询支持

3. **并发控制**
   - 读写锁支持
   - 批量锁操作
   - 锁优先级管理

4. **数据持久化**
   - 异步批量保存
   - 重做日志支持
   - 恢复机制

## 使用示例

### 1. 定义数据源

```go
// 定义 Proto 类型
type PlayerProto struct {
    PlayerId int64
    Name     string
    Level    int32
}

// 定义 Record 类型
type PlayerRecord struct {
    Header
    PlayerProto
}

// 实现 Record 接口
func (r *PlayerRecord) Source() *Source {
    return GetSourceByNS("player")
}

func (r *PlayerRecord) XId() string {
    return fmt.Sprintf("player:%d", r.PlayerId)
}

func (r *PlayerRecord) Lifecycle() Lifecycle {
    return r.Header.Lifecycle()
}

func (r *PlayerRecord) Snapshoot() interface{} {
    return &r.PlayerProto
}

func (r *PlayerRecord) XVersion() int64 {
    return 0 // 实现版本号逻辑
}

// 注册数据源
func init() {
    src := &Source{
        ProtoType:  reflect.TypeOf((*PlayerProto)(nil)).Elem(),
        RecordType: reflect.TypeOf((*PlayerRecord)(nil)).Elem(),
        Namespace:  "player",
        DriverName: "mysql",
        TableName:  "player",
        KeySize:    1,
        PKCreator: func(args []interface{}) (PK, error) {
            // 创建主键
            return &PlayerPK{PlayerId: args[0].(int64)}, nil
        },
        PKOf: func(obj interface{}) PK {
            r := obj.(*PlayerRecord)
            return &PlayerPK{PlayerId: r.PlayerId}
        },
        // ... 其他配置
    }
    RegisterSource(src)
}
```

### 2. 初始化 xdb

```go
type MyConfigurator struct{}

func (c *MyConfigurator) RedoOptions() *RedoOptions {
    return &RedoOptions{
        Dir:     "./redo",
        Enabled: true,
    }
}

func (c *MyConfigurator) DriverOptions(driver string) interface{} {
    // 返回驱动配置
    return &MySQLOptions{
        Host: "localhost",
        Port: 3306,
        // ...
    }
}

func (c *MyConfigurator) DaoOptions(daoKey interface{}) interface{} {
    // 返回 DAO 配置
    return &DaoOptions{
        // ...
    }
}

func (c *MyConfigurator) TableOptions(driver string, table string) *TableOptions {
    return &TableOptions{
        DaoKey:      "mysql",
        Concurrence: 4,
        SaveTimeout: 5 * time.Second,
        SyncInterval: 100 * time.Millisecond,
    }
}

func (c *MyConfigurator) DryRun() bool {
    return false
}

// 在应用启动时调用
func main() {
    ctx := context.Background()
    configurator := &MyConfigurator{}
    err := Setup(ctx, configurator)
    if err != nil {
        panic(err)
    }
}
```

### 3. 使用 CRUD 操作

```go
// 创建记录
player, err := Create[PlayerRecord](ctx, &PlayerProto{
    PlayerId: 1001,
    Name:     "TestPlayer",
    Level:    1,
})

// 获取记录
player, err := Get[PlayerRecord](ctx, int64(1001))

// 更新记录
player.Name = "NewName"
player.GetHeader().SetChanged(FieldName) // 标记字段变更
Save(ctx, player)

// 删除记录
player.Delete(ctx)
Save(ctx, player)

// 同步保存
err := Sync(ctx, player)
```

## 架构说明

### 模块结构

- `record.go`: Record、Model 等核心接口定义
- `header.go`: Header 和 FieldSet 实现
- `src.go`: Source 注册和管理
- `xdb.go`: 主要的 CRUD 操作
- `storage.go`: 存储驱动接口
- `saver.go`: 异步保存器
- `repo.go`: 内存缓存仓库
- `lock.go`: 锁机制
- `context.go`: 上下文管理
- `commitment.go`: 提交对象定义

### 与 zplus-go/orm 的差异

1. **依赖简化**: 移除了对 `zplus-go` 特定包的依赖
2. **接口适配**: 适配了 `lucky/server` 项目的依赖结构
3. **功能精简**: 保留了核心功能，移除了一些高级特性

## 扩展开发

### 实现自定义驱动

```go
type MyDriver struct{}

func (d *MyDriver) Name() string {
    return "mydriver"
}

func (d *MyDriver) Init(ctx context.Context, config interface{}) error {
    // 初始化驱动
    return nil
}

func (d *MyDriver) Validate(src *Source) error {
    // 验证数据源
    return nil
}

func (d *MyDriver) NewDao(ctx context.Context, option interface{}) (Dao, error) {
    // 创建 DAO
    return &MyDao{}, nil
}

func (d *MyDriver) ExtendType(base reflect.Type, extension reflect.Type) {
    // 扩展类型
}

// 注册驱动
func init() {
    RegisterDriver(&MyDriver{})
}
```

## 注意事项

1. **线程安全**: 所有操作都是线程安全的，但需要正确使用锁
2. **内存管理**: 注意缓存大小，避免内存泄漏
3. **错误处理**: 所有操作都可能返回错误，需要妥善处理
4. **性能优化**: 批量操作比单个操作更高效



