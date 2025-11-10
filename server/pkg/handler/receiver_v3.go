package handler

import (
	"fmt"
	"sync"

	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// ErrorWithCode 带错误码的错误类型
type ErrorWithCode struct {
	Code int32
	Err  error
}

func (e *ErrorWithCode) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d, err=%v", e.Code, e.Err)
	}
	return fmt.Sprintf("code=%d", e.Code)
}

// NewErrorWithCode 创建带错误码的错误
func NewErrorWithCode(code int32) *ErrorWithCode {
	return &ErrorWithCode{Code: code}
}

// NewErrorWithCodeAndErr 创建带错误码和原始错误的错误
func NewErrorWithCodeAndErr(code int32, err error) *ErrorWithCode {
	return &ErrorWithCode{Code: code, Err: err}
}

// GenericHandlerFunc 返回 error 的处理器函数类型（泛型版本，最常用）
type GenericHandlerFunc[T any] func(session *cproto.Session, req *T) error

// GenericHandlerFuncWithResponse 返回 (response, error) 的处理器函数类型（泛型版本）
type GenericHandlerFuncWithResponse[TReq any, TResp any] func(session *cproto.Session, req *TReq) (*TResp, error)

// GenericHandlerFuncWithResponseAndActor 返回 (response, error) 的处理器函数类型（泛型版本，带 actor 参数）
type GenericHandlerFuncWithResponseAndActor[TReq any, TResp any] func(session *cproto.Session, req *TReq, actor *pomelo.ActorBase) (*TResp, error)

// GenericHandlerFuncNoReturn 无返回值的处理器函数类型（泛型版本）
type GenericHandlerFuncNoReturn[T any] func(session *cproto.Session, req *T)

// GenericHandlerFuncRemote 返回 (response, int32) 的 Remote 处理器函数类型（泛型版本）
// 用于 center 服务的 Remote 调用
type GenericHandlerFuncRemote[TReq any, TResp any] func(req *TReq) (*TResp, int32)

// HandlerConstraint 处理器约束接口，用于统一注册
// 支持两种类型：
// - GenericHandlerFunc[T] (返回 error)
// - GenericHandlerFuncWithResponse[TReq, TResp] (返回 response, error)
type HandlerConstraint interface {
	~func(*cproto.Session, any) error | ~func(*cproto.Session, any) (any, error)
}

// msgHandlerInfoV3 消息处理器信息（泛型版本，无反射）
type msgHandlerInfoV3 struct {
	actorType ActorType
	route     string
	// 调用函数：直接调用，无需反射（通过闭包保存类型信息）
	callFunc func(actor *pomelo.ActorBase, session *cproto.Session, msg interface{})
	// Remote 调用函数：用于 Remote 调用的 handler（可选）
	// 签名: func(req interface{}) (resp interface{}, code int32)
	// 框架会通过反射正确处理返回值
	remoteCallFunc func(req interface{}) (interface{}, int32)
}

var (
	// msgHandlersV3 存储所有注册的消息处理器（泛型版本）
	msgHandlersV3     = make(map[string]*msgHandlerInfoV3)
	msgHandlersV3Lock sync.Mutex

	// registeredActors 记录每个 actor 已经注册的 route，避免重复注册
	// key: actorPath:actorType, value: set of routes
	registeredActors     = make(map[string]map[string]bool)
	registeredActorsLock sync.Mutex

	// actorTypeHandlers 缓存每种 actor 类型的 handler 列表，避免重复遍历
	// key: actorType, value: handler 列表
	actorTypeHandlers     = make(map[ActorType][]*msgHandlerInfoV3)
	actorTypeHandlersLock sync.Mutex
)

// makeHandlerKey 生成处理器键
func makeHandlerKey(actorType ActorType, route string) string {
	return string(actorType) + ":" + route
}

