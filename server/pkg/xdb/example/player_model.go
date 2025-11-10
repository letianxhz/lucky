package main

import (
	"context"

	"lucky/server/pkg/xdb"
)

func init() {
	xdb.RegisterModel[*PlayerModel](nil)
}

// PlayerModel 玩家模型
// 注意：在 actor 模型中，不需要锁，因为每个 actor 在单线程环境中运行
type PlayerModel struct {
	msg.PlayerRecord
}

// 实现 xdb.Model 接口

// ValidateAffinity 验证亲和性
func (m *PlayerModel) ValidateAffinity() bool {
	return m.GetHeader().ValidateAffinity()
}

// 实现 Listener 接口（空实现，可根据需要添加逻辑）

func (m *PlayerModel) OnCreate(ctx context.Context) {
	// 创建时的回调
}

func (m *PlayerModel) OnLoad(ctx context.Context) {
	// 加载时的回调
}

func (m *PlayerModel) OnUpdate(ctx context.Context, fs xdb.FieldSet) {
	// 更新时的回调
}

func (m *PlayerModel) OnDelete(ctx context.Context) {
	// 删除时的回调
}

func (m *PlayerModel) OnReload(ctx context.Context) {
	// 重新加载时的回调
}

func (m *PlayerModel) OnRefresh(ctx context.Context) {
	// 刷新时的回调
}
