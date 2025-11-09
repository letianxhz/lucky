package alliance

import (
	"lucky/server/app/game/module/shared/handler"
	"lucky/server/gen/msg"
	"lucky/server/pkg/di"

	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// 注意：以下代码是示例代码，实际使用时需要：
// 1. 定义对应的 protobuf 消息类型（CreateAllianceRequest, JoinAllianceRequest 等）
// 2. 实现 IAllianceModule 接口
// 3. 在 init() 中注册模块到 DI 容器
//
// 由于 alliance 模块的接口使用 interface{}，这里暂时使用 msg.None 作为占位符
// 实际使用时应该替换为具体的 protobuf 消息类型

// init 注册联盟模块的消息处理器
func init() {
	var h = &allianceHandler{}
	di.Register(h)

	// 使用新的类型安全注册机制
	// 注意：由于接口使用 interface{}，这里暂时使用 msg.None 作为占位符
	// 实际使用时应该替换为具体的 protobuf 消息类型

	handler.RegisterHandler(handler.ActorTypeAlliance, "createAlliance", h.OnCreateAlliance)
	handler.RegisterHandler(handler.ActorTypeAlliance, "joinAlliance", h.OnJoinAlliance)
	handler.RegisterHandler(handler.ActorTypeAlliance, "leaveAlliance", h.OnLeaveAlliance)
	handler.RegisterHandler(handler.ActorTypeAlliance, "getAllianceInfo", h.OnGetAllianceInfo)
}

type allianceHandler struct {
	alliance IAllianceModule `di:"auto"`
}

// OnCreateAlliance 创建联盟消息处理器
// 注意：由于接口使用 interface{}，这里暂时使用 msg.None 作为占位符
// 实际使用时应该替换为具体的 protobuf 消息类型，如 *msg.CreateAllianceRequest
func (h *allianceHandler) OnCreateAlliance(session *cproto.Session, req *msg.None) (*msg.None, error) {
	// 处理业务逻辑
	response, err := h.alliance.CreateAlliance(session, req)
	if err != nil {
		clog.Warnf("[AllianceModule] CreateAlliance failed: %v", err)
		return nil, handler.NewErrorWithCode(500)
	}
	// 由于接口返回 interface{}，这里需要类型断言
	// 实际使用时应该返回具体的 protobuf 类型
	if response == nil {
		return &msg.None{}, nil
	}
	if resp, ok := response.(*msg.None); ok {
		return resp, nil
	}
	// 如果不是 msg.None，返回默认值
	return &msg.None{}, nil
}

// OnJoinAlliance 加入联盟消息处理器
func (h *allianceHandler) OnJoinAlliance(session *cproto.Session, req *msg.None) (*msg.None, error) {
	response, err := h.alliance.JoinAlliance(session, req)
	if err != nil {
		clog.Warnf("[AllianceModule] JoinAlliance failed: %v", err)
		return nil, handler.NewErrorWithCode(500)
	}
	if response == nil {
		return &msg.None{}, nil
	}
	if resp, ok := response.(*msg.None); ok {
		return resp, nil
	}
	return &msg.None{}, nil
}

// OnLeaveAlliance 离开联盟消息处理器
func (h *allianceHandler) OnLeaveAlliance(session *cproto.Session, req *msg.None) (*msg.None, error) {
	err := h.alliance.LeaveAlliance(session, req)
	if err != nil {
		clog.Warnf("[AllianceModule] LeaveAlliance failed: %v", err)
		return nil, handler.NewErrorWithCode(500)
	}
	return &msg.None{}, nil
}

// OnGetAllianceInfo 获取联盟信息消息处理器
func (h *allianceHandler) OnGetAllianceInfo(session *cproto.Session, req *msg.None) (*msg.None, error) {
	response, err := h.alliance.GetAllianceInfo(session, req)
	if err != nil {
		clog.Warnf("[AllianceModule] GetAllianceInfo failed: %v", err)
		return nil, handler.NewErrorWithCode(500)
	}
	if response == nil {
		return &msg.None{}, nil
	}
	if resp, ok := response.(*msg.None); ok {
		return resp, nil
	}
	return &msg.None{}, nil
}
