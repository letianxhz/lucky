package uuid

import (
	"context"
	"lucky/server/gen/msg"
)

// IUuidModule UUID 模块接口
// 定义 UUID 相关的所有业务操作
type IUuidModule interface {
	// AllocateUUID 分配 UUID 范围（每次 1024 个）
	AllocateUUID(ctx context.Context, name string) (*msg.UuidRange, error)
}

// 错误定义
var (
	ErrInvalidName = &UuidError{Code: 1, Message: "invalid uuid name"}
	ErrDBError     = &UuidError{Code: 2, Message: "database error"}
)

// UuidError UUID 错误
type UuidError struct {
	Code    int32
	Message string
}

func (e *UuidError) Error() string {
	return e.Message
}

// GetErrorCode 将模块错误转换为错误码
func GetErrorCode(err error) int32 {
	if err == nil {
		return 0 // ccode.OK
	}

	if uuidErr, ok := err.(*UuidError); ok {
		return uuidErr.Code
	}

	return 2 // ccode.DBError
}
