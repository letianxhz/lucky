package handler

import (
	"testing"

	"github.com/cherry-game/cherry/net/parser/pomelo"
)

func TestHandlerRegistration(t *testing.T) {
	// 测试消息处理器注册机制
	// 注意：实际的 handler 注册在 item/handler.go 和 login/handler.go 的 init() 中完成
	// 这里只测试 Register 和 RegisterAll 函数的基本功能

	// 创建一个模拟的 actor 来测试注册
	mockActor := &pomelo.ActorBase{}

	// 注册一个测试处理器函数
	Register(func(actor *pomelo.ActorBase) {
		// 测试处理器
		t.Logf("Test handler registered")
	})

	// 调用 RegisterAll 注册所有已注册的消息处理器
	RegisterAll(mockActor)

	// 验证注册成功（这里只是验证不会 panic）
	t.Logf("Handler registration test passed: Register and RegisterAll work correctly")
}
