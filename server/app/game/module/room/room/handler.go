package room

import (
	"lucky/server/app/game/module/shared/handler"
	"lucky/server/gen/msg"


	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// init 注册房间模块的消息处理器
func init() {
	var h = &roomHandler{}
	di.Register(h)

	// 使用统一的 RegisterHandler，所有处理器都返回 (*Response, error)，无反射，高性能
	handler.RegisterHandler(handler.ActorTypeRoom, "createRoom", h.OnCreateRoom)
	handler.RegisterHandler(handler.ActorTypeRoom, "joinRoom", h.OnJoinRoom)
	handler.RegisterHandler(handler.ActorTypeRoom, "leaveRoom", h.OnLeaveRoom)
	handler.RegisterHandler(handler.ActorTypeRoom, "getRoomInfo", h.OnGetRoomInfo)
	handler.RegisterHandler(handler.ActorTypeRoom, "broadcast", h.OnBroadcast)
}

type roomHandler struct {
	room IRoomModule `di:"auto"`
}

// OnCreateRoom 创建房间消息处理器
// 签名: func(session *cproto.Session, req *msg.CreateRoomRequest) (*msg.CreateRoomResponse, error)
// 注册层会自动进行类型转换，这里直接接收具体类型
func (r *roomHandler) OnCreateRoom(session *cproto.Session, req *msg.CreateRoomRequest) (*msg.CreateRoomResponse, error) {
	response, err := r.room.CreateRoom(session, req)
	if err != nil {
		clog.Warnf("[RoomModule] CreateRoom failed: %v", err)
		return nil, err
	}
	return response, nil
}

// OnJoinRoom 加入房间消息处理器
func (r *roomHandler) OnJoinRoom(session *cproto.Session, req *msg.JoinRoomRequest) (*msg.JoinRoomResponse, error) {
	response, err := r.room.JoinRoom(session, req)
	if err != nil {
		clog.Warnf("[RoomModule] JoinRoom failed: %v", err)
		return nil, err
	}
	return response, nil
}

// OnLeaveRoom 离开房间消息处理器
func (r *roomHandler) OnLeaveRoom(session *cproto.Session, req *msg.LeaveRoomRequest) (*msg.None, error) {
	return r.room.LeaveRoom(session, req)
}

// OnGetRoomInfo 获取房间信息消息处理器
func (r *roomHandler) OnGetRoomInfo(session *cproto.Session, req *msg.GetRoomInfoRequest) (*msg.GetRoomInfoResponse, error) {
	response, err := r.room.GetRoomInfo(session, req)
	if err != nil {
		clog.Warnf("[RoomModule] GetRoomInfo failed: %v", err)
		return nil, err
	}
	return response, nil
}

// OnBroadcast 房间广播消息处理器
func (r *roomHandler) OnBroadcast(session *cproto.Session, req *msg.RoomBroadcastRequest) (*msg.None, error) {
	return r.room.Broadcast(session, req)
}
