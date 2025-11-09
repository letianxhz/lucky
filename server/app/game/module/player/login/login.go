package login

import (
	"github.com/cherry-game/cherry/net/parser/pomelo"
	cproto "github.com/cherry-game/cherry/net/proto"
	"lucky/server/gen/msg"
)

// ILoginModule 登录模块接口
// 定义玩家登录相关的所有业务操作
type ILoginModule interface {
	// SelectPlayer 查询角色列表
	SelectPlayer(session *cproto.Session) (*msg.PlayerSelectResponse, error)

	// CreatePlayer 创建角色
	// actor 参数用于发送事件和调用 RPC
	CreatePlayer(session *cproto.Session, req *msg.PlayerCreateRequest, actor *pomelo.ActorBase) (*msg.PlayerCreateResponse, error)

	// EnterPlayer 进入游戏
	// actor 参数用于发送事件和调用 RPC
	EnterPlayer(session *cproto.Session, req *msg.Int64, actor *pomelo.ActorBase) (*msg.PlayerEnterResponse, error)
}
