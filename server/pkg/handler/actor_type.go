package handler

// ActorType Actor 类型定义
// 用于区分不同类型的 Actor，每个 Actor 类型可以有独立的消息处理器
type ActorType string

const (
	// ActorTypePlayer 玩家 Actor
	// 处理玩家相关的消息，如登录、购买道具等
	ActorTypePlayer ActorType = "player"

	// ActorTypeAlliance 联盟 Actor
	// 处理联盟相关的消息，如创建联盟、加入联盟等
	ActorTypeAlliance ActorType = "alliance"

	// ActorTypeRoom 房间 Actor
	// 处理房间相关的消息，如创建房间、加入房间等
	ActorTypeRoom ActorType = "room"

	// ActorTypeGuild 公会 Actor
	// 处理公会相关的消息
	ActorTypeGuild ActorType = "guild"

	// ActorTypeWorld 世界 Actor
	// 处理世界相关的消息，如世界聊天、世界事件等
	ActorTypeWorld ActorType = "world"

	// ActorTypeUuid UUID Actor
	// 处理 UUID 相关的消息（Remote 调用）
	ActorTypeUuid ActorType = "uuid"
)

// String 返回 Actor 类型的字符串表示
func (t ActorType) String() string {
	return string(t)
}

// IsValid 检查 Actor 类型是否有效
func (t ActorType) IsValid() bool {
	switch t {
	case ActorTypePlayer, ActorTypeAlliance, ActorTypeRoom, ActorTypeGuild, ActorTypeWorld, ActorTypeUuid:
		return true
	default:
		return false
	}
}
