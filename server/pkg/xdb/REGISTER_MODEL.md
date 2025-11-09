# RegisterModel 使用说明

## 概述

`RegisterModel` 函数用于注册模型类型到 xdb 系统。模型类型必须实现 `Model` 接口，并且嵌入一个已注册的 `Record` 类型。

## 函数签名

```go
func RegisterModel[T Model](repoOpts *RepoOptions)
```

## 参数说明

- `T`: 必须是实现了 `Model` 接口的类型
- `repoOpts`: 仓库选项，可以为 `nil`（使用默认选项）

## Model 接口要求

模型类型必须实现以下接口：

```go
type Model interface {
    Record
    Listener
    Locker() RWLocker
    LockPriority() int
    ValidateAffinity() bool
}
```

## 使用示例

### 1. 定义模型结构

```go
type PlayerModel struct {
    mu sync.RWMutex
    PlayerRecord  // 嵌入已注册的 Record 类型
}

// 实现 Model 接口
func (m *PlayerModel) Locker() xdb.RWLocker {
    return &m.mu
}

func (m *PlayerModel) LockPriority() int {
    return int(m.PlayerId())
}

func (m *PlayerModel) ValidateAffinity() bool {
    return m.GetHeader().ValidateAffinity()
}

// 实现 Listener 接口（空实现）
func (m *PlayerModel) OnCreate(ctx context.Context) {}
func (m *PlayerModel) OnLoad(ctx context.Context) {}
func (m *PlayerModel) OnUpdate(ctx context.Context, fs xdb.FieldSet) {}
func (m *PlayerModel) OnDelete(ctx context.Context) {}
func (m *PlayerModel) OnReload(ctx context.Context, locked bool) {}
func (m *PlayerModel) OnRefresh(ctx context.Context) {}
```

### 2. 注册模型

在 `init()` 函数中注册模型：

```go
func init() {
    xdb.RegisterModel[*PlayerModel](nil)  // 使用默认选项
    // 或
    xdb.RegisterModel[*PlayerModel](&xdb.RepoOptions{GroupSize: 16})  // 自定义选项
}
```

## 注意事项

1. **注册时机**: `RegisterModel` 必须在 `xdb.Setup()` 之前调用，通常在 `init()` 函数中调用
2. **Record 类型**: 嵌入的 `Record` 类型必须已经通过 `RegisterSource` 注册（通常由代码生成器自动完成）
3. **唯一性**: 每个 `Record` 类型只能注册一个 `Model` 类型
4. **Repo 初始化**: 如果 `Model` 类型已注册，`xdb.Setup()` 时不会再次初始化 repo

## 实现细节

- `RegisterModel` 会查找嵌入的 `Record` 类型
- 设置 `ModelType` 和 `EmbeddedOffset`
- 初始化 `Repo`（如果还没有初始化）
- 注册模型类型到 `typeSrcMap`
- 调用驱动的 `ExtendType` 方法（如果驱动已注册）

## 错误处理

如果出现以下情况会 panic：

1. 模型类型没有嵌入已注册的 `Record` 类型
2. 同一个 `Record` 类型注册了多个 `Model` 类型

