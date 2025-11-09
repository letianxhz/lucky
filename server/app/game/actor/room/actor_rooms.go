package room

import (
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
)

// ActorRooms 房间总管理actor
// 管理所有房间 Actor 的创建和生命周期
type ActorRooms struct {
	pomelo.ActorBase
}

// NewActorRooms 创建房间管理 Actor
func NewActorRooms() *ActorRooms {
	return &ActorRooms{}
}

// AliasID 返回 Actor 的别名 ID
func (r *ActorRooms) AliasID() string {
	return "rooms"
}

// OnInit Actor 初始化
func (r *ActorRooms) OnInit() {
	clog.Infof("[ActorRooms] OnInit. path = %s", r.PathString())
}

// OnFindChild 动态创建 room child actor
func (r *ActorRooms) OnFindChild(msg *cfacade.Message) (cfacade.IActor, bool) {
	// 动态创建 room child actor
	childID := msg.TargetPath().ChildID
	childActor, err := r.Child().Create(childID, &ActorRoom{
		roomId: childID,
	})

	if err != nil {
		clog.Warnf("[ActorRooms] Failed to create room actor: %s, err: %v", childID, err)
		return nil, false
	}

	clog.Infof("[ActorRooms] Created room actor: %s", childID)
	return childActor, true
}

// OnStop Actor 停止
func (r *ActorRooms) OnStop() {
	clog.Infof("[ActorRooms] OnStop. path = %s", r.PathString())
}
