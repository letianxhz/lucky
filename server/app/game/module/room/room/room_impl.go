package room

import (
	"errors"
	"fmt"
	"lucky/server/gen/msg"
	"lucky/server/pkg/di"
	"sync"

	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// RoomModule 房间模块实现
type RoomModule struct {
	// 临时内存存储（用于测试）
	mu    sync.RWMutex
	rooms map[string]*RoomInfo // key: roomId, value: RoomInfo
}

// RoomInfo 房间信息
type RoomInfo struct {
	RoomId     string
	PlayerIds  []int64
	MaxPlayers int32
}

// init 初始化房间模块并注册到 di 容器
func init() {
	var v = &RoomModule{
		rooms: make(map[string]*RoomInfo),
	}
	di.Register(v)
}

// CreateRoom 创建房间
func (m *RoomModule) CreateRoom(session *cproto.Session, req *msg.CreateRoomRequest) (*msg.CreateRoomResponse, error) {
	// 示例实现：创建一个简单的房间
	roomId := "room_001"

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rooms[roomId]; exists {
		return nil, errors.New("room already exists")
	}

	maxPlayers := int32(4)
	if req.MaxPlayers > 0 {
		maxPlayers = req.MaxPlayers
	}

	m.rooms[roomId] = &RoomInfo{
		RoomId:     roomId,
		PlayerIds:  []int64{},
		MaxPlayers: maxPlayers,
	}

	clog.Infof("[RoomModule] Created room: %s, maxPlayers=%d", roomId, maxPlayers)

	return &msg.CreateRoomResponse{
		RoomId:  roomId,
		Success: true,
		Message: "room created successfully",
	}, nil
}

// JoinRoom 加入房间
func (m *RoomModule) JoinRoom(session *cproto.Session, req *msg.JoinRoomRequest) (*msg.JoinRoomResponse, error) {
	// 根据 req.RoomId 生成 roomId 字符串
	roomId := "room_001" // 默认房间
	if req.RoomId > 0 {
		roomId = fmt.Sprintf("room_%03d", req.RoomId)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	room, exists := m.rooms[roomId]
	if !exists {
		// 如果房间不存在，创建一个默认房间
		room = &RoomInfo{
			RoomId:     roomId,
			PlayerIds:  []int64{},
			MaxPlayers: 4,
		}
		m.rooms[roomId] = room
		clog.Infof("[RoomModule] Created new room: %s", roomId)
	}

	if len(room.PlayerIds) >= int(room.MaxPlayers) {
		return nil, errors.New("room is full")
	}

	// 获取玩家ID（优先从请求中获取，如果没有则从 session 获取）
	playerId := req.PlayerId
	if playerId == 0 {
		playerId = int64(session.Uid)
	}

	// 检查玩家是否已在房间中
	for _, id := range room.PlayerIds {
		if id == playerId {
			clog.Infof("[RoomModule] Player %d already in room %s", playerId, roomId)
			return &msg.JoinRoomResponse{
				RoomId:      roomId,
				PlayerCount: int32(len(room.PlayerIds)),
				MaxPlayers:  room.MaxPlayers,
				Success:     true,
				Message:     "already in room",
			}, nil
		}
	}

	// 添加玩家到房间
	room.PlayerIds = append(room.PlayerIds, playerId)

	clog.Infof("[RoomModule] Player %d joined room %s, current playerCount=%d",
		playerId, roomId, len(room.PlayerIds))

	return &msg.JoinRoomResponse{
		RoomId:      roomId,
		PlayerCount: int32(len(room.PlayerIds)),
		MaxPlayers:  room.MaxPlayers,
		Success:     true,
		Message:     "joined room successfully",
	}, nil
}

// LeaveRoom 离开房间
func (m *RoomModule) LeaveRoom(session *cproto.Session, req *msg.LeaveRoomRequest) (*msg.None, error) {
	roomId := "room_001" // 示例
	playerId := req.PlayerId

	m.mu.Lock()
	defer m.mu.Unlock()

	room, exists := m.rooms[roomId]
	if !exists {
		return nil, errors.New("room not found")
	}

	// 从列表中移除玩家
	for i, id := range room.PlayerIds {
		if id == playerId {
			room.PlayerIds = append(room.PlayerIds[:i], room.PlayerIds[i+1:]...)
			clog.Infof("[RoomModule] Player %d left room %s", playerId, roomId)
			return &msg.None{}, nil
		}
	}

	clog.Warnf("[RoomModule] Player %d not found in room %s", playerId, roomId)
	return &msg.None{}, nil // 玩家不在房间中，不算错误
}

// GetRoomInfo 获取房间信息
func (m *RoomModule) GetRoomInfo(session *cproto.Session, req *msg.GetRoomInfoRequest) (*msg.GetRoomInfoResponse, error) {
	roomId := "room_001" // 示例

	m.mu.RLock()
	defer m.mu.RUnlock()

	room, exists := m.rooms[roomId]
	if !exists {
		// 如果房间不存在，返回默认信息
		clog.Debugf("[RoomModule] Room %s not found, returning default info", roomId)
		return &msg.GetRoomInfoResponse{
			RoomId:      roomId,
			PlayerCount: 0,
			MaxPlayers:  4,
			PlayerIds:   []int64{},
		}, nil
	}

	clog.Debugf("[RoomModule] Getting room info: roomId=%s, playerCount=%d, playerIds=%v",
		room.RoomId, len(room.PlayerIds), room.PlayerIds)

	return &msg.GetRoomInfoResponse{
		RoomId:      room.RoomId,
		PlayerCount: int32(len(room.PlayerIds)),
		MaxPlayers:  room.MaxPlayers,
		PlayerIds:   room.PlayerIds,
	}, nil
}

// Broadcast 房间广播
func (m *RoomModule) Broadcast(session *cproto.Session, req *msg.RoomBroadcastRequest) (*msg.None, error) {
	roomId := "room_001" // 示例

	m.mu.RLock()
	defer m.mu.RUnlock()

	room, exists := m.rooms[roomId]
	if !exists {
		return nil, errors.New("room not found")
	}

	clog.Infof("[RoomModule] Broadcasting to room %s, players: %v, message: %s", roomId, room.PlayerIds, req.Message)
	return &msg.None{}, nil
}
