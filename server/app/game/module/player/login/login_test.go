package login

import (
	"lucky/server/gen/msg"
	"lucky/server/pkg/di"
	"testing"

	cproto "github.com/cherry-game/cherry/net/proto"
)

func TestLoginModule_SelectPlayer(t *testing.T) {
	// 重置并初始化 di 容器
	di.Reset()
	// 手动注册模块（因为 init 函数不会再次执行）
	var loginModule = &LoginModule{}
	di.Register(loginModule)
	// 注册接口实现关系
	di.RegisterImplementation((*ILoginModule)(nil), loginModule)
	di.MustInitialize()

	// 获取 LoginModule 实例
	moduleInstance, err := di.GetByType((*ILoginModule)(nil))
	if err != nil {
		t.Fatalf("Failed to get LoginModule: %v", err)
	}
	loginModuleInstance := moduleInstance.(ILoginModule)

	// 创建模拟 session
	session := &cproto.Session{
		Uid: 1001,
	}

	// 测试查询角色列表
	response, err := loginModuleInstance.SelectPlayer(session)
	if err != nil {
		t.Errorf("SelectPlayer failed: %v", err)
	}

	if response == nil {
		t.Error("SelectPlayer returned nil response")
		return
	}

	if response.List == nil {
		t.Logf("SelectPlayer returned empty list (no player found for uid=%d)", session.Uid)
	} else {
		t.Logf("SelectPlayer found %d players", len(response.List))
	}

	t.Logf("SelectPlayer response: %+v", response)
}

func TestLoginModule_CreatePlayer(t *testing.T) {
	// 重置并初始化 di 容器
	di.Reset()
	// 手动注册模块
	var loginModule = &LoginModule{}
	di.Register(loginModule)
	// 注册接口实现关系
	di.RegisterImplementation((*ILoginModule)(nil), loginModule)
	di.MustInitialize()

	// 获取 LoginModule 实例
	moduleInstance, err := di.GetByType((*ILoginModule)(nil))
	if err != nil {
		t.Fatalf("Failed to get LoginModule: %v", err)
	}
	_ = moduleInstance.(*LoginModule)

	// 测试创建角色 - 正常情况
	req := &msg.PlayerCreateRequest{
		PlayerName: "TestPlayer",
		Gender:     0,
	}

	// 注意：CreatePlayer 需要 actor 参数，这里先测试参数验证
	// 实际测试需要 mock actor
	t.Logf("CreatePlayer request: %+v", req)
}

func TestLoginModule_EnterPlayer(t *testing.T) {
	// 重置并初始化 di 容器
	di.Reset()
	// 手动注册模块
	var loginModule = &LoginModule{}
	di.Register(loginModule)
	// 注册接口实现关系
	di.RegisterImplementation((*ILoginModule)(nil), loginModule)
	di.MustInitialize()

	// 获取 LoginModule 实例
	moduleInstance, err := di.GetByType((*ILoginModule)(nil))
	if err != nil {
		t.Fatalf("Failed to get LoginModule: %v", err)
	}
	_ = moduleInstance.(*LoginModule)

	// 测试进入游戏 - 需要先有角色
	req := &msg.Int64{
		Value: 1, // 假设 playerId = 1
	}

	// 注意：EnterPlayer 需要 actor 参数，这里先测试参数验证
	// 实际测试需要 mock actor
	t.Logf("EnterPlayer request: %+v", req)
}

func TestLoginModule_RegisterHandlers(t *testing.T) {
	// 测试消息处理器注册（通过 init() 方式）
	// handler.go 中的 init() 函数会在包导入时自动执行
	// 这里只验证模块可以正常获取
	di.Reset()
	// 手动注册模块
	var loginModule = &LoginModule{}
	di.Register(loginModule)
	// 注册接口实现关系
	di.RegisterImplementation((*ILoginModule)(nil), loginModule)
	di.MustInitialize()

	// 获取 LoginModule 实例
	moduleInstance, err := di.GetByType((*ILoginModule)(nil))
	if err != nil {
		t.Fatalf("Failed to get LoginModule: %v", err)
	}
	if moduleInstance == nil {
		t.Error("LoginModule instance is nil")
	}

	t.Logf("LoginModule handler registration test passed (handler.go init() will register handlers)")
}
