package xdb

// 这个文件包含使用示例，实际使用时应该删除或移到测试文件

/*
示例：如何定义和使用 xdb

1. 定义 Proto 类型（对应数据库表结构）
type PlayerProto struct {
    PlayerId int64 `json:"player_id"`
    Name     string `json:"name"`
    Level    int32 `json:"level"`
}

2. 定义 Record 类型
type PlayerRecord struct {
    Header
    PlayerProto
}

func (r *PlayerRecord) Source() *Source {
    return GetSourceByNS("player")
}

func (r *PlayerRecord) XId() string {
    return fmt.Sprintf("player:%d", r.PlayerId)
}

func (r *PlayerRecord) Lifecycle() Lifecycle {
    return r.Header.Lifecycle()
}

func (r *PlayerRecord) Snapshoot() interface{} {
    return &r.PlayerProto
}

func (r *PlayerRecord) XVersion() int64 {
    return 0
}

func (r *PlayerRecord) MarshalJSON() ([]byte, error) {
    return json.Marshal(r.PlayerProto)
}

func (r *PlayerRecord) UnmarshalJSON(data []byte) error {
    return json.Unmarshal(data, &r.PlayerProto)
}

func (r *PlayerRecord) String() string {
    return fmt.Sprintf("Player{Id:%d, Name:%s}", r.PlayerId, r.Name)
}

func (r *PlayerRecord) GetHeader() *Header {
    return &r.Header
}

3. 实现 MutableRecord 接口
func (r *PlayerRecord) Init(ctx context.Context, data interface{}) error {
    proto := data.(*PlayerProto)
    r.PlayerProto = *proto
    r.Header.Init(LifecycleNew)
    return nil
}

func (r *PlayerRecord) Update(ctx context.Context, changes interface{}, fs FieldSet) error {
    proto := changes.(*PlayerProto)
    // 根据 FieldSet 更新字段
    if fs.Contains(FieldName) {
        r.Name = proto.Name
    }
    if fs.Contains(FieldLevel) {
        r.Level = proto.Level
    }
    return nil
}

func (r *PlayerRecord) Delete(ctx context.Context) bool {
    return r.Header.MarkAsDeleted(ctx)
}

func (r *PlayerRecord) Commit(ctx context.Context) (Commitment, FieldSet) {
    // 创建提交对象
    commitment := r.CreateCommitment()
    changes := r.Header.Changes()
    return commitment, changes
}

func (r *PlayerRecord) Committing() bool {
    return r.Header.Committing()
}

func (r *PlayerRecord) Dirty() bool {
    return r.Header.Dirty()
}

func (r *PlayerRecord) SavingIndex() int32 {
    return r.Header.SavingIndex
}

func (r *PlayerRecord) SetSavingIndex(idx int32) {
    r.Header.SavingIndex = idx
}

4. 定义主键类型
type PlayerPK struct {
    PlayerId int64
}

func (pk *PlayerPK) Source() *Source {
    return GetSourceByNS("player")
}

func (pk *PlayerPK) String() string {
    return fmt.Sprintf("player:%d", pk.PlayerId)
}

func (pk *PlayerPK) HashGroup() int {
    return int(pk.PlayerId % 16)
}

func (pk *PlayerPK) Empty() bool {
    return pk.PlayerId == 0
}

func (pk *PlayerPK) PrefixOf(key Key) bool {
    other, ok := key.(*PlayerPK)
    if !ok {
        return false
    }
    // 实现前缀匹配逻辑
    return true
}

func (pk *PlayerPK) Full() bool {
    return pk.PlayerId > 0
}

func (pk *PlayerPK) FetchFilter() interface{} {
    return map[string]interface{}{
        "player_id": pk.PlayerId,
    }
}

5. 注册数据源
func init() {
    src := &Source{
        ProtoType:  reflect.TypeOf((*PlayerProto)(nil)).Elem(),
        RecordType: reflect.TypeOf((*PlayerRecord)(nil)).Elem(),
        PKType:     reflect.TypeOf((*PlayerPK)(nil)).Elem(),
        Namespace:  "player",
        DriverName: "mysql",
        DBName:     "game_db",
        TableName:  "player",
        KeySize:    1,
        PKCreator: func(args []interface{}) (PK, error) {
            if len(args) < 1 {
                return nil, fmt.Errorf("invalid args")
            }
            return &PlayerPK{PlayerId: args[0].(int64)}, nil
        },
        PKOf: func(obj interface{}) PK {
            r := obj.(*PlayerRecord)
            return &PlayerPK{PlayerId: r.PlayerId}
        },
        PKComparator: func(a, b interface{}) int {
            pk1 := a.(*PlayerPK)
            pk2 := b.(*PlayerPK)
            if pk1.PlayerId < pk2.PlayerId {
                return -1
            } else if pk1.PlayerId > pk2.PlayerId {
                return 1
            }
            return 0
        },
        CreateCommitment: func() Commitment {
            // 返回具体的 Commitment 实现
            return &PlayerCommitment{}
        },
    }
    RegisterSource(src)
}

6. 使用
func main() {
    ctx := context.Background()

    // 初始化
    configurator := &MyConfigurator{}
    err := Setup(ctx, configurator)
    if err != nil {
        panic(err)
    }

    // 创建
    player, err := Create[PlayerRecord](ctx, &PlayerProto{
        PlayerId: 1001,
        Name:     "Test",
        Level:    1,
    })

    // 获取
    player, err := Get[PlayerRecord](ctx, int64(1001))

    // 更新
    player.Name = "NewName"
    player.GetHeader().SetChanged(FieldName)
    Save(ctx, player)

    // 同步
    Sync(ctx, player)
}
*/
