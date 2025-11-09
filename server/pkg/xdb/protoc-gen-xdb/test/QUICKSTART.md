# 快速开始 - protoc-gen-xdb 测试示例

这是一个完整的测试示例，展示如何使用 protoc-gen-xdb 生成代码并使用。

## 目录结构

```
test/
├── player.proto          # 测试用的 proto 文件
├── generate.sh           # 代码生成脚本
├── run_example.sh        # 运行完整示例的脚本
├── example.go            # 使用示例程序
├── example_test.go       # 测试代码
├── go.mod                # Go 模块文件
├── README.md             # 详细说明
└── QUICKSTART.md         # 本文件
```

## 快速开始（3 步）

### 步骤 1: 构建工具

```bash
cd lucky/server/pkg/xdb/protoc-gen-xdb
./build.sh
```

### 步骤 2: 生成代码

```bash
cd test
./generate.sh
```

这将生成 `player_xdb.pb.go` 文件。

### 步骤 3: 运行示例

```bash
./run_example.sh
```

或者直接运行示例程序：

```bash
go run example.go
```

## 详细步骤说明

### 1. 查看 Proto 文件

`player.proto` 定义了两个 message：

- **Player**: 玩家表，单主键（player_id）
- **Item**: 道具表，复合主键（player_id + item_id）

### 2. 生成代码

运行 `generate.sh` 会：

1. 检查并构建 `protoc-gen-xdb` 工具
2. 使用 `protoc` 生成 `player_xdb.pb.go`
3. 显示生成的文件

### 3. 查看生成的代码

生成的文件包含：

- **字段常量**: `FieldPlayerId`, `FieldName`, `FieldLevel`, etc.
- **PK 结构体**: `PlayerPK`, `ItemPK`
- **Record 结构体**: `PlayerRecord`, `ItemRecord`
- **Commitment 结构体**: `PlayerCommitment`, `ItemCommitment`
- **Source 配置**: `_PlayerSource`, `_ItemSource`
- **初始化代码**: `init()` 函数自动注册 Source

### 4. 使用生成的代码

`example.go` 展示了：

- ✅ 创建记录
- ✅ 获取记录
- ✅ 更新记录
- ✅ 同步保存
- ✅ 复合主键使用
- ✅ 字段常量使用
- ✅ Source 信息查询

## 预期输出

运行 `go run example.go` 应该看到：

```
=== protoc-gen-xdb 使用示例 ===

1. 创建玩家记录
   ✓ 创建成功: player:1001
   玩家信息: ID=1001, Name=TestPlayer, Level=1

2. 获取玩家记录
   ✓ 获取成功: player:1001
   玩家信息: Name=TestPlayer, Level=1, Exp=0

3. 更新玩家记录
   ✓ 更新成功
   变更: Name TestPlayer -> UpdatedPlayer, Level 1 -> 10

4. 同步保存
   ✓ 同步成功

5. 创建道具记录（复合主键）
   ✓ 创建成功
   道具信息: PlayerID=1001, ItemID=2001, Count=10

6. 获取道具记录
   ✓ 获取成功
   道具信息: Count=10

7. 更新道具数量
   ✓ 更新成功
   变更: Count 10 -> 20

8. 字段常量
   FieldPlayerId: 0
   FieldName: 1
   FieldLevel: 2
   FieldExp: 3

9. Source 信息
   ✓ Source 已注册
   Namespace: player
   TableName: player
   DriverName: mysql

=== 示例完成 ===
```

## 常见问题

### Q: protoc 未找到

**A**: 安装 Protocol Buffers 编译器：

```bash
# macOS
brew install protobuf

# Ubuntu/Debian
apt-get install protobuf-compiler

# 验证安装
protoc --version
```

### Q: 生成的代码编译错误

**A**: 确保：

1. xdb 模块已正确实现
2. proto 文件语法正确
3. 所有必需的字段都已定义

### Q: 找不到 extension.proto

**A**: 确保 `generate.sh` 中的 `--proto_path` 正确：

```bash
--proto_path="$XDB_DIR"
```

### Q: 运行时错误

**A**: 这个示例主要展示代码生成和使用流程。实际数据库操作需要：

1. 配置数据库驱动
2. 初始化 xdb
3. 设置数据库连接

## 下一步

1. **查看生成的代码**: 打开 `player_xdb.pb.go` 了解生成的结构
2. **修改 proto 文件**: 添加更多字段或 message
3. **集成到项目**: 在实际项目中使用生成的代码
4. **配置数据库**: 设置真实的数据库连接

## 相关文档

- `README.md`: 详细使用说明
- `../README.md`: protoc-gen-xdb 工具说明
- `../USAGE.md`: 完整使用指南

