package actor

import (
	"lucky/server/app/center/actor/account"
)

// NewActorAccount 创建账号actor
func NewActorAccount() *account.ActorAccount {
	return &account.ActorAccount{}
}
