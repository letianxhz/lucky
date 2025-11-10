package entity

import (
	"context"
	"lucky/server/gen/db/center"
	"lucky/server/pkg/xdb"
)

func init() {
	xdb.RegisterModel[*UUIDEntity](nil)
}

type UUIDEntity struct {
	center.UuidRecord
}

// OnCreate 创建时的回调
func (U *UUIDEntity) OnCreate(ctx context.Context) {
	// UUID 创建时不需要特殊处理
}

// OnLoad 加载时的回调
func (U *UUIDEntity) OnLoad(ctx context.Context) {
	// UUID 加载时不需要特殊处理
}

// OnUpdate 更新时的回调
func (U *UUIDEntity) OnUpdate(ctx context.Context, fs xdb.FieldSet) {
	// UUID 更新时不需要特殊处理
	// 可以在这里添加日志、事件通知等
}

// OnDelete 删除时的回调
func (U *UUIDEntity) OnDelete(ctx context.Context) {
	// UUID 删除时不需要特殊处理
}

// OnReload 重新加载时的回调
func (U *UUIDEntity) OnReload(ctx context.Context) {
	// UUID 重新加载时不需要特殊处理
}

// OnRefresh 刷新时的回调
func (U *UUIDEntity) OnRefresh(ctx context.Context) {
	// UUID 刷新时不需要特殊处理
}
