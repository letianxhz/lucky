package uuid

import (
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	"lucky/server/pkg/handler"
)

type ActorUuid struct {
	cactor.Base
}

func (p *ActorUuid) AliasID() string {
	return "uuid"
}

// OnInit 注册 remote 函数
func (p *ActorUuid) OnInit() {
	clog.Debugf("[ActorUuid] path = %s init!", p.PathString())

	// 使用统一的 RegisterAllToActorByTypeV3Remote 注册所有 Remote 消息处理器
	handler.RegisterAllToActorByTypeV3Remote(handler.ActorTypeUuid, p)
}
