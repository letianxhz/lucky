package xdb

import (
	"context"

	"github.com/pkg/errors"
)

// MustInitialize 初始化 xdb（类似 orm.MustInitialize）
// 如果失败则 panic
func MustInitialize(ctx context.Context, c Configurator) {
	if err := Setup(ctx, c); err != nil {
		panic(errors.Wrap(err, "failed to initialize xdb"))
	}
}