// RegisterHandler 统一的注册函数，自动识别处理器返回值类型
// 支持以下两种处理器签名：
// 1. func(session *cproto.Session, req *TReq) (*TResp, error)  - 带返回值
// 2. func(session *cproto.Session, req *T) error                - 只返回 error
//
// 用法（统一接口，自动识别）:
//   - handler.RegisterHandler(handler.ActorTypeRoom, "createRoom", h.OnCreateRoom)  // 自动识别为带返回值
//   - handler.RegisterHandler(handler.ActorTypeRoom, "leaveRoom", h.OnLeaveRoom)      // 自动识别为只返回 error
//
// 实现说明：由于 Go 不支持函数重载，我们通过泛型类型推断来实现统一接口
// 编译器会根据处理器的返回值类型自动选择正确的实现
func RegisterHandler[TReq any, TResp any](actorType ActorType, route string, handler GenericHandlerFuncWithResponse[TReq, TResp]) {
	registerHandlerWithResponse(actorType, route, handler)
}

// RegisterHandlerWithActor 注册带 actor 参数的处理器
// 处理器签名: func(session *cproto.Session, req *TReq, actor *pomelo.ActorBase) (*TResp, error)
func RegisterHandlerWithActor[TReq any, TResp any](actorType ActorType, route string, handler GenericHandlerFuncWithResponseAndActor[TReq, TResp]) {
	registerHandlerWithResponseAndActor(actorType, route, handler)
}

// RegisterHandlerError 注册返回 error 的处理器（用于只返回 error 的情况）
// 当处理器只返回 error 时，使用这个函数
// 注意：虽然函数名不同，但用法与 RegisterHandler 相同，都是统一接口
func RegisterHandlerError[T any](actorType ActorType, route string, handler GenericHandlerFunc[T]) {
	registerHandlerError(actorType, route, handler)
}

// RegisterHandlerRemote 注册 Remote 消息处理器（用于 center 服务）
// 处理器签名: func(req *TReq) (*TResp, int32)
// 用法: handler.RegisterHandlerRemote(handler.ActorTypeUuid, "allocateUUID", h.OnAllocateUUID)
func RegisterHandlerRemote[TReq any, TResp any](actorType ActorType, route string, handler GenericHandlerFuncRemote[TReq, TResp]) {
	registerHandlerRemote(actorType, route, handler)
}

// registerHandlerError 内部函数：注册只返回 error 的处理器
func registerHandlerError[T any](actorType ActorType, route string, handler GenericHandlerFunc[T]) {
	registerHandlerInternal(actorType, route, handler, func(actor *pomelo.ActorBase, session *cproto.Session, msg interface{}) {
		app := actor.App()
		if app == nil {
			clog.Warnf("[Handler] Actor app is nil for route=%s", route)
			actor.ResponseCode(session, 500)
			return
		}

		// 类型转换：将 interface{} 转换为 *T
		var req *T
		if msg == nil {
			var zero T
			req = &zero
		} else if msgPtr, ok := msg.(*T); ok {
			req = msgPtr
		} else {
			clog.Warnf("[Handler] Message type mismatch for route=%s: expected *%T, got %T, attempting conversion", route, (*T)(nil), msg)
			msgBytes, err := app.Serializer().Marshal(msg)
			if err != nil {
				clog.Warnf("[Handler] Failed to marshal message for route=%s: %v", route, err)
				actor.ResponseCode(session, 500)
				return
			}
			var zero T
			req = &zero
			if err := app.Serializer().Unmarshal(msgBytes, req); err != nil {
				clog.Warnf("[Handler] Failed to unmarshal message for route=%s: %v", route, err)
				actor.ResponseCode(session, 500)
				return
			}
		}

		// 直接调用 handler，无需反射
		err := handler(session, req)
		if err != nil {
			clog.Warnf("[Handler] Handler returned error for route=%s: %v", route, err)
			// 检查是否是带错误码的错误
			if errWithCode, ok := err.(*ErrorWithCode); ok {
				actor.ResponseCode(session, errWithCode.Code)
			} else {
				actor.ResponseCode(session, 500)
			}
			return
		}
	})
}

