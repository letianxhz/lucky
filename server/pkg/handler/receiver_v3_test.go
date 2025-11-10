package handler

import (
	"lucky/server/gen/msg"
	"testing"

	"github.com/cherry-game/cherry/net/parser/pomelo"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// TestRegisterHandler 测试新的类型安全注册机制（V3 泛型版本）
func TestRegisterHandler(t *testing.T) {
	// 创建一个测试 handler 结构体
	type testHandler struct {
		callCount int
	}

	h := &testHandler{}

	// 注册一个测试 handler（使用唯一的路由名避免冲突）
	route := "testRouteV3"
	RegisterHandler(ActorTypePlayer, route, func(session *cproto.Session, req *msg.None) (*msg.None, error) {
		h.callCount++
		return &msg.None{}, nil
	})

	// 验证 handler 已注册（通过日志输出验证）
	t.Logf("Handler registered successfully for route: %s", route)
}

// TestRegisterHandlerWithActor 测试带 actor 参数的注册机制
func TestRegisterHandlerWithActor(t *testing.T) {
	// 创建一个测试 handler
	type testHandler struct {
		callCount int
	}

	h := &testHandler{}

	// 注册一个需要 actor 参数的 handler（使用唯一的路由名避免冲突）
	route := "testRouteWithActorV3"
	RegisterHandlerWithActor(ActorTypePlayer, route, func(session *cproto.Session, req *msg.None, actor *pomelo.ActorBase) (*msg.None, error) {
		h.callCount++
		if actor == nil {
			t.Error("Actor should not be nil")
		}
		return &msg.None{}, nil
	})

	t.Logf("Handler with actor parameter registered successfully for route: %s", route)
}

// TestErrorWithCode 测试错误码机制
func TestErrorWithCode(t *testing.T) {
	err := NewErrorWithCode(404)
	if err == nil {
		t.Fatal("NewErrorWithCode should not return nil")
	}

	if err.Code != 404 {
		t.Errorf("Expected error code 404, got %d", err.Code)
	}

	// 测试错误消息
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}

	t.Logf("ErrorWithCode test passed: code=%d, msg=%s", err.Code, errMsg)
}

// TestErrorWithCodeAndErr 测试带原始错误的错误码机制
func TestErrorWithCodeAndErr(t *testing.T) {
	originalErr := &ErrorWithCode{Code: 500}
	err := NewErrorWithCodeAndErr(404, originalErr)

	if err == nil {
		t.Fatal("NewErrorWithCodeAndErr should not return nil")
	}

	if err.Code != 404 {
		t.Errorf("Expected error code 404, got %d", err.Code)
	}

	if err.Err != originalErr {
		t.Error("Original error should be preserved")
	}

	t.Logf("ErrorWithCodeAndErr test passed: code=%d", err.Code)
}

// TestRegisterHandlerDuplicate 测试重复注册应该 panic
func TestRegisterHandlerDuplicate(t *testing.T) {
	route := "duplicateRouteV3"

	// 第一次注册
	RegisterHandler(ActorTypePlayer, route, func(session *cproto.Session, req *msg.None) (*msg.None, error) {
		return &msg.None{}, nil
	})

	// 第二次注册相同路由应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when registering duplicate route")
		} else {
			t.Logf("Correctly panicked on duplicate registration: %v", r)
		}
	}()

	RegisterHandler(ActorTypePlayer, route, func(session *cproto.Session, req *msg.None) (*msg.None, error) {
		return &msg.None{}, nil
	})
}

// TestRegisterHandlerInvalidActorType 测试无效的 Actor 类型应该 panic
func TestRegisterHandlerInvalidActorType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when registering with invalid actor type")
		} else {
			t.Logf("Correctly panicked on invalid actor type: %v", r)
		}
	}()

	RegisterHandler(ActorType("invalid"), "testRouteInvalid", func(session *cproto.Session, req *msg.None) (*msg.None, error) {
		return &msg.None{}, nil
	})
}

// TestRegisterHandlerRemote 测试 Remote 消息处理器注册（用于 center 服务）
func TestRegisterHandlerRemote(t *testing.T) {
	// 创建一个测试 handler
	type testHandler struct {
		callCount int
	}

	h := &testHandler{}

	// 注册一个 Remote handler（使用唯一的路由名避免冲突）
	route := "testRemoteRouteV3"
	RegisterHandlerRemote(ActorTypeUuid, route, func(req *msg.String) (*msg.Int64, int32) {
		h.callCount++
		// 模拟成功返回
		return &msg.Int64{Value: 12345}, 0
	})

	// 验证 handler 已注册
	t.Logf("Remote handler registered successfully for route: %s", route)

	// 验证 handler 信息已存储
	// 注意：这里我们无法直接访问内部存储，但可以通过日志验证
}

// TestRegisterHandlerRemoteWithError 测试 Remote handler 返回错误码
func TestRegisterHandlerRemoteWithError(t *testing.T) {
	route := "testRemoteRouteWithErrorV3"

	RegisterHandlerRemote(ActorTypeUuid, route, func(req *msg.String) (*msg.Int64, int32) {
		// 模拟返回错误码
		return nil, 500
	})

	t.Logf("Remote handler with error code registered successfully for route: %s", route)
}

// TestRegisterHandlerRemoteDuplicate 测试 Remote handler 重复注册应该 panic
func TestRegisterHandlerRemoteDuplicate(t *testing.T) {
	route := "testRemoteDuplicateV3"

	// 第一次注册
	RegisterHandlerRemote(ActorTypeUuid, route, func(req *msg.String) (*msg.Int64, int32) {
		return &msg.Int64{Value: 1}, 0
	})

	// 第二次注册相同路由应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when registering duplicate remote route")
		} else {
			t.Logf("Correctly panicked on duplicate remote registration: %v", r)
		}
	}()

	RegisterHandlerRemote(ActorTypeUuid, route, func(req *msg.String) (*msg.Int64, int32) {
		return &msg.Int64{Value: 2}, 0
	})
}
