package module

import (
	// Player 相关模块
	_ "lucky/server/app/game/module/player"           // 触发 init 函数（handler.go）
	_ "lucky/server/app/game/module/player/equipment" // 触发 init 函数
	_ "lucky/server/app/game/module/player/item"      // 触发 init 函数（包括 handler.go）
	_ "lucky/server/app/game/module/player/login"     // 触发 init 函数（包括 handler.go）

	// Room 相关模块
	_ "lucky/server/app/game/module/room/room" // 触发 init 函数（包括 handler.go）

	// Alliance 相关模块
	_ "lucky/server/app/game/module/alliance/alliance" // 触发 init 函数（包括 handler.go）

	// 共享模块
	_ "lucky/server/app/game/module/shared/online" // 触发 init 函数
)
