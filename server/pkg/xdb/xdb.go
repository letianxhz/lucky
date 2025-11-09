package xdb

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

var ErrDup = errors.New("duplicated")

// TableOptions 表选项
type TableOptions struct {
	DaoKey       interface{}
	Concurrence  uint32        // 存储并发协程数
	SaveTimeout  time.Duration // 存储超时时间
	SyncInterval time.Duration // 存储队列刷新间隔
}

// DatabaseConfig 数据库配置接口（可选）
type DatabaseConfig interface {
	// InitializeDatabase 初始化数据库连接
	// 如果实现了此接口，会在驱动初始化之前调用
	InitializeDatabase() error
}

// Configurator 配置器接口
type Configurator interface {
	RedoOptions() *RedoOptions
	DriverOptions(driver string) interface{}
	DaoOptions(daoKey interface{}) interface{} // *storage_mongo.DaoOptions OR *storage_mysql.DaoOptions
	TableOptions(driver string, table string) *TableOptions
	// DryRun 返回true则数据的修改不入库，生产环境以及通常的开发和测试环境应该返回false
	DryRun() bool
}

var initialized bool

type daoCreation struct {
	dao     Dao
	driver  Driver
	daoOpts interface{}
}

// Setup 初始化 xdb
func Setup(ctx context.Context, c Configurator) (err error) {
	// 如果已经初始化，先清理所有 Source 的 repo
	if initialized {
		for _, src := range nsSrcMap {
			if src.repo != nil {
				src.repo = nil
			}
		}
		initialized = false
	}

	redoOptions = c.RedoOptions()

	// 如果配置器实现了 DatabaseConfig 接口，先初始化数据库
	if dbConfig, ok := c.(DatabaseConfig); ok {
		if err = dbConfig.InitializeDatabase(); err != nil {
			return errors.Wrap(err, "failed to initialize database")
		}
	}

	// collect all table, driver and dao creation info.
	options := map[*Source]*TableOptions{}
	drivers := map[Driver]string{}
	creations := map[interface{}]*daoCreation{}

	for _, src := range nsSrcMap {
		if src.DriverName == "none" {
			// 对于 none 驱动，使用默认选项
			opts := c.TableOptions(src.DriverName, src.Namespace)
			if opts == nil {
				opts = &TableOptions{
					DaoKey:       nil,
					Concurrence:  1,
					SaveTimeout:  5 * time.Second,
					SyncInterval: 100 * time.Millisecond,
				}
			}
			src.Init(opts, NoStorageTable{}, c.DryRun())
			if src.ModelType == nil {
				if src.repo == nil {
					src.repo = &Repo{}
				}
				src.repo.Init(true, &RepoOptions{GroupSize: 1}, src.Namespace, src.PKComparator, &waitGroup)
			}
			continue
		}

		opts := c.TableOptions(src.DriverName, src.Namespace)
		if opts == nil {
			return errors.Errorf("missing table options for driver: %s table src: %s", src.DriverName, src.Namespace)
		}
		if opts.DaoKey == nil {
			return errors.Errorf("invalid DaoKey for driver: %s table src: %s", src.DriverName, src.Namespace)
		}
		options[src] = opts

		daoOpts := c.DaoOptions(opts.DaoKey)
		if daoOpts == nil {
			return errors.Errorf("missing driver options DaoKey: %s for driver: %s table src: %s", opts.DaoKey, src.DriverName, src.Namespace)
		}

		driver := GetDriver(src.DriverName)
		if driver == nil {
			return errors.Errorf("invalid driver: %s table src: %s", src.DriverName, src.Namespace)
		}

		if _, ok := drivers[driver]; !ok {
			drivers[driver] = src.DriverName
		}

		creations[opts.DaoKey] = &daoCreation{
			driver:  driver,
			daoOpts: daoOpts,
		}
	}

	// init all drivers by user options.
	for driver, name := range drivers {
		if err = driver.Init(ctx, c.DriverOptions(name)); err != nil {
			return err
		}
	}

	// validate all sources.
	for _, src := range nsSrcMap {
		if src.DriverName == "none" {
			continue
		}
		driver := GetDriver(src.DriverName)
		err = driver.Validate(src)
		if err != nil {
			return err
		}
	}

	// create all daos.
	for _, creation := range creations {
		creation.dao, err = creation.driver.NewDao(ctx, creation.daoOpts)
		if err != nil {
			return err
		}
	}

	// init all tables and table source.
	for _, src := range nsSrcMap {
		if src.DriverName == "none" {
			src.Init(nil, NoStorageTable{}, true)
		} else {
			tableOpts := options[src]
			creation := creations[tableOpts.DaoKey]
			table := creation.dao.Table(src)
			src.Init(tableOpts, table, c.DryRun())
		}

		// src.ModelType 不为 null的repo在RegisterModel的时候已经初始化
		if src.ModelType == nil {
			if src.repo == nil {
				src.repo = &Repo{}
			}
			// 检查是否已经初始化，避免重复初始化
			// 如果已经初始化，先重置
			if src.repo.initialized {
				src.repo = &Repo{}
			}
			src.repo.Init(true, &RepoOptions{GroupSize: 1}, src.Namespace, src.PKComparator, &waitGroup)
		}
	}

	// recover
	for _, src := range nsSrcMap {
		src.Recover(ctx, &waitGroup)
	}

	waitGroup.Wait()

	// run saver
	for _, src := range nsSrcMap {
		src.RunSavers(ctx, &waitGroup)
	}

	initialized = true

	return nil
}

