package di_test

import (
	"fmt"
	"lucky/server/pkg/di"
	"testing"
)

// 示例：定义接口和实现

type IUserService interface {
	GetUser(id int64) string
}

type UserService struct {
	ID int64
}

func (s *UserService) GetUser(id int64) string {
	return fmt.Sprintf("user_%d", id)
}

type IOrderService interface {
	CreateOrder(userId int64) string
}

type OrderService struct {
	UserService IUserService `inject:"auto"` // 自动注入
}

func (s *OrderService) CreateOrder(userId int64) string {
	user := s.UserService.GetUser(userId)
	return fmt.Sprintf("order_for_%s", user)
}

func TestIOC_RegisterAndGet(t *testing.T) {
	// 清空容器
	di.Clear()

	// 注册服务实例（单例）
	userService := &UserService{ID: 1}
	di.RegisterWithName("userService", userService, true)

	// 获取服务
	service, err := di.Get("userService")
	if err != nil {
		t.Fatalf("Get service failed: %v", err)
	}

	userSvc, ok := service.(*UserService)
	if !ok {
		t.Fatalf("Type assertion failed")
	}

	if userSvc.ID != 1 {
		t.Fatalf("Expected ID 1, got %d", userSvc.ID)
	}
}

func TestIOC_RegisterWithOption(t *testing.T) {
	// 清空容器
	di.Clear()

	// 使用 Option 模式注册（类似 claim）
	userService := &UserService{ID: 2}
	di.Register(userService, di.WithName("userService"))

	// 获取服务
	service, err := di.Get("userService")
	if err != nil {
		t.Fatalf("Get service failed: %v", err)
	}

	userSvc := service.(*UserService)
	if userSvc.ID != 2 {
		t.Fatalf("Expected ID 2, got %d", userSvc.ID)
	}
}

func TestIOC_RegisterWithReplace(t *testing.T) {
	// 清空容器
	di.Clear()

	// 第一次注册
	userService1 := &UserService{ID: 1}
	di.Register(userService1, di.WithName("userService"))

	// 替换注册
	userService2 := &UserService{ID: 2}
	di.Register(userService2, di.WithName("userService"), di.WithReplace(true))

	// 获取服务，应该是新的实例
	service, err := di.Get("userService")
	if err != nil {
		t.Fatalf("Get service failed: %v", err)
	}

	userSvc := service.(*UserService)
	if userSvc.ID != 2 {
		t.Fatalf("Expected ID 2 after replace, got %d", userSvc.ID)
	}
}

func TestIOC_RegisterFactory(t *testing.T) {
	// 清空容器
	di.Clear()

	// 注册工厂函数
	di.RegisterFactory("userService", func() any {
		return &UserService{ID: 2}
	}, true)

	// 获取服务
	service, err := di.Get("userService")
	if err != nil {
		t.Fatalf("Get service failed: %v", err)
	}

	userSvc := service.(*UserService)
	if userSvc.ID != 2 {
		t.Fatalf("Expected ID 2, got %d", userSvc.ID)
	}
}

func TestIOC_Resolve(t *testing.T) {
	// 清空容器
	di.Clear()

	// 注册服务实例（直接注册，不需要先注册接口）
	userService := &UserService{ID: 3}
	di.Register(userService, di.WithName("userService"))

	// 创建需要注入的对象
	orderService := &OrderService{}

	// 解析依赖（自动根据类型查找）
	err := di.Resolve(orderService)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// 验证注入
	if orderService.UserService == nil {
		t.Fatalf("UserService not injected")
	}

	result := orderService.CreateOrder(123)
	if result != "order_for_user_123" {
		t.Fatalf("Expected 'order_for_user_123', got '%s'", result)
	}
}

func TestIOC_GetByType(t *testing.T) {
	// 清空容器
	di.Clear()

	// 注册服务实例
	userService := &UserService{ID: 4}
	di.Register(userService, di.WithName("userService"))

	// 根据类型获取（直接根据接口类型查找）
	service, err := di.GetByType((*IUserService)(nil))
	if err != nil {
		t.Fatalf("GetByType failed: %v", err)
	}

	userSvc, ok := service.(IUserService)
	if !ok {
		t.Fatalf("Type assertion failed, got type: %T", service)
	}

	result := userSvc.GetUser(456)
	if result != "user_456" {
		t.Fatalf("Expected 'user_456', got '%s'", result)
	}
}
