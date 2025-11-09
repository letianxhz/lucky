package mongo

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"lucky/server/pkg/xdb"
)

// DriverOptions MongoDB 驱动选项
type DriverOptions struct {
	// 可以添加嵌套配置等
}

// DaoOptions MongoDB DAO 选项
type DaoOptions struct {
	URI             string // MongoDB 连接 URI
	PoolSize        int32  // 连接池大小
	TransformDBName func(string) string
}

var driver = &Driver{}

func init() {
	xdb.RegisterDriver(driver)
}

// Driver MongoDB 驱动
type Driver struct{}

func (d *Driver) Name() string {
	return "mongo"
}

func (d *Driver) Init(ctx context.Context, config interface{}) error {
	// MongoDB 驱动初始化逻辑
	// 可以在这里处理驱动级别的配置
	return nil
}

func (d *Driver) Validate(src *xdb.Source) error {
	if src.DriverName != d.Name() {
		return errors.Errorf("driver name not match: %s against to %s in %s", src.DriverName, d.Name(), src.Namespace)
	}
	// 可以添加更多验证逻辑
	return nil
}

func (d *Driver) NewDao(ctx context.Context, option interface{}) (xdb.Dao, error) {
	opts, ok := option.(*DaoOptions)
	if !ok {
		return nil, errors.Errorf("invalid %s options", d.Name())
	}

	if opts.URI == "" {
		return nil, errors.New("MongoDB URI is required")
	}

	// 创建 MongoDB 客户端
	clientOptions := options.Client().ApplyURI(opts.URI)
	if opts.PoolSize > 0 {
		clientOptions.SetMaxPoolSize(uint64(opts.PoolSize))
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to MongoDB")
	}

	// 测试连接
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping MongoDB")
	}

	return newDao(client, opts.TransformDBName), nil
}

func (d *Driver) ExtendType(base reflect.Type, extension reflect.Type) {
	// 类型扩展逻辑
}