// registerHandlerWithResponseAndActor 内部函数：注册带返回值和 actor 参数的处理器
func registerHandlerWithResponseAndActor[TReq any, TResp any](actorType ActorType, route string, handler GenericHandlerFuncWithResponseAndActor[TReq, TResp]) {
	registerHandlerInternal(actorType, route, handler, func(actor *pomelo.ActorBase, session *cproto.Session, msg interface{}) {
		app := actor.App()
		if app == nil {
			clog.Warnf("[Handler] Actor app is nil for route=%s", route)
			actor.ResponseCode(session, 500)
			return
		}

		// 类型转换：将 interface{} 转换为 *TReq
		var req *TReq
		if msg == nil {
			var zero TReq
			req = &zero
		} else if msgPtr, ok := msg.(*TReq); ok {
			req = msgPtr
		} else {
			clog.Warnf("[Handler] Message type mismatch for route=%s: expected *%T, got %T, attempting conversion", route, (*TReq)(nil), msg)
			msgBytes, err := app.Serializer().Marshal(msg)
			if err != nil {
				clog.Warnf("[Handler] Failed to marshal message for route=%s: %v", route, err)
				actor.ResponseCode(session, 500)
				return
			}
			var zero TReq
			req = &zero
			if err := app.Serializer().Unmarshal(msgBytes, req); err != nil {
				clog.Warnf("[Handler] Failed to unmarshal message for route=%s: %v", route, err)
				actor.ResponseCode(session, 500)
				return
			}
		}

		// 直接调用 handler，传递 actor 作为第三个参数
		response, err := handler(session, req, actor)
		if err != nil {
			clog.Warnf("[Handler] Handler returned error for route=%s: %v", route, err)
			// 检查是否是带错误码的错误
			if errWithCode, ok := err.(*ErrorWithCode); ok {
				actor.ResponseCode(session, errWithCode.Code)
			} else {
				actor.ResponseCode(session, 500)
			}
			return
		}
		if response != nil {
			actor.Response(session, response)
		}
	})
}

// registerHandlerWithResponse 内部函数：注册带返回值的处理器
func registerHandlerWithResponse[TReq any, TResp any](actorType ActorType, route string, handler GenericHandlerFuncWithResponse[TReq, TResp]) {
	registerHandlerInternal(actorType, route, handler, func(actor *pomelo.ActorBase, session *cproto.Session, msg interface{}) {
		app := actor.App()
		if app == nil {
			clog.Warnf("[Handler] Actor app is nil for route=%s", route)
			actor.ResponseCode(session, 500)
			return
		}

		// 类型转换：将 interface{} 转换为 *TReq
		var req *TReq
		if msg == nil {
			var zero TReq
			req = &zero
		} else if msgPtr, ok := msg.(*TReq); ok {
			req = msgPtr
		} else {
			clog.Warnf("[Handler] Message type mismatch for route=%s: expected *%T, got %T, attempting conversion", route, (*TReq)(nil), msg)
			msgBytes, err := app.Serializer().Marshal(msg)
			if err != nil {
				clog.Warnf("[Handler] Failed to marshal message for route=%s: %v", route, err)
				actor.ResponseCode(session, 500)
				return
			}
			var zero TReq
			req = &zero
			if err := app.Serializer().Unmarshal(msgBytes, req); err != nil {
				clog.Warnf("[Handler] Failed to unmarshal message for route=%s: %v", route, err)
				actor.ResponseCode(session, 500)
				return
			}
		}

		// 直接调用 handler，无需反射
		response, err := handler(session, req)
		if err != nil {
			clog.Warnf("[Handler] Handler returned error for route=%s: %v", route, err)
			// 检查是否是带错误码的错误
			if errWithCode, ok := err.(*ErrorWithCode); ok {
				actor.ResponseCode(session, errWithCode.Code)
			} else {
				actor.ResponseCode(session, 500)
			}
			return
		}
		if response != nil {
			actor.Response(session, response)
		}
	})
}

