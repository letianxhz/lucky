package di

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	clog "github.com/cherry-game/cherry/logger"
	"github.com/sandwich-go/boost/xcontainer/syncmap"
)

// State 容器状态
type State int8

const (
	StateInitializing State = iota
	StateInitialized
)

// Container IOC 容器
type Container struct {
	state           State                                     // 容器状态
	namedInstances  syncmap.Map[string, any]                  // 服务实例缓存（按名称）
	typedInstances  syncmap.Map[reflect.Type, map[any]bool]   // 服务实例缓存（按类型），类似 claim 的 map[reflect.Type]map[any]bool
	factories       syncmap.Map[string, func() any]           // 服务工厂函数
	singletons      syncmap.Map[string, bool]                 // 单例标记
	interfaces      syncmap.Map[string, reflect.Type]         // 接口类型映射
	implementations syncmap.Map[reflect.Type, []reflect.Type] // 接口到实现的映射
}

var (
	defaultContainer *Container
	once             sync.Once
)

var container = Container{}

func init() {
	reset()
}

func reset() {
	container.state = StateInitializing
	container.namedInstances = syncmap.Map[string, any]{}
	container.typedInstances = syncmap.Map[reflect.Type, map[any]bool]{}
	container.factories = syncmap.Map[string, func() any]{}
	container.singletons = syncmap.Map[string, bool]{}
	container.interfaces = syncmap.Map[string, reflect.Type]{}
	container.implementations = syncmap.Map[reflect.Type, []reflect.Type]{}
}

// GetContainer 获取默认容器
func GetContainer() *Container {
	once.Do(func() {
		defaultContainer = NewContainer()
	})
	return defaultContainer
}

// NewContainer 创建新的容器
func NewContainer() *Container {
	return &Container{}
}

// register 内部注册方法（支持 Option，类似 claim 的实现）
func (c *Container) register(v any, opt *Option) {
	if c.state != StateInitializing {
		panic("can only register during initialization phase")
	}

	instanceType := reflect.TypeOf(v)
	if instanceType == nil {
		panic("cannot register nil value")
	}

	// 按类型注册（类似 claim 的 typedInstances map[reflect.Type]map[any]bool）
	instances, _ := c.typedInstances.LoadOrStore(instanceType, make(map[any]bool))
	instancesMap := instances
	// 检查是否已存在
	if instancesMap[v] {
		return // 已存在，直接返回
	}
	instancesMap[v] = true

	// 注意：接口实现关系不需要显式注册
	// GetByType 会通过遍历 typedInstances 自动查找实现接口的类型
	// 这样更灵活，不需要预先注册接口类型

	// 如果指定了名称，按名称注册
	if opt != nil && opt.Name != "" {
		// 检查是否已存在
		if orig, ok := c.namedInstances.Load(opt.Name); ok {
			if !opt.Replace {
				panic(fmt.Sprintf("service with name '%s' already exists (use WithReplace(true) to replace)", opt.Name))
			}
			// 替换时，从旧类型中删除
			origType := reflect.TypeOf(orig)
			if origInstances, ok := c.typedInstances.Load(origType); ok {
				delete(origInstances, orig)
			}
		}
		c.namedInstances.Store(opt.Name, v)
	}
}

// Register 注册服务实例（支持 Option 模式）
// v: 服务实例
// opts: 选项（名称、是否替换等）
func (c *Container) Register(v any, opts ...Option) {
	opt := NewOptions(opts...)
	c.register(v, opt)
}

// RegisterWithName 注册服务实例（按名称，兼容旧接口）
// name: 服务名称
// instance: 服务实例
// singleton: 是否为单例
func (c *Container) RegisterWithName(name string, instance any, singleton bool) {
	c.namedInstances.Store(name, instance)
	c.singletons.Store(name, singleton)

	// 同时按类型注册
	instanceType := reflect.TypeOf(instance)
	if instanceType != nil {
		// 获取或创建该类型的实例集合
		instances, _ := c.typedInstances.LoadOrStore(instanceType, make(map[any]bool))
		instancesMap := instances
		instancesMap[instance] = true
	}
}

// RegisterFactory 注册服务工厂
// name: 服务名称
// factory: 工厂函数
// singleton: 是否为单例
func (c *Container) RegisterFactory(name string, factory func() any, singleton bool) {
	c.factories.Store(name, factory)
	c.singletons.Store(name, singleton)
}

// RegisterInterface 注册接口类型
// name: 服务名称
// iface: 接口类型
func (c *Container) RegisterInterface(name string, iface reflect.Type) {
	c.interfaces.Store(name, iface)
}

// RegisterImplementation 注册接口实现
// iface: 接口类型
// impl: 实现类型
func (c *Container) RegisterImplementation(iface reflect.Type, impl reflect.Type) {
	impls, _ := c.implementations.LoadOrStore(iface, []reflect.Type{})
	implsSlice := impls
	implsSlice = append(implsSlice, impl)
	c.implementations.Store(iface, implsSlice)
}

