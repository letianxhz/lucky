package actor

import (
	"lucky/server/app/game/actor/alliance"

	cfacade "github.com/cherry-game/cherry/facade"
)

// RegisterActors 注册所有 Actor
// 统一管理所有 Actor 的注册，便于维护和扩展
// 新增 Actor 时，只需在此函数中添加即可
func RegisterActors() []cfacade.IActorHandler {
	return []cfacade.IActorHandler{
		// 玩家 Actor（管理所有玩家子 Actor）
		NewActorPlayers(),

		// 房间 Actor（管理所有房间子 Actor）
		NewActorRooms(),

		// 联盟 Actor（如果需要管理子 Actor，可以创建 ActorAlliances）
		// 目前联盟 Actor 是单例，不需要管理器
		// alliance.NewActorAlliance(),

		// 未来可以添加更多 Actor：
		// - NewActorGuilds()      // 公会管理 Actor
		// - NewActorWorld()        // 世界 Actor
		// - NewActorChat()         // 聊天 Actor
	}
}

// GetActorByAlias 根据别名获取 Actor 构造函数（用于动态注册）
// 这允许通过配置来动态决定注册哪些 Actor
func GetActorByAlias(alias string) cfacade.IActorHandler {
	switch alias {
	case "player":
		return NewActorPlayers()
	case "rooms":
		return NewActorRooms()
	case "alliance":
		return alliance.NewActorAlliance()
	default:
		return nil
	}
}