// RegisterHandlerNoReturn 注册无返回值的消息处理器（泛型版本）
func RegisterHandlerNoReturn[T any](actorType ActorType, route string, handler GenericHandlerFuncNoReturn[T]) {
	registerHandlerInternal(actorType, route, handler, func(actor *pomelo.ActorBase, session *cproto.Session, msg interface{}) {
		app := actor.App()
		if app == nil {
			clog.Warnf("[Handler] Actor app is nil for route=%s", route)
			actor.ResponseCode(session, 500)
			return
		}

		// 类型转换：将 interface{} 转换为 *T
		var req *T
		if msg == nil {
			var zero T
			req = &zero
		} else if msgPtr, ok := msg.(*T); ok {
			req = msgPtr
		} else {
			clog.Warnf("[Handler] Message type mismatch for route=%s: expected *%T, got %T, attempting conversion", route, (*T)(nil), msg)
			msgBytes, err := app.Serializer().Marshal(msg)
			if err != nil {
				clog.Warnf("[Handler] Failed to marshal message for route=%s: %v", route, err)
				actor.ResponseCode(session, 500)
				return
			}
			var zero T
			req = &zero
			if err := app.Serializer().Unmarshal(msgBytes, req); err != nil {
				clog.Warnf("[Handler] Failed to unmarshal message for route=%s: %v", route, err)
				actor.ResponseCode(session, 500)
				return
			}
		}

		// 直接调用 handler，无需反射
		handler(session, req)
	})
}

// registerHandlerInternal 内部注册函数
func registerHandlerInternal(actorType ActorType, route string, handler interface{}, callFunc func(actor *pomelo.ActorBase, session *cproto.Session, msg interface{})) {
	if !actorType.IsValid() {
		panic(fmt.Sprintf("handler: invalid actor type: %s", actorType))
	}
	if route == "" {
		panic("handler: route cannot be empty")
	}
	if handler == nil {
		panic("handler: handler cannot be nil")
	}

	msgHandlersV3Lock.Lock()
	defer msgHandlersV3Lock.Unlock()

	key := makeHandlerKey(actorType, route)
	if _, exists := msgHandlersV3[key]; exists {
		panic(fmt.Sprintf("handler: duplicate handler for actor=%s, route=%s", actorType, route))
	}

	msgHandlersV3[key] = &msgHandlerInfoV3{
		actorType: actorType,
		route:     route,
		callFunc:  callFunc,
	}

	clog.Infof("[Handler] Registered handler (generic, no reflection): actorType=%s, route=%s", actorType, route)
}

// registerHandlerRemote 内部函数：注册 Remote 处理器
// 优化：为 Remote 调用创建专门的 remoteCallFunc，直接调用 handler 并返回结果
func registerHandlerRemote[TReq any, TResp any](actorType ActorType, route string, handler GenericHandlerFuncRemote[TReq, TResp]) {
	if !actorType.IsValid() {
		panic(fmt.Sprintf("handler: invalid actor type: %s", actorType))
	}
	if route == "" {
		panic("handler: route cannot be empty")
	}

	msgHandlersV3Lock.Lock()
	defer msgHandlersV3Lock.Unlock()

	key := makeHandlerKey(actorType, route)
	if _, exists := msgHandlersV3[key]; exists {
		panic(fmt.Sprintf("handler: duplicate handler for actor=%s, route=%s", actorType, route))
	}

	// 为 Remote 调用创建专门的 remoteCallFunc
	// Remote 调用的 handler 签名是 func(req *TReq) (*TResp, int32)
	// remoteCallFunc 的签名是 func(req interface{}) (interface{}, int32)
	// 框架会通过反射处理返回值，支持 (resp, int32) 格式
	remoteCallFunc := func(req interface{}) (interface{}, int32) {
		// 类型转换：将 interface{} 转换为 *TReq
		var typedReq *TReq
		if req == nil {
			var zero TReq
			typedReq = &zero
		} else if reqPtr, ok := req.(*TReq); ok {
			typedReq = reqPtr
		} else {
			clog.Warnf("[Handler] Message type mismatch for route=%s: expected *%T, got %T", route, (*TReq)(nil), req)
			// 返回错误码
			return nil, 500
		}

		// 直接调用 Remote handler，返回 (resp, code)
		// 框架会通过反射正确处理返回值
		resp, code := handler(typedReq)
		return resp, code
	}

	msgHandlersV3[key] = &msgHandlerInfoV3{
		actorType:      actorType,
		route:          route,
		remoteCallFunc: remoteCallFunc,
	}

	clog.Infof("[Handler] Registered remote handler (generic, no reflection): actorType=%s, route=%s", actorType, route)
}