// Get 获取服务实例
// name: 服务名称
func (c *Container) Get(name string) (any, error) {
	// 如果是单例且已存在实例，直接返回
	if singleton, ok := c.singletons.Load(name); ok && singleton {
		if instance, exists := c.namedInstances.Load(name); exists {
			return instance, nil
		}
	}

	// 如果有工厂函数，使用工厂创建
	if factory, ok := c.factories.Load(name); ok {
		factoryFunc := factory
		instance := factoryFunc()
		if singleton, ok := c.singletons.Load(name); ok && singleton {
			c.namedInstances.Store(name, instance)
		}
		return instance, nil
	}

	// 直接返回已注册的实例
	if instance, ok := c.namedInstances.Load(name); ok {
		return instance, nil
	}

	return nil, fmt.Errorf("service not found: %s", name)
}

// GetByType 根据类型获取服务实例
// iface: 接口类型或具体类型
func (c *Container) GetByType(iface reflect.Type) (any, error) {
	// 如果是接口类型，查找实现
	if iface.Kind() == reflect.Interface {
		// 首先从已注册的服务中查找实现该接口的实例
		// 参考 claim 的实现：遍历所有 typedInstances，查找实现接口的类型
		var foundInstance any
		c.typedInstances.Range(func(serviceType reflect.Type, instancesMap map[any]bool) bool {
			// 检查该类型是否实现了接口
			if serviceType.Implements(iface) {
				// 从 map 中取第一个实例
				for instance := range instancesMap {
					foundInstance = instance
					clog.Debugf("[DI] Found implementation %T for interface %s", instance, iface.String())
					break
				}
				if foundInstance != nil {
					return false // 停止遍历
				}
			}
			return true
		})
		if foundInstance != nil {
			return foundInstance, nil
		}

		// 如果没找到，查找接口的实现类型
		if implsValue, ok := c.implementations.Load(iface); ok {
			impls := implsValue
			if len(impls) > 0 {
				impl := impls[0]
				// 查找该类型的已注册实例
				if instancesMap, ok := c.typedInstances.Load(impl); ok {
					// 从 map 中取第一个实例
					for instance := range instancesMap {
						return instance, nil
					}
				}
				// 如果没有已注册实例，创建新实例
				if impl.Kind() == reflect.Ptr {
					return reflect.New(impl.Elem()).Interface(), nil
				}
				return reflect.New(impl).Interface(), nil
			}
		}

		// 如果没找到，记录错误信息（包含已注册的类型信息，便于调试）
		var registeredTypes []string
		c.typedInstances.Range(func(serviceType reflect.Type, instancesMap map[any]bool) bool {
			registeredTypes = append(registeredTypes, serviceType.String())
			return true
		})
		clog.Warnf("[DI] No implementation found for interface %s. Registered types: %v", iface.String(), registeredTypes)
		return nil, fmt.Errorf("no implementation found for interface: %s", iface.String())
	}

	// 如果是具体类型，直接查找
	if instancesMap, ok := c.typedInstances.Load(iface); ok {
		// 从 map 中取第一个实例
		for instance := range instancesMap {
			return instance, nil
		}
	}

	return nil, fmt.Errorf("no instance found for type: %s", iface.String())
}

// Resolve 解析依赖并注入
// target: 目标对象（必须是指针）
// 支持两种标签：
//   - inject:"auto" 或 inject:"" - 根据字段类型自动注入
//   - inject:"serviceName" - 根据服务名称注入
//     兼容 claim 的 ioc 标签
func (c *Container) Resolve(target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetType := targetValue.Elem().Type()
	if targetType.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	targetValue = targetValue.Elem()

	// 遍历结构体字段
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetValue.Field(i)

		// 检查是否有 di、inject 或 ioc 标签（兼容 claim 的 ioc 标签）
		// 优先级：di > inject > ioc
		tag := field.Tag.Get("di")
		if tag == "" {
			tag = field.Tag.Get("inject")
		}
		if tag == "" {
			tag = field.Tag.Get("ioc")
		}
		if tag == "" || tag == "-" {
			// 如果是匿名结构体，递归处理
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				embeddedValue := targetValue.Field(i)
				if embeddedValue.CanAddr() {
					if err := c.Resolve(embeddedValue.Addr().Interface()); err != nil {
						return err
					}
				}
			}
			continue
		}

		// 获取服务实例
		var instance any
		var err error

		if tag == "auto" || tag == "" {
			// 自动注入：根据字段类型查找
			instance, err = c.GetByType(field.Type)
		} else {
			// 按名称注入
			instance, err = c.Get(tag)
		}

		if err != nil {
			return fmt.Errorf("failed to resolve dependency for field %s.%s: %v", targetType.Name(), field.Name, err)
		}

		// 设置字段值
		// 如果字段不可设置（未导出），使用 unsafe 包来设置
		instanceValue := reflect.ValueOf(instance)
		if !fieldValue.CanSet() {
			// 字段未导出，使用 unsafe 包来设置
			// 参考 claim 的实现：通过 unsafe.Pointer 来设置未导出的字段
			if !fieldValue.CanAddr() {
				return fmt.Errorf("field %s.%s cannot be set (not addressable)", targetType.Name(), field.Name)
			}
			// 获取字段的地址，然后通过 unsafe.Pointer 设置
			fieldPtr := unsafe.Pointer(fieldValue.UnsafeAddr())
			fieldValuePtr := reflect.NewAt(field.Type, fieldPtr).Elem()
			if instanceValue.Type().AssignableTo(field.Type) {
				fieldValuePtr.Set(instanceValue)
			} else if instanceValue.Type().ConvertibleTo(field.Type) {
				fieldValuePtr.Set(instanceValue.Convert(field.Type))
			} else {
				return fmt.Errorf("type mismatch for field %s.%s: expected %s, got %s",
					targetType.Name(), field.Name, field.Type, instanceValue.Type())
			}
		} else {
			// 字段可设置（已导出），直接设置
			if instanceValue.Type().AssignableTo(field.Type) {
				fieldValue.Set(instanceValue)
			} else if instanceValue.Type().ConvertibleTo(field.Type) {
				fieldValue.Set(instanceValue.Convert(field.Type))
			} else {
				return fmt.Errorf("type mismatch for field %s.%s: expected %s, got %s",
					targetType.Name(), field.Name, field.Type, instanceValue.Type())
			}
		}
	}

	return nil
}

