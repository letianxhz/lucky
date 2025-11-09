package di

import (
	"fmt"
	"reflect"
)

// 全局便捷函数，使用默认容器

// Register 注册服务实例（支持 Option 模式，类似 claim）
// v: 服务实例
// opts: 选项（名称、是否替换等）
func Register(v any, opts ...Option) {
	GetContainer().Register(v, opts...)
}

// RegisterWithName 注册服务实例（按名称，兼容旧接口）
// name: 服务名称
// instance: 服务实例
// singleton: 是否为单例
func RegisterWithName(name string, instance any, singleton bool) {
	GetContainer().RegisterWithName(name, instance, singleton)
}

// RegisterFactory 注册服务工厂
func RegisterFactory(name string, factory func() any, singleton bool) {
	GetContainer().RegisterFactory(name, factory, singleton)
}

// RegisterInterface 注册接口类型
func RegisterInterface(name string, iface any) {
	ifaceType := reflect.TypeOf(iface)
	if ifaceType == nil {
		panic("RegisterInterface: iface cannot be nil")
	}
	if ifaceType.Kind() == reflect.Ptr {
		ifaceType = ifaceType.Elem()
	}
	GetContainer().RegisterInterface(name, ifaceType)
}

// RegisterImplementation 注册接口实现
func RegisterImplementation(iface any, impl any) {
	ifaceType := reflect.TypeOf(iface)
	if ifaceType == nil {
		panic("RegisterImplementation: iface cannot be nil")
	}
	if ifaceType.Kind() == reflect.Ptr {
		ifaceType = ifaceType.Elem()
	}

	implType := reflect.TypeOf(impl)
	if implType == nil {
		panic("RegisterImplementation: impl cannot be nil")
	}
	if implType.Kind() == reflect.Ptr {
		implType = implType.Elem()
	}
	GetContainer().RegisterImplementation(ifaceType, implType)
}

// Get 获取服务实例
func Get(name string) (any, error) {
	return GetContainer().Get(name)
}

// GetByType 根据类型获取服务实例
func GetByType(iface any) (any, error) {
	ifaceType := reflect.TypeOf(iface)
	if ifaceType == nil {
		return nil, fmt.Errorf("GetByType: iface cannot be nil")
	}
	if ifaceType.Kind() == reflect.Ptr {
		ifaceType = ifaceType.Elem()
	}
	return GetContainer().GetByType(ifaceType)
}

// Resolve 解析依赖并注入
func Resolve(target any) error {
	return GetContainer().Resolve(target)
}

// Has 检查服务是否存在
func Has(name string) bool {
	return GetContainer().Has(name)
}

// Clear 清空容器
func Clear() {
	GetContainer().Clear()
}

// Reset 重置容器（用于测试）
// 注意：这个函数会重置容器状态，主要用于测试
func Reset() {
	// 通过 GetContainer 获取容器并清空，然后重置状态
	container := GetContainer()
	container.Clear()
	// Clear 已经将状态重置为 StateInitializing
}

// MustInitialize 初始化容器，为所有已注册的组件注入依赖
// 参考 claim ioc 的实现，在服务启动时调用
func MustInitialize() {
	GetContainer().initialize()
}
