package room

import (
	roomModule "lucky/server/app/game/module/room/room"
	"lucky/server/app/game/module/shared/handler"
	"lucky/server/gen/msg"
	"lucky/server/pkg/di"

	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// ActorRoom 房间 Actor
// 处理房间相关的业务逻辑
type ActorRoom struct {
	pomelo.ActorBase
	roomId string
}

// NewActorRoom 创建新的房间 Actor
func NewActorRoom() *ActorRoom {
	return &ActorRoom{}
}

// AliasID 返回 Actor 的别名 ID
func (r *ActorRoom) AliasID() string {
	return "room"
}

// OnInit Actor 初始化
func (r *ActorRoom) OnInit() {
	clog.Debugf("[ActorRoom] path = %s init!", r.PathString())

	// 注册所有房间相关的消息处理器（Local 消息，来自客户端）
	// 使用新的注解式注册方式，自动注册所有为 ActorTypeRoom 注册的处理器
	handler.RegisterAllToActorByType(handler.ActorTypeRoom, &r.ActorBase)

	// 注册 Remote 消息处理器（来自其他 Actor 的调用）
	r.Remote().Register("joinRoom", r.onRemoteJoinRoom)
	r.Remote().Register("leaveRoom", r.onRemoteLeaveRoom)
	r.Remote().Register("getRoomInfo", r.onRemoteGetRoomInfo)
}

// onRemoteJoinRoom 处理来自 Player Actor 的 joinRoom 调用
// 注意：本地 Remote 调用时，参数是原始对象；集群调用时，参数是 []byte
func (r *ActorRoom) onRemoteJoinRoom(req interface{}) interface{} {
	clog.Infof("[ActorRoom] Received joinRoom request from another actor: %+v", req)

	// 处理参数：可能是原始对象或 []byte
	var joinReq *msg.JoinRoomRequest
	if argBytes, ok := req.([]byte); ok {
		// 集群调用：参数是 []byte，需要解码
		var decodedReq msg.JoinRoomRequest
		if err := r.App().Serializer().Unmarshal(argBytes, &decodedReq); err == nil {
			joinReq = &decodedReq
			clog.Infof("[ActorRoom] Decoded joinRoom request (cluster): roomId=%d", joinReq.RoomId)
		} else {
			clog.Warnf("[ActorRoom] Failed to unmarshal joinRoom request: %v", err)
			joinReq = &msg.JoinRoomRequest{} // 使用默认值
		}
	} else if reqJoin, ok := req.(*msg.JoinRoomRequest); ok {
		// 本地调用：参数是原始对象
		joinReq = reqJoin
		clog.Infof("[ActorRoom] Received joinRoom request (local): roomId=%d", joinReq.RoomId)
	} else if reqInt64, ok := req.(*msg.Int64); ok {
		// 兼容旧的 Int64 格式（临时兼容）
		clog.Warnf("[ActorRoom] Received old Int64 format, converting to JoinRoomRequest: value=%d", reqInt64.Value)
		joinReq = &msg.JoinRoomRequest{
			RoomId: reqInt64.Value,
		}
	} else {
		clog.Warnf("[ActorRoom] Unexpected request type: %T, using default", req)
		joinReq = &msg.JoinRoomRequest{} // 使用默认值
	}

	// 调用 RoomModule 处理业务逻辑
	roomModuleInstance, err := di.GetByType((*roomModule.IRoomModule)(nil))
	if err != nil {
		clog.Warnf("[ActorRoom] RoomModule not found in DI container: %v", err)
	} else {
		if roomMod, ok := roomModuleInstance.(roomModule.IRoomModule); ok {
			// 创建临时 session（Remote 调用没有 session，使用请求中的 playerId）
			session := &cproto.Session{
				Uid: joinReq.PlayerId, // 从请求中获取 playerId
			}
			clog.Infof("[ActorRoom] Calling RoomModule.JoinRoom: playerId=%d, roomId=%d", joinReq.PlayerId, joinReq.RoomId)
			response, err := roomMod.JoinRoom(session, joinReq)
			if err == nil {
				response.RoomId = r.roomId // 确保返回正确的 roomId
				clog.Infof("[ActorRoom] RoomModule.JoinRoom success: roomId=%s, playerCount=%d", response.RoomId, response.PlayerCount)
				return response
			}
			clog.Warnf("[ActorRoom] RoomModule.JoinRoom failed: %v", err)
		} else {
			clog.Warnf("[ActorRoom] RoomModule type assertion failed: %T", roomModuleInstance)
		}
	}

	// 如果 RoomModule 调用失败，返回默认响应
	return &msg.JoinRoomResponse{
		RoomId:      r.roomId,
		PlayerCount: 1, // 实际应该从 RoomModule 获取
		MaxPlayers:  4,
		Success:     true,
		Message:     "joined room successfully",
	}
}

// onRemoteLeaveRoom 处理来自 Player Actor 的 leaveRoom 调用
func (r *ActorRoom) onRemoteLeaveRoom(req interface{}) interface{} {
	clog.Infof("[ActorRoom] Received leaveRoom request from another actor: %+v", req)

	// 处理参数：可能是原始对象或 []byte
	var leaveReq *msg.LeaveRoomRequest
	if argBytes, ok := req.([]byte); ok {
		// 集群调用：参数是 []byte，需要解码
		var decodedReq msg.LeaveRoomRequest
		if err := r.App().Serializer().Unmarshal(argBytes, &decodedReq); err == nil {
			leaveReq = &decodedReq
			clog.Infof("[ActorRoom] Decoded leaveRoom request (cluster): playerId=%d", leaveReq.PlayerId)
		} else {
			clog.Warnf("[ActorRoom] Failed to unmarshal leaveRoom request: %v", err)
			leaveReq = &msg.LeaveRoomRequest{} // 使用默认值
		}
	} else if reqLeave, ok := req.(*msg.LeaveRoomRequest); ok {
		// 本地调用：参数是原始对象
		leaveReq = reqLeave
		clog.Infof("[ActorRoom] Received leaveRoom request (local): playerId=%d", leaveReq.PlayerId)
	} else {
		clog.Warnf("[ActorRoom] Unexpected request type: %T, using default", req)
		leaveReq = &msg.LeaveRoomRequest{} // 使用默认值
	}

	// 调用 RoomModule 处理业务逻辑
	roomModuleInstance, err := di.GetByType((*roomModule.IRoomModule)(nil))
	if err == nil {
		if roomMod, ok := roomModuleInstance.(roomModule.IRoomModule); ok {
			// 创建临时 session
			session := &cproto.Session{
				Uid: leaveReq.PlayerId,
			}
			_, err := roomMod.LeaveRoom(session, leaveReq)
			if err == nil {
				return &msg.None{}
			}
			clog.Warnf("[ActorRoom] RoomModule.LeaveRoom failed: %v", err)
		}
	} else {
		clog.Debugf("[ActorRoom] RoomModule not found in DI container: %v", err)
	}

	// 如果 RoomModule 调用失败，返回默认响应
	return &msg.None{}
}

// onRemoteGetRoomInfo 处理来自 Player Actor 的 getRoomInfo 调用
func (r *ActorRoom) onRemoteGetRoomInfo(req interface{}) interface{} {
	clog.Infof("[ActorRoom] Received getRoomInfo request from another actor: %+v", req)

	// getRoomInfo 通常没有参数或参数为空，直接处理
	if req != nil {
		if argBytes, ok := req.([]byte); ok && len(argBytes) > 0 {
			// 集群调用：参数是 []byte
			clog.Debugf("[ActorRoom] getRoomInfo request (cluster): %d bytes", len(argBytes))
		} else {
			// 本地调用：参数可能是 msg.None 或其他类型
			clog.Debugf("[ActorRoom] getRoomInfo request (local): %T", req)
		}
	}

	// 调用 RoomModule 处理业务逻辑
	roomModuleInstance, err := di.GetByType((*roomModule.IRoomModule)(nil))
	if err == nil {
		if roomMod, ok := roomModuleInstance.(roomModule.IRoomModule); ok {
			// 创建临时 session
			session := &cproto.Session{
				Uid: 0, // getRoomInfo 不需要 UID
			}
			getInfoReq := &msg.GetRoomInfoRequest{}
			response, err := roomMod.GetRoomInfo(session, getInfoReq)
			if err == nil {
				response.RoomId = r.roomId // 确保返回正确的 roomId
				return response
			}
			clog.Warnf("[ActorRoom] RoomModule.GetRoomInfo failed: %v", err)
		}
	} else {
		clog.Debugf("[ActorRoom] RoomModule not found in DI container: %v", err)
	}

	// 如果 RoomModule 调用失败，返回默认响应
	return &msg.GetRoomInfoResponse{
		RoomId:      r.roomId,
		PlayerCount: 1, // 实际应该从 RoomModule 获取
		MaxPlayers:  4,
		PlayerIds:   []int64{}, // 实际应该从 RoomModule 获取房间内玩家ID列表
	}
}

// OnStop Actor 停止
func (r *ActorRoom) OnStop() {
	clog.Debugf("[ActorRoom] path = %s exit!", r.PathString())
}
