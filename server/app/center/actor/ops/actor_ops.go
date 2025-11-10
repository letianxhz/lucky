package ops

import (
	ccode "github.com/cherry-game/cherry/code"
	cactor "github.com/che
	"lucky/server/gen/msg"
)

var (
	pingReturn = &msg.Bool{Value: true}
)

type (
	ActorOps struct {
		cactor.Base
	}
)

func (p *ActorOps) AliasID() string {
	return "ops"
}

// OnInit 注册remote函数
func (p *ActorOps) OnInit() {
	p.Remote().Register("ping", p.ping)
}

// ping 请求center是否响应
func (p *ActorOps) ping() (*msg.Bool, int32) {
	return pingReturn, ccode.OK
}
