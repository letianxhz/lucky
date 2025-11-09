package player

import (
	"lucky/server/app/game/module/shared/handler"
	"lucky/server/app/game/module/shared/online"

	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
)

type (
	// ActorPlayer 每位登录的玩家对应一个子actor
	// 注意：首字母大写以便在其他包中使用
	actorPlayer struct {
		pomelo.ActorBase
		isOnline bool // 玩家是否在线
		playerId int64
		uid      int64
	}
)

// newActorPlayer 创建新的actorPlayer实例
func newActorPlayer() *actorPlayer {
	return &actorPlayer{
		isOnline: false,
	}
}

func (p *actorPlayer) OnInit() {
	clog.Debugf("[actorPlayer] path = %s init!", p.PathString())

	// 注册 session关闭的remote函数(网关触发连接断开后，会调用RPC发送该消息)
	p.Remote().Register("sessionClose", p.sessionClose)

	// 注册所有模块的消息处理器（通过 init() 方式注册）
	// 使用新的类型安全注册机制（V3 泛型版本，无反射，高性能）
	// 所有模块都使用 RegisterHandler 或 RegisterHandlerWithActor 在 init() 中注册
	// 优化：每个 actor 实例只注册一次，避免重复注册（内部会检查并跳过已注册的 handler）
	handler.RegisterAllToActorByType(handler.ActorTypePlayer, &p.ActorBase)
}

func (p *actorPlayer) OnStop() {
	clog.Debugf("[actorPlayer] path = %s exit!", p.PathString())
}

// sessionClose 接收角色session关闭处理
func (p *actorPlayer) sessionClose() {
	online.UnBindPlayer(p.uid)
	p.isOnline = false
	p.Exit()

	clog.Debugf("[actorPlayer] exit! uis = %d", p.uid)
}
