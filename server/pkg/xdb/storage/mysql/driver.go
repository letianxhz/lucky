package mysql

import (
	"context"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"lucky/server/pkg/xdb"
)

// DriverOptions MySQL 驱动选项
type DriverOptions struct {
	// 可以添加驱动级别的配置
}

// DaoOptions MySQL DAO 选项
type DaoOptions struct {
	DBName       string        // 数据库名
	Host         string        // 主机地址
	Port         int32         // 端口
	Username     string        // 用户名
	Password     string        // 密码
	Charset      string        // 字符集，默认 utf8mb4
	MaxOpenConns int32         // 最大打开连接数
	QueryTimeout time.Duration // 查询超时时间
}

var driver = &Driver{}

func init() {
	xdb.RegisterDriver(driver)
}

// Driver MySQL 驱动
type Driver struct{}

func (d *Driver) Name() string {
	return "mysql"
}

func (d *Driver) Init(ctx context.Context, config interface{}) error {
	// MySQL 驱动初始化逻辑
	return nil
}

func (d *Driver) Validate(src *xdb.Source) error {
	if src.DriverName != d.Name() {
		return errors.Errorf("driver name not match: %s against to %s in %s", src.DriverName, d.Name(), src.Namespace)
	}
	// MySQL 不需要 DBName（使用 DaoOptions 中的 DBName）
	return nil
}

func (d *Driver) NewDao(ctx context.Context, option interface{}) (xdb.Dao, error) {
	opts, ok := option.(*DaoOptions)
	if !ok {
		return nil, errors.Errorf("invalid %s options", d.Name())
	}

	return newDao(ctx, opts)
}

func (d *Driver) ExtendType(base reflect.Type, extension reflect.Type) {
	// 类型扩展逻辑（如果需要）
}
