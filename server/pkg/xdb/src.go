package xdb

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

var typeSrcMap = make(map[reflect.Type]*Source)
var nsSrcMap = make(map[string]*Source)
var waitGroup sync.WaitGroup

// RegisterSource 注册数据源
func RegisterSource(src *Source) {
	typeSrcMap[src.RecordType] = src
	typeSrcMap[src.ProtoType] = src
	typeSrcMap[src.PKType] = src

	if _, ok := nsSrcMap[src.Namespace]; ok {
		panic(fmt.Sprintf("namespace : %s has registed", src.Namespace))
	}
	nsSrcMap[src.Namespace] = src
}

// SourceOf 获取对象的源
func SourceOf(obj interface{}) *Source {
	if r, ok := obj.(Record); ok {
		return r.Source()
	}

	return GetSourceByType(reflect.TypeOf(obj))
}

// GetSource 根据类型获取源
func GetSource[T Record]() *Source {
	return GetSourceByType(reflect.TypeOf((*T)(nil)).Elem())
}

// GetSourceByType 根据类型获取源
func GetSourceByType(tPtr reflect.Type) *Source {
	return typeSrcMap[tPtr]
}

// GetSourceByNS 根据命名空间获取源
func GetSourceByNS(ns string) *Source {
	return nsSrcMap[ns]
}

// Namespace 获取命名空间
func Namespace[T Record]() (string, bool) {
	src := GetSourceByType(reflect.TypeOf((*T)(nil)).Elem())
	if src == nil {
		return "", false
	}

	return src.Namespace, true
}

// Sources 获取所有源
func Sources() []*Source {
	ret := make([]*Source, 0, len(nsSrcMap))
	for _, src := range nsSrcMap {
		ret = append(ret, src)
	}
	return ret
}

// RegisterModel 注册模型类型
// T 必须是 Model 类型，并且嵌入了一个 Record 类型
// repoOpts 为 nil 时使用默认选项
func RegisterModel[T Model](repoOpts *RepoOptions) {
	registerModel[T](repoOpts)
}

func registerModel[T Model](repoOpts *RepoOptions) {
	// 获取模型类型
	var zero T
	modelType := reflect.TypeOf(zero)

	// 获取结构体类型（用于查找字段）
	var tStruct reflect.Type
	if modelType.Kind() == reflect.Ptr {
		tStruct = modelType.Elem()
	} else {
		tStruct = modelType
		// 如果 T 不是指针类型，保存指向结构体的指针类型
		modelType = reflect.PtrTo(modelType)
	}

	// 确保是结构体类型
	if tStruct.Kind() != reflect.Struct {
		panic(fmt.Sprintf("registerModel: model type %s is not a struct type", tStruct.Name()))
	}

	// 查找嵌入的 Record 类型
	var src *Source
	var rField reflect.StructField
	for i, n := 0, tStruct.NumField(); i < n; i++ {
		f := tStruct.Field(i)
		if !f.Anonymous {
			continue
		}

		// 检查是否是已注册的 Record 类型
		// 先尝试指针类型
		src = typeSrcMap[reflect.PtrTo(f.Type)]
		if src == nil {
			// 再尝试值类型
			src = typeSrcMap[f.Type]
		}
		if src != nil {
			rField = f
			break
		}
	}

	if src == nil {
		panic(fmt.Sprintf("registerModel: model type %s does not embed a registered Record type", tStruct.Name()))
	}

	// 检查是否已经注册过
	if src.ModelType != nil && src.ModelType != modelType {
		panic(fmt.Sprintf("Registered model type %s for record type %s", src.ModelType.Name(), src.RecordType.Name()))
	}

	// 设置模型类型和嵌入偏移
	src.ModelType = modelType
	src.EmbeddedOffset = rField.Offset

	// 初始化 repo（如果还没有初始化）
	if src.repo == nil {
		src.repo = &Repo{}
	}

	// 检查是否是 Volatile 类型
	volatileType := reflect.TypeOf((*Volatile)(nil)).Elem()
	canonical := !modelType.Implements(volatileType)

	// 设置默认选项
	opts := repoOpts
	if opts == nil {
		opts = &RepoOptions{GroupSize: 1}
	}

	// 初始化 repo
	src.repo.Init(canonical, opts, src.Namespace, src.PKComparator, &waitGroup)

	// 注册模型类型到 typeSrcMap
	typeSrcMap[modelType] = src

	// 调用驱动的 ExtendType（如果驱动已注册）
	driver := GetDriver(src.DriverName)
	if driver != nil {
		driver.ExtendType(src.RecordType, src.ModelType)
	}
}

// FieldDesc 字段描述
type FieldDesc struct {
	Code Field
	Name string
	Key  bool
}

// PK 主键接口
type PK interface {
	SourceInterface
	fmt.Stringer
	Key
	Full() bool
	FetchFilter() interface{}
}

// Key 键接口（用于缓存）
type Key interface {
	HashGroup() int
	Empty() bool
	PrefixOf(key Key) bool
}

// Source 数据源
type Source struct {
	ProtoType        reflect.Type
	RecordType       reflect.Type
	ModelType        reflect.Type
	PKType           reflect.Type
	EmbeddedOffset   uintptr
	DriverName       string
	DBName           string
	TableName        string
	KeySize          int
	PKComparator     func(interface{}, interface{}) int
	PKCreator        func([]interface{}) (PK, error)
	PKOf             func(interface{}) PK
	Part             func(interface{}, FieldSet) (interface{}, FieldSet)
	TicketExpected   func(message interface{}) bool
	CreateCommitment func() Commitment
	Projection       interface{}
	Options          interface{}
	Namespace        string
	Replica          bool // 同一个数据源的对象在其他服务中修改，需要自动通知
	FieldSetSave     FieldSet
	Fields           []*FieldDesc
	table            Table
	repo             *Repo
	saver            *Saver
}

// GetLegalType 获取合法类型
func (src *Source) GetLegalType() reflect.Type {
	if src.ModelType != nil {
		return src.ModelType
	}
	return src.RecordType
}

// Init 初始化源
func (src *Source) Init(opts *TableOptions, table Table, dryRun bool) {
	src.table = table
	if !dryRun {
		src.saver = NewSaver(src, opts.Concurrence, opts.SaveTimeout, opts.SyncInterval)
	}
}

// Recover 恢复数据
func (src *Source) Recover(ctx context.Context, wg *sync.WaitGroup) {
	if src.saver != nil {
		src.saver.Recover(ctx, wg)
	}
}

// RunSavers 运行保存器
func (src *Source) RunSavers(ctx context.Context, wg *sync.WaitGroup) {
	if src.saver != nil {
		src.saver.Run(ctx, wg)
	}
}

// Table 获取表接口
func (src *Source) Table() Table {
	return src.table
}

// NoStorage 检查是否无存储
func (src *Source) NoStorage() bool {
	var table NoStorageTable
	return src.table == table
}

// Validate 验证源
func (src *Source) Validate() error {
	driver := GetDriver(src.DriverName)
	if driver == nil {
		return fmt.Errorf("invalid driver: %s table src: %s", src.DriverName, src.Namespace)
	}
	return driver.Validate(src)
}

// FieldNames 获取字段名列表
func (src *Source) FieldNames(withRuntimeField bool) []string {
	names := make([]string, 0, len(src.Fields))
	for _, field := range src.Fields {
		if withRuntimeField || src.FieldSetSave.Contains(field.Code) {
			names = append(names, field.Name)
		}
	}
	return names
}
