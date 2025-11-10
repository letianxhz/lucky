package handler

import (
	"sync"

	clog "github.com/cherry-game/cherry/logger"
	"github.com/cherry-game/cherry/net/parser/pomelo"
)

// HandlerFunc 消息处理器注册函数类型
// 模块在 init() 中注册此函数，用于注册消息处理器
type HandlerFunc func(actor *pomelo.ActorBase)

var (
	// handlerFuncs 存储所有注册的消息处理器函数
	handlerFuncs []HandlerFunc
	mu           sync.Mutex
)

// Register 注册消息处理器函数
// 模块在 init() 中调用此函数注册自己的消息处理器
func Register(fn HandlerFunc) {
	mu.Lock()
	defer mu.Unlock()
	handlerFuncs = append(handlerFuncs, fn)
	clog.Debugf("[Handler] Registered message handler function")
}

// RegisterAll 注册所有已注册的消息处理器到 actor
// 在 actorPlayer.OnInit() 中调用
func RegisterAll(actor *pomelo.ActorBase) {
	mu.Lock()
	defer mu.Unlock()

	for _, fn := range handlerFuncs {
		fn(actor)
	}

	clog.Debugf("[Handler] Registered %d message handler(s) to actor", len(handlerFuncs))
}