// Create 创建记录
func Create[T MutableRecord](ctx context.Context, proto interface{}) (T, error) {
	return CreateS[T](ctx, getTypeSource[T](), proto)
}

// CreateNS 根据命名空间创建记录
func CreateNS[T MutableRecord](ctx context.Context, ns string, proto interface{}) (T, error) {
	var empty T
	src, err := getNSSource(ns)
	if err != nil {
		return empty, err
	}

	return CreateS[T](ctx, src, proto)
}

// CreateS 根据源创建记录
func CreateS[T MutableRecord](ctx context.Context, src *Source, proto interface{}) (T, error) {
	return createS[T](ctx, src, proto, false, false)
}

func createS[T MutableRecord](ctx context.Context, src *Source, proto interface{}, tmp bool, mirror bool) (T, error) {
	if src.TicketExpected != nil && src.TicketExpected(proto) {
		return create[T](ctx, src, proto, tmp, mirror)
	}

	pk := src.PKOf(proto)
	if !SafeMode(ctx) {
		v, err := GetS[T](ctx, src, pk)
		if err != nil {
			return v, err
		}

		if !isNil(v) {
			return v, errors.Wrapf(ErrDup, "duplicated [%s]: %s", src.Namespace, pk)
		}
	}

	for {
		ret, err := create[T](ctx, src, proto, tmp, mirror)
		if err == nil || !errors.Is(err, ErrDup) {
			return ret, err
		}

		orig, _ := src.repo.Get(pk)
		if orig != nil {
			return orig.(T), err
		}
	}
}

func create[T MutableRecord](ctx context.Context, src *Source, proto interface{}, tmp bool, mirror bool) (T, error) {
	legalType := src.GetLegalType()
	// 如果是指针类型，获取指向的类型
	if legalType.Kind() == reflect.Ptr {
		legalType = legalType.Elem()
	}
	obj := reflect.New(legalType).Interface()

	ret := obj.(T)
	header := ret.GetHeader()
	header.SetTmp(tmp)
	if mirror {
		header.EnableMirror()
	}

	err := ret.Init(ctx, proto)
	if err != nil {
		var empty T
		return empty, err
	}

	return ret, nil
}

// Get 获取记录
func Get[T Record](ctx context.Context, args ...any) (T, error) {
	src := getTypeSource[T]()
	pk, err := src.PKCreator(args)
	if err != nil {
		return zero[T](), err
	}
	return GetS[T](ctx, src, pk)
}

// GetNS 根据命名空间获取记录
func GetNS[T any](ctx context.Context, ns string, args []any) (T, error) {
	var empty T
	src, err := getNSSource(ns)
	if err != nil {
		return empty, err
	}

	pk, err := src.PKCreator(args)
	if err != nil {
		return empty, err
	}

	return GetS[T](ctx, src, pk)
}

