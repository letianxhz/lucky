package handler

import (
	"github.com/cherry-game/cherry/net/parser/pomelo"
)

// RegisterAllToActor 将所有注册的消息处理器注册到 actor（默认 Player Actor）
// 在 actorPlayer.OnInit() 中调用
// 等价于: RegisterAllToActorByType(ActorTypePlayer, actor)
func RegisterAllToActor(actor *pomelo.ActorBase) {
	RegisterAllToActorByType(ActorTypePlayer, actor)
}

// RegisterAllToActorByType 将指定 Actor 类型的所有消息处理器注册到 actor
// 在 actor.OnInit() 中调用
// 用法: handler.RegisterAllToActorByType(ActorTypeAlliance, actor)
// 注意：此函数直接使用 V3 版本（泛型，类型安全，无反射）
func RegisterAllToActorByType(actorType ActorType, actor *pomelo.ActorBase) {
	RegisterAllToActorByTypeV3(actorType, actor)
}
