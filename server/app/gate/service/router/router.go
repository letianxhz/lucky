package router

import (
	"github.com/cherry-game/cherry"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cconnector "github.com/cherry-game/cherry/net/connector"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	"lucky/server/app/gate/actor"
	"lucky/server/app/gate/config"
)

// Router 网关路由管理器
type Router struct {
	app *cherry.AppBuilder
}

// NewRouter 创建网关路由管理器
func NewRouter(app *cherry.AppBuilder) *Router {
	return &Router{
		app: app,
	}
}

// BuildParser 构建网络解析器（使用 pomelo 协议）
func (r *Router) BuildParser() cfacade.INetParser {
	cfg := config.Get()

	// 使用pomelo网络数据包解析器
	agentActor := pomelo.NewActor("user")

	// 创建一个tcp监听，用于client/robot压测机器人连接网关tcp
	if cfg.TCPAddr != "" {
		agentActor.AddConnector(cconnector.NewTCP(cfg.TCPAddr))
		clog.Infof("[GateRouter] TCP connector added: %s", cfg.TCPAddr)
	}

	// 再创建一个websocket监听，用于h5客户端建立连接
	if cfg.WSAddr != "" {
		agentActor.AddConnector(cconnector.NewWS(cfg.WSAddr))
		clog.Infof("[GateRouter] WebSocket connector added: %s", cfg.WSAddr)
	} else if r.app != nil {
		agentActor.AddConnector(cconnector.NewWS(r.app.Address()))
		clog.Infof("[GateRouter] WebSocket connector added: %s", r.app.Address())
	}

	// 当有新连接创建Agent时，启动一个自定义(ActorAgent)的子actor
	agentActor.SetOnNewAgent(func(newAgent *pomelo.Agent) {
		childActor := actor.NewAgentActor()
		// 设置关闭回调
		newAgent.AddOnClose(func(agent *pomelo.Agent) {
			childActor.OnSessionClose(agent)
		})
		// 创建子actor，actorID == sid
		agentActor.Child().Create(newAgent.SID(), childActor)
	})

	// 设置数据路由函数
	agentActor.SetOnDataRoute(onPomeloDataRoute)

	return agentActor
}
