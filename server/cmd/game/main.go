package main

import (
	"os"
	"path/filepath"

	"lucky/server/app/game/actor"
	"lucky/server/app/game/db"
	_ "lucky/server/app/game/module" // 触发模块的 init 函数
	checkCenter "lucky/server/pkg/component/check_center"
	"lucky/server/pkg/data"
	"lucky/server/pkg/di"

	"github.com/cherry-game/cherry"
	cherrySnowflake "github.com/cherry-game/cherry/extend/snowflake"
	cstring "github.com/cherry-game/cherry/extend/string"
	cherryUtils "github.com/cherry-game/cherry/extend/utils"
	clog "github.com/cherry-game/cherry/logger"
	cherryCron "github.com/cherry-game/components/cron"
	cherryGops "github.com/cherry-game/components/gops"
)

func main() {
	// 获取节点ID（从命令行参数或环境变量）
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "10001" // 默认节点ID
	}

	if !cherryUtils.IsNumeric(nodeID) {
		panic("node parameter must is number.")
	}

	// snowflake global id
	serverId, _ := cstring.ToInt64(nodeID)
	cherrySnowflake.SetDefaultNode(serverId)

	// 获取profiles目录路径
	var profilesPath string
	if info, err := os.Stat("profiles"); err == nil && info.IsDir() {
		profilesPath = "./profiles"
	} else if info, err := os.Stat("config"); err == nil && info.IsDir() {
		profilesPath = "./config"
	} else {
		profilesPath = "./profiles"
	}

	// 配置游戏服务器
	profileFilePath := filepath.Join(profilesPath, "server.json")
	app := cherry.Configure(profileFilePath, nodeID, false, cherry.Cluster)

	// 注册组件
	app.Register(cherryGops.New())
	app.Register(cherryCron.New())
	app.Register(data.New())
	app.Register(checkCenter.New())
	app.Register(db.New())

	// 统一初始化 di 容器，为所有已注册的组件注入依赖
	// 参考 claim ioc 的实现
	// 各个模块通过 init 函数自动注册，这里统一注入依赖
	di.MustInitialize()

	// 注册所有 Actor（统一管理，便于维护和扩展）
	// 新增 Actor 时，只需在 actor/registry.go 中添加即可
	app.AddActors(actor.RegisterActors()...)

	clog.Info("Game server starting...")

	// 启动服务器
	app.Startup()
}
