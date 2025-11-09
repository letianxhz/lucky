package main

import (
	"os"
	"path/filepath"

	"github.com/cherry-game/cherry"
	clog "github.com/cherry-game/cherry/logger"
	cherryCron "github.com/cherry-game/components/cron"
	"lucky/server/app/center/actor"
	"lucky/server/app/center/db"
	"lucky/server/pkg/data"
)

func main() {
	// 获取节点ID（从命令行参数或环境变量）
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "gc-center" // 默认节点ID
	}

	// 获取profiles目录路径
	var profilesPath string
	if info, err := os.Stat("profiles"); err == nil && info.IsDir() {
		profilesPath = "./profiles"
	} else if info, err := os.Stat("config"); err == nil && info.IsDir() {
		profilesPath = "./config"
	} else {
		profilesPath = "./profiles"
	}

	// 配置中心服务器
	profileFilePath := filepath.Join(profilesPath, "server.json")
	app := cherry.Configure(profileFilePath, nodeID, false, cherry.Cluster)

	// 注册组件
	app.Register(cherryCron.New())
	app.Register(data.New())
	app.Register(db.New())

	// 注册Actor
	app.AddActors(
		actor.NewActorAccount(),
		actor.NewActorOps(),
	)

	clog.Info("Center server starting...")

	// 启动服务器
	app.Startup()
}
