package room

import (
	"lucky/server/gen/msg"

	cproto "github.com/cherry-game/cherry/net/proto"
)

// IRoomModule 房间模块接口
// 定义房间相关的所有业务操作
type IRoomModule interface {
	// CreateRoom 创建房间
	CreateRoom(session *cproto.Session, req *msg.CreateRoomRequest) (*msg.CreateRoomResponse, error)

	// JoinRoom 加入房间
	JoinRoom(session *cproto.Session, req *msg.JoinRoomRequest) (*msg.JoinRoomResponse, error)

	// LeaveRoom 离开房间
	LeaveRoom(session *cproto.Session, req *msg.LeaveRoomRequest) (*msg.None, error)

	// GetRoomInfo 获取房间信息
	GetRoomInfo(session *cproto.Session, req *msg.GetRoomInfoRequest) (*msg.GetRoomInfoResponse, error)

	// Broadcast 房间广播
	Broadcast(session *cproto.Session, req *msg.RoomBroadcastRequest) (*msg.None, error)
}
