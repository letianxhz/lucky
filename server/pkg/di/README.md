# DI 容器模块

基于依赖注入的 DI（Dependency Injection）容器实现，使用 `syncmap` 实现线程安全，无需显式加锁。

## 功能特性

1. **服务注册**: 支持注册服务实例和工厂函数
2. **单例模式**: 支持单例和原型模式
3. **接口注册**: 支持接口类型注册和实现映射
4. **依赖注入**: 通过结构体标签自动注入依赖
5. **类型安全**: 编译期类型检查
6. **线程安全**: 使用 `sync.Map` 保证并发安全，无需显式加锁

## 使用示例

### 1. 基本使用

```go
import "lucky/server/pkg/di"

// 注册服务实例（单例）
userService := &UserService{}
di.Register(userService, di.WithName("userService"))

// 获取服务
service, err := di.Get("userService")
if err != nil {
    // 处理错误
}
userSvc := service.(*UserService)
```

### 2. 使用工厂函数

```go
// 注册工厂函数
di.RegisterFactory("userService", func() interface{} {
    return &UserService{ID: 1}
}, true)

// 获取服务（单例模式下，第一次调用时创建）
service, err := di.Get("userService")
```

### 3. 接口注册和实现

```go
// 定义接口
type IUserService interface {
    GetUser(id int64) string
}

// 注册接口（可选）
di.RegisterInterface("IUserService", (*IUserService)(nil))

// 注册实现（可选）
di.RegisterImplementation((*IUserService)(nil), (*UserService)(nil))

// 注册服务实例
userService := &UserService{}
di.Register("userService", userService, true)

// 根据类型获取（自动查找实现该接口的实例）
service, err := di.GetByType((*IUserService)(nil))
userSvc := service.(IUserService)
```

### 4. 依赖注入

```go
// 定义需要注入的结构体
type OrderService struct {
    UserService IUserService `inject:"auto"` // 自动注入
    // 或者指定服务名称
    // UserService IUserService `inject:"userService"`
    // 兼容 claim 的 ioc 标签
    // UserService IUserService `ioc:"auto"`
}

// 注册依赖的服务
userService := &UserService{}
di.Register("userService", userService, true)

// 创建对象并解析依赖
orderService := &OrderService{}
err := di.Resolve(orderService)
if err != nil {
    // 处理错误
}

// 现在 orderService.UserService 已经被自动注入
```

### 5. 在模块中使用

```go
// module/item/item_impl.go
type ItemModule struct {
    // 通过依赖注入获取其他模块
    EquipmentModule IEquipmentModule `inject:"auto"`
}

// 初始化时注册
func InitModules() {
    // 注册服务
    itemModule := &ItemModule{}
    di.Register("itemModule", itemModule, true)
    
    // 解析依赖
    di.Resolve(itemModule)
}
```

## API 文档

### 容器操作

- `GetContainer()`: 获取默认容器
- `NewContainer()`: 创建新容器
- `Clear()`: 清空容器
- `Has(name string)`: 检查服务是否存在

### 服务注册

- `Register(name, instance, singleton)`: 注册服务实例
- `RegisterFactory(name, factory, singleton)`: 注册工厂函数
- `RegisterInterface(name, iface)`: 注册接口类型
- `RegisterImplementation(iface, impl)`: 注册接口实现

### 服务获取

- `Get(name)`: 根据名称获取服务
- `GetByType(iface)`: 根据类型获取服务
- `Resolve(target)`: 解析依赖并注入

## 标签说明

### `inject` 标签（推荐）

- `inject:"auto"` 或 `inject:""`: 根据字段类型自动查找并注入
- `inject:"serviceName"`: 根据服务名称注入

### `ioc` 标签（兼容 claim）

- `ioc:"auto"` 或 `ioc:""`: 根据字段类型自动查找并注入
- `ioc:"serviceName"`: 根据服务名称注入

## 线程安全

使用 `sync.Map` 实现线程安全，所有操作都是并发安全的：

- `Register`: 并发安全
- `Get`: 并发安全
- `GetByType`: 并发安全
- `Resolve`: 并发安全

## 最佳实践

1. **接口优先**: 使用接口定义依赖，提高可测试性
2. **单例模式**: 对于无状态服务，使用单例模式
3. **工厂函数**: 对于需要复杂初始化的服务，使用工厂函数
4. **依赖注入**: 在结构体中使用 `inject` 标签，避免手动获取依赖
5. **错误处理**: 始终检查 `Get` 和 `Resolve` 的返回值

## 注意事项

1. **线程安全**: 容器是线程安全的，可以在并发环境下使用
2. **类型检查**: `Resolve` 会进行类型检查，类型不匹配会返回错误
3. **循环依赖**: 容器不检测循环依赖，需要开发者避免
4. **性能**: 反射操作有一定性能开销，适合在初始化阶段使用
5. **sync.Map**: 使用 `sync.Map` 替代 `sync.RWMutex`，性能更好，代码更简洁

## 与模块管理器的集成

可以结合模块管理器使用：

```go
// module/manager.go
func InitModules() {
    // 使用 IOC 容器管理模块
    itemModule := item.NewItemModule()
    di.Register("itemModule", itemModule, true)
    
    equipmentModule := equipment.NewEquipmentModule(itemModule)
    di.Register("equipmentModule", equipmentModule, true)
    
    // 或者使用依赖注入
    equipmentModule2 := &equipment.EquipmentModule{}
    di.Resolve(equipmentModule2)
    di.Register("equipmentModule", equipmentModule2, true)
}
```

## 性能优化

- 使用 `sync.Map` 替代 `sync.RWMutex`，减少锁竞争
- 类型信息缓存，避免重复反射
- 单例模式减少对象创建
