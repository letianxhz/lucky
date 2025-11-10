package uuid

import (
	"context"
	"lucky/server/app/center/entity"
	"time"

	clog "github.com/cherry-game/cherry/logger"
	"lucky/server/gen/db/center"
	"lucky/server/gen/msg"
	"lucky/server/pkg/di"
	"lucky/server/pkg/xdb"
)

const (
	// UUIDBatchSize 每次分配的 UUID 数量
	UUIDBatchSize = 1024
	// UUIDName 默认 UUID 名称
	UUIDName = "default"
)

// UuidModule UUID 模块实现
// 负责 UUID 相关的业务逻辑
type UuidModule struct {
	// 可以注入依赖，如缓存、数据库等
	// cache uuid.IUuidCache `di:"auto"`
	// db    *db.DB          `di:"auto"`
}

// init 初始化 UUID 模块并注册到 di 容器
func init() {
	var v = &UuidModule{}
	di.Register(v)
	di.RegisterImplementation((*IUuidModule)(nil), v)
}

// AllocateUUID 分配 UUID 范围（每次 1024 个）
func (m *UuidModule) AllocateUUID(ctx context.Context, name string) (*msg.UuidRange, error) {
	if name == "" {
		name = UUIDName
	}

	// 获取或创建 UUID 记录
	pk := &center.UuidPK{Name: name}
	record, err := xdb.Get[*entity.UUIDEntity](ctx, pk)
	if err != nil {
		clog.Errorf("[UuidModule] get uuid record failed: %v", err)
		return nil, ErrDBError
	}

	// 如果记录不存在，创建新记录
	if record == nil {
		now := time.Now().Unix()
		newProto := &center.Uuid{
			Name:  name,
			Value: 0, // 初始值为 0
			Ctime: now,
			Mtime: now,
		}

		record, err = xdb.Create[*entity.UUIDEntity](ctx, newProto)
		if err != nil {
			clog.Errorf("[UuidModule] create uuid record failed: %v", err)
			return nil, ErrDBError
		}
	}

	// 计算新的值范围
	startValue := record.Value + 1
	endValue := record.Value + UUIDBatchSize

	// 更新记录
	record.Value = endValue
	record.Mtime = time.Now().Unix()

	// 标记字段变更
	record.GetHeader().SetChanged(center.UuidFieldValue, center.UuidFieldMtime)

	// 保存到数据库
	xdb.Save(ctx, record)

	clog.Infof("[UuidModule] allocated UUID range: name=%s, start=%d, end=%d", name, startValue, endValue)

	return &msg.UuidRange{
		Start: startValue,
		End:   endValue,
	}, nil
}
