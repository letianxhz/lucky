package main

import (
	"os"
	"path/filepath"

	"github.com/cherry-game/cherry"
	clog "github.com/cherry-game/cherry/logger"
	cherryGops "github.com/ch
	"lucky/server/app/gate/config"
	"lucky/server/app/gate/service/router"
	checkCenter "lucky/server/pkg/component/check_center"
	"lucky/server/pkg/data"
)

func main() {
	// 获取profiles目录路径
	var profilesPath string
	if info, err := os.Stat("profiles"); err == nil && info.IsDir() {
		profilesPath = "./profiles"
	} else if info, err := os.Stat("config"); err == nil && info.IsDir() {
		profilesPath = "./config"
	} else {
		profilesPath = "./profiles"
	}

	// 初始化服务特定配置
	config.MustInitialize(filepath.Join(profilesPath, "server.json"))

	// 获取节点ID（从命令行参数或环境变量）
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "gc-gate-1" // 默认节点ID
	}

	// 配置网关服务器
	profileFilePath := filepath.Join(profilesPath, "server.json")
	app := cherry.Configure(profileFilePath, nodeID, true, cherry.Cluster)

	// 注册组件
	app.Register(cherryGops.New())
	app.Register(checkCenter.New())
	app.Register(data.New())

	// 创建并注册路由
	gateRouter := router.NewRouter(app)
	parser := gateRouter.BuildParser()
	app.SetNetParser(parser)

	clog.Info("Gate server starting...")

	// 启动服务器
	app.Startup()
}
