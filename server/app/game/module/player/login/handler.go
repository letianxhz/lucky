package login

import (
	"lucky/server/app/game/module/shared/handler"
	"lucky/server/gen/msg"

	"lucky/server/pkg/code"
	"lucky/server/gen/msg"

	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// init 注册登录模块的消息处理器
func init() {
	var h = &loginHandler{}
	di.Register(h)

	// 使用新的类型安全注册机制（V3 泛型版本，无反射，高性能）
	handler.RegisterHandler(handler.ActorTypePlayer, "select", h.OnSelect)
	// Create 和 Enter 需要 actor 参数，使用 RegisterHandlerWithActor
	handler.RegisterHandlerWithActor(handler.ActorTypePlayer, "create", h.OnCreate)
	handler.RegisterHandlerWithActor(handler.ActorTypePlayer, "enter", h.OnEnter)
}

type loginHandler struct {
	login ILoginModule `di:"auto"`
}

// OnSelect 查询角色列表消息处理器
// 签名: func(session *cproto.Session, req *msg.None) (*msg.PlayerSelectResponse, error)
// 注册层会自动进行类型转换，这里直接接收具体类型，无需类型断言
func (h *loginHandler) OnSelect(session *cproto.Session, _ *msg.None) (*msg.PlayerSelectResponse, error) {
	response, err := h.login.SelectPlayer(session)
	if err != nil {
		clog.Warnf("[LoginHandler] SelectPlayer failed: %v", err)
		return nil, handler.NewErrorWithCode(code.Error)
	}
	return response, nil
}

// OnCreate 创建角色消息处理器（带 actor 参数）
// 签名: func(session *cproto.Session, req *msg.PlayerCreateRequest, actor *pomelo.ActorBase) (*msg.PlayerCreateResponse, error)
// 注册层会自动进行类型转换，这里直接接收具体类型，无需类型断言
func (h *loginHandler) OnCreate(session *cproto.Session, req *msg.PlayerCreateRequest, actor *pomelo.ActorBase) (*msg.PlayerCreateResponse, error) {
	response, err := h.login.CreatePlayer(session, req, actor)
	if err != nil {
		clog.Warnf("[LoginHandler] CreatePlayer failed: %v", err)
		return nil, handler.NewErrorWithCode(code.PlayerCreateFail)
	}
	return response, nil
}

// OnEnter 进入游戏消息处理器（带 actor 参数）
// 签名: func(session *cproto.Session, req *msg.Int64, actor *pomelo.ActorBase) (*msg.PlayerEnterResponse, error)
// 注册层会自动进行类型转换，这里直接接收具体类型，无需类型断言
func (h *loginHandler) OnEnter(session *cproto.Session, req *msg.Int64, actor *pomelo.ActorBase) (*msg.PlayerEnterResponse, error) {
	response, err := h.login.EnterPlayer(session, req, actor)
	if err != nil {
		clog.Warnf("[LoginHandler] EnterPlayer failed: %v", err)
		return nil, handler.NewErrorWithCode(code.PlayerIDError)
	}
	return response, nil
}
