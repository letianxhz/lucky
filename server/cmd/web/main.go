package main

import (
	"os"
	"path/filepath"

	"github.com/cherry-game/cherry"
	cherryFile "github.com/cherry-game/cherry/extend/file"
	clog "github.com/cherry-game/cherry/logger"
	cherryCron "github.com/cherry-game/components/cron"
	cherryGin "github.com/cherry-game/components/gin"
	"github.com/gin-gonic/gin"
	"lucky/server/app/web/controller"
	"lucky/server/app/web/sdk"
	checkCenter "lucky/server/pkg/component/check_center"
	"lucky/server/pkg/data"
)

func main() {
	// 获取节点ID（从命令行参数或环境变量）
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "gc-web-1" // 默认节点ID
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

	// 获取工作目录（用于查找 static 和 view 目录）
	workDir, _ := os.Getwd()

	// 配置web服务器
	profileFilePath := filepath.Join(profilesPath, "server.json")
	app := cherry.Configure(profileFilePath, nodeID, false, cherry.Cluster)

	// 注册组件
	app.Register(cherryCron.New())
	app.Register(checkCenter.New())
	app.Register(data.New())
	app.Register(httpServerComponent(app.Address(), workDir))

	// 加载sdk逻辑
	sdk.Init(app)

	clog.Info("Web server starting...")

	// 启动服务器
	app.Startup()
}

func httpServerComponent(addr string, workDir string) *cherryGin.Component {
	gin.SetMode(gin.DebugMode)

	// new http server
	httpServer := cherryGin.NewHttp("http_server", addr)
	httpServer.Use(cherryGin.Cors())
	httpServer.Use(cherryGin.RecoveryWithZap(true))

	// 映射h5客户端静态文件到static目录
	staticPath := filepath.Join(workDir, "app/web/static")
	httpServer.Static("/static", staticPath)

	// 加载./view目录的html模板文件
	viewPath := filepath.Join(workDir, "app/web/view")
	viewFiles := cherryFile.WalkFiles(viewPath, ".html")
	if len(viewFiles) < 1 {
		panic("view files not found.")
	}
	httpServer.LoadHTMLFiles(viewFiles...)

	// 注册 controller
	httpServer.Register(new(controller.Controller))

	return httpServer
}
