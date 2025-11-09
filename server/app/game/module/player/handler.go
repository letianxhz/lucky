package player

import (
	"fmt"

	"lucky/server/app/game/module/shared/handler"
	"lucky/server/gen/msg"

	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// init 注册玩家模块的消息处理器（用于与 Room Actor 通信）
func init() {
	var h = &playerHandler{}
	// 注意：playerHandler 不需要注册到 di，因为它不需要依赖注入

	// 使用新的类型安全注册机制，需要 actor 参数来调用 Room Actor
	handler.RegisterHandlerWithActor(handler.ActorTypePlayer, "joinRoom", h.OnJoinRoom)
	handler.RegisterHandlerWithActor(handler.ActorTypePlayer, "leaveRoom", h.OnLeaveRoom)
	handler.RegisterHandlerWithActor(handler.ActorTypePlayer, "getRoomInfo", h.OnGetRoomInfo)
}

type playerHandler struct {
	// 不需要依赖注入，因为只是转发请求到 Room Actor
}

// OnJoinRoom 玩家加入房间消息处理器
// 这个处理器会调用 Room Actor 的 joinRoom 方法
func (h *playerHandler) OnJoinRoom(session *cproto.Session, req *msg.JoinRoomRequest, actor *pomelo.ActorBase) (*msg.JoinRoomResponse, error) {
	// 如果请求中没有指定房间ID，使用默认值
	roomId := "room_001"
	if req.RoomId > 0 {
		roomId = fmt.Sprintf("room_%03d", req.RoomId)
	}

	// 构建 Room Actor 的路径: rooms.room_001 (rooms 是管理 Actor，room_001 是子 Actor)
	roomActorPath := cfacade.NewChildPath("", "rooms", roomId)

	playerUid := session.Uid
	clog.Infof("[PlayerHandler] Player %d requesting to join room %s", playerUid, roomId)

	// 将玩家 UID 添加到请求中（RoomModule 需要从 session 获取，但 Remote 调用没有 session）
	joinReq := &msg.JoinRoomRequest{
		RoomId:   req.RoomId,
		PlayerId: int64(playerUid),
	}

	// 使用 CallWait 等待 Room Actor 的响应
	var reply msg.JoinRoomResponse
	code := actor.CallWait(roomActorPath, "joinRoom", joinReq, &reply)
	if code != 0 {
		clog.Warnf("[PlayerHandler] Room Actor joinRoom failed: code=%d", code)
		return nil, handler.NewErrorWithCode(int32(code))
	}

	clog.Infof("[PlayerHandler] Player joined room successfully: roomId=%s, playerCount=%d", reply.RoomId, reply.PlayerCount)
	return &reply, nil
}

// OnLeaveRoom 玩家离开房间消息处理器
func (h *playerHandler) OnLeaveRoom(session *cproto.Session, req *msg.LeaveRoomRequest, actor *pomelo.ActorBase) (*msg.LeaveRoomResponse, error) {
	roomId := "room_001"
	roomActorPath := cfacade.NewChildPath("", "rooms", roomId)

	clog.Infof("[PlayerHandler] Player %d requesting to leave room %s", session.Uid, roomId)

	// 如果请求中没有 PlayerId，使用 session.Uid
	leaveReq := &msg.LeaveRoomRequest{
		PlayerId: req.PlayerId,
	}
	if leaveReq.PlayerId == 0 {
		leaveReq.PlayerId = int64(session.Uid)
	}

	var reply msg.LeaveRoomResponse
	code := actor.CallWait(roomActorPath, "leaveRoom", leaveReq, &reply)
	if code != 0 {
		clog.Warnf("[PlayerHandler] Room Actor leaveRoom failed: code=%d", code)
		return nil, handler.NewErrorWithCode(int32(code))
	}

	clog.Infof("[PlayerHandler] Player left room successfully: %s", reply.Message)
	return &reply, nil
}

// OnGetRoomInfo 获取房间信息消息处理器
func (h *playerHandler) OnGetRoomInfo(session *cproto.Session, req *msg.GetRoomInfoRequest, actor *pomelo.ActorBase) (*msg.GetRoomInfoResponse, error) {
	roomId := "room_001"
	roomActorPath := cfacade.NewChildPath("", "rooms", roomId)

	clog.Infof("[PlayerHandler] Player %d requesting room info for %s", session.Uid, roomId)

	var reply msg.GetRoomInfoResponse
	code := actor.CallWait(roomActorPath, "getRoomInfo", req, &reply)
	if code != 0 {
		clog.Warnf("[PlayerHandler] Room Actor getRoomInfo failed: code=%d", code)
		return nil, handler.NewErrorWithCode(int32(code))
	}

	clog.Infof("[PlayerHandler] Got room info: roomId=%s, playerCount=%d, maxPlayers=%d",
		reply.RoomId, reply.PlayerCount, reply.MaxPlayers)
	return &reply, nil
}
