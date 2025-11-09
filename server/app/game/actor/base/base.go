package base

import (
	"lucky/server/app/game/module/shared/handler"

	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
)

// BaseActor Actor 基类
// 提供统一的初始化逻辑，减少重复代码
type BaseActor struct {
	pomelo.ActorBase
	actorType handler.ActorType
}

// NewBaseActor 创建基础 Actor
func NewBaseActor(actorType handler.ActorType) *BaseActor {
	return &BaseActor{
		actorType: actorType,
	}
}

// OnInit 统一的初始化逻辑
// 子类可以重写此方法，但应该先调用 base.OnInit()
func (b *BaseActor) OnInit() {
	clog.Debugf("[BaseActor] path = %s, type = %s init!", b.PathString(), b.actorType)

	// 自动注册所有为该 Actor 类型注册的 Handler
	handler.RegisterAllToActorByType(b.actorType, &b.ActorBase)
}

// OnStop 统一的停止逻辑
// 子类可以重写此方法，但应该先调用 base.OnStop()
func (b *BaseActor) OnStop() {
	clog.Debugf("[BaseActor] path = %s, type = %s exit!", b.PathString(), b.actorType)
}

// ActorType 返回 Actor 类型
func (b *BaseActor) ActorType() handler.ActorType {
	return b.actorType
}
