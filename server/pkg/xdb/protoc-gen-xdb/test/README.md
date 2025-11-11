# protoc-gen-xdb 测试示例

这个目录包含了 protoc-gen-xdb 的完整测试示例。

## 文件说明

- `player.proto`: 测试用的 proto 文件，包含 Player 和 Item 两个 message
- `generate.sh`: 代码生成脚本
- `example_test.go`: 使用示例和测试代码
- `README.md`: 本文件

## 快速开始

### 1. 生成代码

```bash
cd lucky/server/pkg/xdb/protoc-gen-xdb/test
./generate.sh
```

这将生成：
- `player_xdb.pb.go`: 包含 Player 和 Item 的 xdb 代码

### 2. 运行测试

```bash
go test -v
```

或者运行示例：

```bash
go test -run ExamplePlayerRecord -v
go test -run ExampleItemRecord -v
```

## 测试内容

### Player 示例

展示了基本的 CRUD 操作：
- 创建玩家记录
- 获取玩家记录
- 更新玩家记录
- 同步保存

### Item 示例

展示了复合主键的使用：
- 创建道具记录（player_id + item_id）
- 获取道具记录
- 更新道具数量

### 字段常量测试

验证字段常量是否正确生成。

### PK 创建测试

验证主键创建功能。

### Source 注册测试

验证数据源是否正确注册。

## 注意事项

1. **需要先运行 generate.sh**: 在运行测试之前，必须先生成代码
2. **需要 xdb 模块**: 确保 xdb 模块已正确实现
3. **需要数据库驱动**: 实际使用时需要配置数据库驱动

## 预期输出

运行 `go test -v` 应该看到类似输出：

```
=== RUN   TestFieldConstants
Field constants:
  FieldPlayerId: 0
  FieldName: 1
  FieldLevel: 2
  FieldExp: 3
--- PASS: TestFieldConstants (0.00s)

=== RUN   TestPKCreation
Created PK: player:1001
--- PASS: TestPKCreation (0.00s)

=== RUN   TestSourceRegistration
Source registered: player
--- PASS: TestSourceRegistration (0.00s)
```

## 故障排除

### 问题：找不到 protoc-gen-xdb

**解决**: 确保已运行 `build.sh` 构建工具：

```bash
cd lucky/server/pkg/xdb/protoc-gen-xdb
./build.sh
```

### 问题：找不到 extension.proto

**解决**: 确保 `generate.sh` 中的 `--proto_path` 正确指向 xdb 目录。

### 问题：生成的代码编译错误

**解决**: 
1. 检查 proto 文件语法
2. 确保所有必需的字段都已定义
3. 确保 xdb 模块已正确实现