// GetS 根据源获取记录
func GetS[T any](ctx context.Context, s *Source, pk PK) (T, error) {
	var ret any
	var err error

	// 如果注册了 Model 类型，总是返回 Model 类型（以注册的 Model 为准）
	if s.ModelType != nil {
		ret, err = getModel(ctx, s, pk)
	} else {
		ret, err = getNonModel(ctx, s, pk)
	}

	if ret == nil || err != nil {
		return zero[T](), err
	}

	// 类型转换处理
	var zeroT T
	retType := reflect.TypeOf(ret)
	targetType := reflect.TypeOf(zeroT)

	// 如果返回类型和请求类型相同，直接返回
	if retType == targetType {
		return ret.(T), nil
	}

	// 如果注册了 Model 类型，返回的总是 Model 类型
	// 无论用户请求的是 *msg.PlayerRecord 还是 *PlayerModel，都返回 *PlayerModel
	// 但由于类型系统限制，*PlayerModel 不能直接转换为 *msg.PlayerRecord
	// 如果用户请求的是 Record 类型，但注册了 Model，提示用户使用 Model 类型
	if s.ModelType != nil {
		// 返回的是 Model 类型
		if targetType == s.ModelType {
			// 用户请求的就是 Model 类型，直接返回
			return ret.(T), nil
		} else if targetType == reflect.PtrTo(s.RecordType) {
			// 用户请求的是 Record 类型，但注册了 Model
			// 由于类型系统限制，不能直接转换，提示用户使用 Model 类型
			panic(fmt.Sprintf("registered Model type is *%s, cannot return as *%s. Use xdb.Get[*%s] instead",
				s.ModelType.Elem().Name(), s.RecordType.Name(), s.ModelType.Elem().Name()))
		}
	}

	// 尝试直接类型断言
	return ret.(T), nil
}

func getNonModel(ctx context.Context, src *Source, pk PK) (any, error) {
	curr, err := src.Table().Fetch(ctx, true, pk)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = curr.Close(ctx)
	}()

	if !curr.Next(ctx) {
		return nil, nil
	}

	val := reflect.New(src.RecordType)
	err = curr.Decode(val.Interface())
	if err != nil {
		return nil, err
	}

	// 设置记录状态为已加载
	if record, ok := val.Interface().(Record); ok {
		header := record.GetHeader()
		header.LoadComplete()
		// 设置生命周期为 Normal（从数据库读取的记录应该是 Normal 状态）
		header.Init(LifecycleNormal)
	}

	// 返回指针类型
	return val.Interface(), nil
}

func getModel(ctx context.Context, src *Source, pk PK) (any, error) {
	ret, ok := src.repo.Get(pk)
	if ok {
		return ret, nil
	}

	curr, err := src.Table().Fetch(ctx, true, pk)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = curr.Close(ctx)
	}()

	if !curr.Next(ctx) {
		return src.repo.SetOnFetch(pk, nil, true, nil), nil
	}

	// src.ModelType 是指针类型（如 *PlayerModel），需要获取元素类型
	var modelElemType reflect.Type
	if src.ModelType.Kind() == reflect.Ptr {
		modelElemType = src.ModelType.Elem()
	} else {
		modelElemType = src.ModelType
	}
	val := reflect.New(modelElemType)
	err = curr.Decode(val.Interface())
	if err != nil {
		return nil, err
	}

	// val 是指向 PlayerModel 的指针，需要转换为 *PlayerModel
	ret = val.Interface()
	var volatile bool
	if vol, ok := ret.(Volatile); ok {
		volatile = vol.IsVolatile()
	}

	return src.repo.SetOnFetch(pk, ret, volatile, func(obj any) {
		m := obj.(Model)
		// 设置记录状态为已加载
		header := m.GetHeader()
		header.LoadComplete()
		// 设置生命周期为 Normal（从数据库读取的记录应该是 Normal 状态）
		header.Init(LifecycleNormal)
		if src.Replica {
			m.OnReload(ctx)
		} else {
			m.OnLoad(ctx)
		}
	}), nil
}

