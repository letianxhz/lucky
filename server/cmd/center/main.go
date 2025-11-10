package main

import (
	"os"
	"path/filepath"

	"github.com/cherry-game/cherry"
	clog "github.com/cherry-game/cherry/logger"
	cherryCron "github.com/cherry-game/components/cron"
	"lucky/server/app/center/actor"
	"lucky/server/app/center/db"
	_ "lucky/server/app/center/module" // 触发模块的 init 函数
	"lucky/server/pkg/data"
	"lucky/server/pkg/di"
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

	// 统一初始化 di 容器，为所有已注册的组件注入依赖
	// 参考 game 的实现，各个模块通过 init 函数自动注册，这里统一注入依赖
	di.MustInitialize()

	// 注册Actor
	app.AddActors(
		actor.NewActorAccount(),
		actor.NewActorOps(),
		actor.NewActorUuid(),
	)

	clog.Info("Center server starting...")

	// 启动服务器
	app.Startup()
}
