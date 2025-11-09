# 快速修复：undefined: SetupXdb

## 问题
IDE 只编译了 `main.go` 文件，而没有包含 `config.go` 文件（其中定义了 `SetupXdb` 函数）。

## 解决方案

### 方案 1: 使用包模式运行（推荐）

在 GoLand 中：
1. 右键点击 `example` 目录
2. 选择 "Run 'go build example'" 或 "Run 'go run example'"
3. 或者使用运行配置 "xdb example"（包模式）

### 方案 2: 使用命令行

```bash
cd lucky/server/pkg/xdb/example
go run .
```

或者指定所有文件：
```bash
go run main.go config.go player.pb.go player_xdb.pb.go
```

### 方案 3: 使用构建脚本

```bash
cd lucky/server/pkg/xdb/example
./build.sh
./xdb_example
```

### 方案 4: 修改 IDE 运行配置

如果 IDE 仍然只编译单个文件，请：
1. Run → Edit Configurations
2. 选择你的运行配置
3. 确保 "Kind" 设置为 "Package" 而不是 "File"
4. 或者使用 "xdb example (single file)" 配置，它明确指定了所有需要的文件

## 验证

运行以下命令验证编译是否成功：

```bash
cd lucky/server/pkg/xdb/example
go build -o /tmp/xdb_test main.go config.go player.pb.go player_xdb.pb.go
```

如果成功，说明所有文件都能正确编译。