// Save 保存记录
func Save(ctx context.Context, v MutableRecord) {
	suc, lif, changes := save(ctx, v, false)

	if suc {
		// 可以在这里触发保存事件
		_ = lif
		_ = changes
	}
}

func save(ctx context.Context, v MutableRecord, locked bool) (bool, Lifecycle, FieldSet) {
	src := v.Source()
	if m, ok := v.(Model); ok {
		// 在 actor 模型中不需要锁，直接检查过期状态
		if m.GetHeader().IsExpired() {
			return false, 0, 0
		}

		if !m.GetHeader().IsMirror() && m.Lifecycle() == LifecycleNormal {
			var volatile bool
			if vol, ok := m.(Volatile); ok {
				volatile = vol.IsVolatile()
			}

			if !src.repo.SetOnStore(src.PKOf(v), m, volatile) {
				panic(errors.Wrapf(ErrDup, "duplicated [%s]: %s", src.Namespace, PKOf(m)))
			}
		}
	}

	if !v.Dirty() {
		return false, 0, 0
	}

	lif := v.Lifecycle()
	c, changes := v.Commit(ctx)

	if c != nil && src.saver != nil {
		si := src.saver.Put(ctx, c, v.SavingIndex())
		if si >= -1 {
			v.SetSavingIndex(si)
		}
	}

	return true, lif, changes
}

// Sync 同步等待数据入库
func Sync(ctx context.Context, v Record) error {
	src := v.Source()
	if src == nil {
		return errors.Errorf("unregistered RecordSource for %s", reflect.TypeOf(v).Name())
	}

	Save(ctx, v.(MutableRecord))
	if src.saver != nil {
		src.saver.Sync(PKOf(v))
	}
	return nil
}

// Stop 停止 xdb
func Stop(ctx context.Context) {
	for _, src := range nsSrcMap {
		if src.saver != nil {
			src.saver.Close()
		}
	}

	waitGroup.Wait()
}

// 辅助函数
func getTypeSource[T Record]() *Source {
	var zero T
	t := reflect.TypeOf(zero)
	// 如果是指针类型，获取指向的类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 先尝试直接查找
	src := GetSourceByType(t)
	if src == nil {
		// 如果是 Model 类型，尝试查找嵌入的 Record 类型
		if t.Kind() == reflect.Struct {
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				if f.Anonymous {
					// 尝试查找嵌入的 Record 类型
					src = GetSourceByType(f.Type)
					if src == nil {
						src = GetSourceByType(reflect.PtrTo(f.Type))
					}
					if src != nil {
						// 检查是否是 Model 类型
						if src.ModelType != nil && src.ModelType == reflect.PtrTo(t) {
							break
						}
						// 如果 ModelType 不匹配，继续查找
						if src.ModelType == nil {
							// 可能还没有注册 Model，使用 Record 类型
							break
						}
					}
				}
			}
		}
	}

	if src == nil {
		panic(fmt.Sprintf("unregistered type: %s", t.Name()))
	}

	// 如果注册了 Model 类型，允许使用 Record 类型来查找，但返回的 Source 会使用 ModelType
	// 这样 xdb.Get[*msg.PlayerRecord] 可以返回 *PlayerModel
	if src.GetLegalType() != t {
		// 检查是否是 Record 类型，但注册了 Model 类型
		if src.ModelType != nil && src.RecordType == t {
			// 允许使用 Record 类型查找，但实际返回 Model 类型
			return src
		}
		// 如果是 Model 类型，允许使用 ModelType
		if src.ModelType != nil && src.ModelType == reflect.PtrTo(t) {
			return src
		}
		panic(errors.New(fmt.Sprintf("illegal type: %s", t)))
	}

	return src
}

func getNSSource(ns string) (*Source, error) {
	src := GetSourceByNS(ns)
	if src == nil {
		return nil, errors.New(fmt.Sprintf("unregistered namespace: %s", ns))
	}

	return src, nil
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

func zero[T any]() T {
	var zero T
	return zero
}
