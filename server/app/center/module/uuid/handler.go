package uuid

import (
	"context"

	clog "github.com/cherry-game/cherry/logger"
	"lucky/server/gen/msg"
	"lucky/server/pkg/di"
	"lucky/server/pkg/handler"
)

// init 注册 UUID 模块的消息处理器
func init() {
	var h = &uuidHandler{}
	di.Register(h)

	// 使用统一的 RegisterHandlerRemote，所有处理器都返回 (*Response, int32)，无反射，高性能
	handler.RegisterHandlerRemote(handler.ActorTypeUuid, "allocateUUID", h.OnAllocateUUID)
}

type uuidHandler struct {
	uuidModule IUuidModule `di:"auto"`
}

// OnAllocateUUID 分配 UUID 消息处理器（Remote 调用）
// 签名: func(req *msg.String) (*msg.UuidRange, int32)
func (h *uuidHandler) OnAllocateUUID(req *msg.String) (*msg.UuidRange, int32) {
	if req == nil {
		clog.Warn("[OnAllocateUUID] request is nil")
		return nil, GetErrorCode(ErrInvalidName)
	}

	ctx := context.Background()
	name := req.Value

	// 调用模块处理业务逻辑
	range_, err := h.uuidModule.AllocateUUID(ctx, name)
	if err != nil {
		clog.Errorf("[OnAllocateUUID] allocate uuid failed: %v", err)
		return nil, GetErrorCode(err)
	}

	return range_, 0 // 0 表示成功
}
