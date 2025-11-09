package alliance

import (
	"lucky/server/app/game/module/shared/handler"

	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
)

// ActorAlliance 联盟 Actor
// 处理联盟相关的业务逻辑
type ActorAlliance struct {
	pomelo.ActorBase
	allianceId int64
}

// NewActorAlliance 创建新的联盟 Actor
func NewActorAlliance() *ActorAlliance {
	return &ActorAlliance{}
}

// AliasID 返回 Actor 的别名 ID
func (a *ActorAlliance) AliasID() string {
	return "alliance"
}

// OnInit Actor 初始化
func (a *ActorAlliance) OnInit() {
	clog.Debugf("[ActorAlliance] path = %s init!", a.PathString())

	// 注册所有联盟相关的消息处理器
	// 使用新的注解式注册方式，自动注册所有为 ActorTypeAlliance 注册的处理器
	handler.RegisterAllToActorByType(handler.ActorTypeAlliance, &a.ActorBase)
}

// OnStop Actor 停止
func (a *ActorAlliance) OnStop() {
	clog.Debugf("[ActorAlliance] path = %s exit!", a.PathString())
}
