package main

import (
	"os"
	"path/filepath"

	"lucky/server/app/game/actor"
	"lucky/server/app/game/db"
	checkCenter "lucky/server/pkg/component/check_center"
	"lucky/server/pkg/data"

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

	// 注册Actor
	app.AddActors(
		actor.NewActorPlayers(),
	)

	clog.Info("Game server starting...")

	// 启动服务器
	app.Startup()
}