// RegisterAllToActorByTypeV3 将指定 Actor 类型的所有消息处理器注册到 actor（泛型版本）
// 优化：每种 actor 类型只遍历一次 handler 列表，后续实例直接使用缓存的列表
func RegisterAllToActorByTypeV3(actorType ActorType, actor *pomelo.ActorBase) {
	if !actorType.IsValid() {
		panic(fmt.Sprintf("handler: invalid actor type: %s", actorType))
	}

	// 获取 actor 的唯一标识
	actorPath := actor.PathString()
	actorKey := fmt.Sprintf("%s:%s", actorPath, actorType)

	// 检查该 actor 实例是否已经注册过
	registeredActorsLock.Lock()
	registeredRoutes, alreadyRegistered := registeredActors[actorKey]
	if !alreadyRegistered {
		registeredRoutes = make(map[string]bool)
		registeredActors[actorKey] = registeredRoutes
	}
	registeredActorsLock.Unlock()

	// 如果该 actor 实例已经注册过，直接返回
	if alreadyRegistered && len(registeredRoutes) > 0 {
		clog.Debugf("[Handler] Actor %s (type=%s) already registered handlers, skipping", actorPath, actorType)
		return
	}

	// 注意：即使 alreadyRegistered 为 false，registeredRoutes 也可能已经存在（但为空）
	// 我们需要确保 registeredRoutes 不为 nil，以便后续检查
	if registeredRoutes == nil {
		registeredActorsLock.Lock()
		registeredRoutes = make(map[string]bool)
		registeredActors[actorKey] = registeredRoutes
		registeredActorsLock.Unlock()
	}

	// 获取或构建该 actor 类型的 handler 列表（每种类型只遍历一次）
	handlerList := getOrBuildActorTypeHandlers(actorType)

	// 注册所有 handler 到该 actor 实例
	count := 0
	for _, info := range handlerList {
		// 检查该 route 是否已经注册到该 actor 实例（V3 版本）
		if registeredRoutes[info.route] {
			continue
		}
		registerToActorV3(actor, info.route, info)
		registeredRoutes[info.route] = true
		count++
	}

	if count > 0 {
		clog.Debugf("[Handler] Registered %d handler(s) to actor %s (type=%s, generic)", count, actorPath, actorType)
	}
}

// getOrBuildActorTypeHandlers 获取或构建指定 actor 类型的 handler 列表
// 优化：每种 actor 类型只遍历一次 handler 列表，后续直接使用缓存
func getOrBuildActorTypeHandlers(actorType ActorType) []*msgHandlerInfoV3 {
	actorTypeHandlersLock.Lock()
	defer actorTypeHandlersLock.Unlock()

	// 如果已缓存，直接返回
	if cached, exists := actorTypeHandlers[actorType]; exists {
		return cached
	}

	// 第一次：遍历所有 handler，找出属于该 actor 类型的 handler
	msgHandlersV3Lock.Lock()
	handlerList := make([]*msgHandlerInfoV3, 0)
	for _, info := range msgHandlersV3 {
		if info.actorType == actorType {
			handlerList = append(handlerList, info)
		}
	}
	msgHandlersV3Lock.Unlock()

	// 缓存该 actor 类型的 handler 列表
	actorTypeHandlers[actorType] = handlerList

	clog.Debugf("[Handler] Built handler cache for actor type=%s, count=%d", actorType, len(handlerList))
	return handlerList
}

