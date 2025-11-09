package actor

import (
	"lucky/server/app/game/actor/room"
)

// NewActorRooms 创建房间管理 Actor
func NewActorRooms() *room.ActorRooms {
	return room.NewActorRooms()
}
