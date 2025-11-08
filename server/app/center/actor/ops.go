package actor

import (
	"lucky/server/app/center/actor/ops"
)

// NewActorOps 创建运维actor
func NewActorOps() *ops.ActorOps {
	return &ops.ActorOps{}
}