// registerToActorV3 将单个消息处理器注册到 actor（泛型版本，无反射）
// 注意：如果 route 已经注册，cherry 框架会记录错误但不 panic，这里不重复检查
func registerToActorV3(actor *pomelo.ActorBase, route string, info *msgHandlerInfoV3) {
	// 直接使用闭包中保存的 callFunc，无需反射
	actor.Local().Register(route, func(session *cproto.Session, msg interface{}) {
		info.callFunc(actor, session, msg)
	})
}

// RegisterAllToActorByTypeV3Remote 将指定 Actor 类型的所有消息处理器注册到 actor 的 Remote（泛型版本）
// 用于 center 服务的 Remote 调用
// actor 参数应该是实现了 Remote() 方法的类型，如 *cactor.Base
func RegisterAllToActorByTypeV3Remote(actorType ActorType, actor interface{}) {
	if !actorType.IsValid() {
		panic(fmt.Sprintf("handler: invalid actor type: %s", actorType))
	}

	// 获取或构建该 actor 类型的 handler 列表（每种类型只遍历一次）
	handlerList := getOrBuildActorTypeHandlers(actorType)

	// 注册所有 handler 到该 actor 实例的 Remote
	count := 0
	for _, info := range handlerList {
		registerToActorV3Remote(actor, info.route, info)
		count++
	}

	if count > 0 {
		clog.Debugf("[Handler] Registered %d remote handler(s) to actor (type=%s, generic)", count, actorType)
	}
}

// registerToActorV3Remote 将单个消息处理器注册到 actor 的 Remote（泛型版本，无反射）
// 优化：优先使用 remoteCallFunc，直接处理 Remote 调用，无需适配
// actor 参数应该是实现了 Remote() 方法的类型，如 *cactor.Base
func registerToActorV3Remote(actor interface{}, route string, info *msgHandlerInfoV3) {
	// 类型断言获取 Remote 方法
	// 注意：这里假设 actor 是 *cactor.Base 类型
	actorBase, ok := actor.(*cactor.Base)
	if !ok {
		clog.Warnf("[Handler] Actor type assertion failed for route=%s, expected *cactor.Base", route)
		return
	}

	// 优先使用 remoteCallFunc（如果存在），这是专门为 Remote 调用优化的
	if info.remoteCallFunc != nil {
		actorBase.Remote().Register(route, info.remoteCallFunc)
		clog.Debugf("[Handler] Registered remote handler using remoteCallFunc: route=%s", route)
		return
	}

	// 如果没有 remoteCallFunc，回退到使用 callFunc（兼容旧代码）
	// 注意：这种情况不应该出现，因为 Remote 处理器应该使用 remoteCallFunc
	if info.callFunc != nil {
		actorBase.Remote().Register(route, func(req interface{}) interface{} {
			// 对于没有 remoteCallFunc 的情况，使用 callFunc 适配
			// 注意：这需要创建临时的 session，性能不如直接使用 remoteCallFunc
			// Remote 调用不需要 actor 和 session，传递 nil
			var nilSession *cproto.Session
			var nilActor *pomelo.ActorBase
			info.callFunc(nilActor, nilSession, req)
			// 返回 nil，因为 callFunc 不返回值
			return nil
		})
		clog.Debugf("[Handler] Registered remote handler using callFunc (fallback): route=%s", route)
		return
	}

	clog.Warnf("[Handler] No callFunc or remoteCallFunc found for route=%s", route)
}
