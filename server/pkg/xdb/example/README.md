# xdb 完整功能示例

这是一个完整的 xdb 模块使用示例，展示了如何：
1. 定义 proto 文件
2. 生成 xdb 代码
3. 配置和初始化 xdb
4. 使用 CRUD 操作

## 文件结构

```
example/
├── player.proto          # Proto 定义文件
├── generate.sh           # 代码生成脚本
├── build.sh              # 构建脚本
├── config.go             # 配置器实现
├── main.go               # 主程序
├── main_test.go          # 测试文件
├── player.pb.go          # Proto 生成的 Go 代码（生成）
├── player_xdb.pb.go      # xdb 生成的代码（生成）
└── README.md             # 本文档
```

## 快速开始

### 1. 生成代码

```bash
./generate.sh
```

这会生成：
- `player.pb.go` - Proto 生成的 Go 代码
- `player_xdb.pb.go` - xdb 生成的代码

### 2. 构建示例

**方式一：使用构建脚本**
```bash
./build.sh
```

**方式二：直接使用 go build**
```bash
go build -o xdb_example main.go config.go player.pb.go player_xdb.pb.go
```

**方式三：使用 go run（推荐用于开发）**
```bash
go run main.go config.go player.pb.go player_xdb.pb.go
```

### 3. 运行测试

```bash
cd ../../  # 回到 lucky/server 目录
go test -v ./pkg/xdb/example
```

## 功能说明

### Proto 定义

`player.proto` 定义了：
- **Player**: 玩家数据表（单主键）
- **Item**: 道具表（复合主键）

### 配置器

`config.go` 实现了 `xdb.Configurator` 接口：
- `RedoOptions()`: 重做日志配置
- `DriverOptions()`: 驱动配置
- `DaoOptions()`: DAO 配置
- `TableOptions()`: 表配置
- `DryRun()`: 是否干运行（测试模式）

### 主程序

`main.go` 演示了：
1. 初始化 xdb 模块
2. 检查 Source 注册
3. 测试 PK 创建
4. 展示 CRUD 操作流程

## 注意事项

1. **编译要求**: 必须同时编译所有相关文件：
   - `main.go`
   - `config.go`
   - `player.pb.go`
   - `player_xdb.pb.go`

2. **驱动类型**: 示例中使用 `DRIVER_NONE` 进行测试，不会实际保存到数据库

3. **DryRun 模式**: 配置为 `true`，数据修改不会入库

4. **代码生成**: 需要先运行 `generate.sh` 生成代码

## IDE 配置

如果使用 IDE（如 GoLand），需要确保：
1. 将所有 `.go` 文件添加到构建配置中
2. 或者使用 `go run` 命令运行整个包

## 下一步

要使用真实的数据库驱动：
1. 修改 `player.proto` 中的 `driver` 选项为 `DRIVER_MYSQL` 或其他驱动
2. 在配置器中实现真实的数据库连接配置
3. 设置 `DryRun()` 返回 `false`
