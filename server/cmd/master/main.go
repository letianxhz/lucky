package main

import (
	"os"
	"path/filepath"

	"github.com/cherry-game/cherry"
	clog "github.com/cherry-game/cherry/logger"
)

func main() {
	// 获取节点ID（从命令行参数或环境变量）
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "gc-master" // 默认节点ID
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

	// 配置master服务器
	profileFilePath := filepath.Join(profilesPath, "server.json")
	app := cherry.Configure(profileFilePath, nodeID, false, cherry.Cluster)

	clog.Info("Master server starting...")

	// 启动服务器
	app.Startup()
}
