package actor

import (
	"lucky/server/app/game/actor/player"
)

// NewActorPlayers 创建玩家总管理actor
func NewActorPlayers() *player.ActorPlayers {
	return &player.ActorPlayers{}
}