// Clear 清空容器并重置状态
func (c *Container) Clear() {
	c.namedInstances = syncmap.Map[string, any]{}
	c.typedInstances = syncmap.Map[reflect.Type, map[any]bool]{}
	c.factories = syncmap.Map[string, func() any]{}
	c.singletons = syncmap.Map[string, bool]{}
	c.interfaces = syncmap.Map[string, reflect.Type]{}
	c.implementations = syncmap.Map[reflect.Type, []reflect.Type]{}
	// 重置状态为初始化中，允许重新注册
	c.state = StateInitializing
}

// Has 检查服务是否存在
func (c *Container) Has(name string) bool {
	_, hasService := c.namedInstances.Load(name)
	_, hasFactory := c.factories.Load(name)
	return hasService || hasFactory
}

// GetTypedInstances 获取按类型注册的实例映射（用于遍历所有已注册的实例）
func (c *Container) GetTypedInstances() *syncmap.Map[reflect.Type, map[any]bool] {
	return &c.typedInstances
}

// initialize 初始化容器，为所有已注册的组件注入依赖
// 参考 claim ioc 的实现：遍历所有已注册的实例，为每个实例注入依赖
// 注意：在注入时，所有实例都已经注册完成，可以安全地查找依赖
func (c *Container) initialize() {
	if c.state != StateInitializing {
		panic("can only initialize during initialization phase")
	}

	// 统计需要注入的实例数量
	totalCount := 0
	injectCount := 0

	// 遍历所有已注册的实例，为每个实例注入依赖
	// 注意：instance 必须是指针类型，因为 Resolve 需要指针
	c.typedInstances.Range(func(instanceType reflect.Type, instancesMap map[any]bool) bool {
		for instance := range instancesMap {
			totalCount++
			// 检查实例类型是否有需要注入的字段（有 di、inject 或 ioc 标签）
			if c.hasInjectFields(instanceType) {
				// instance 应该已经是指针（注册时传入的，如 &ItemModule{}）
				// 但为了安全，检查一下
				instanceValue := reflect.ValueOf(instance)
				if instanceValue.Kind() != reflect.Ptr {
					// 如果不是指针，尝试获取指针
					if instanceValue.CanAddr() {
						instance = instanceValue.Addr().Interface()
					} else {
						panic(fmt.Sprintf("instance %T must be a pointer for dependency injection", instance))
					}
				}
				// 注入依赖（此时所有实例都已注册，可以安全查找依赖）
				clog.Debugf("[DI] Injecting dependencies for %T", instance)
				if err := c.Resolve(instance); err != nil {
					panic(fmt.Sprintf("failed to inject dependencies for %T: %v", instance, err))
				}
				injectCount++
			}
		}
		return true
	})

	clog.Infof("[DI] Initialized container: total instances=%d, injected=%d", totalCount, injectCount)
	c.state = StateInitialized
}

// hasInjectFields 检查类型是否有需要注入的字段
func (c *Container) hasInjectFields(typ reflect.Type) bool {
	if typ.Kind() != reflect.Ptr {
		return false
	}

	elemType := typ.Elem()
	if elemType.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		tag := field.Tag.Get("di")
		if tag == "" {
			tag = field.Tag.Get("inject")
		}
		if tag == "" {
			tag = field.Tag.Get("ioc")
		}
		if tag != "" && tag != "-" {
			return true
		}
	}

	return false
}
