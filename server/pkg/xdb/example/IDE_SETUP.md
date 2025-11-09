# IDE 配置说明

## GoLand / IntelliJ IDEA 配置

### 问题
如果 IDE 只编译 `main.go` 文件，会出现 `undefined: SetupXdb` 错误。

### 解决方案

#### 方案 1: 使用包模式运行（推荐）

1. 在 GoLand 中，右键点击 `example` 目录
2. 选择 "Run 'go build example'" 或 "Run 'go run example'"
3. 或者创建运行配置：
   - Run → Edit Configurations
   - 添加 "Go Build"
   - Package path: `lucky/server/pkg/xdb/example`
   - Working directory: `$PROJECT_DIR$/lucky/server/pkg/xdb/example`

#### 方案 2: 使用构建脚本

直接使用提供的构建脚本：
```bash
./build.sh
```

#### 方案 3: 手动指定所有文件

在运行配置中，指定所有需要编译的文件：
- `main.go`
- `config.go`
- `player.pb.go`
- `player_xdb.pb.go`

#### 方案 4: 使用命令行

在终端中运行：
```bash
cd lucky/server/pkg/xdb/example
go run main.go config.go player.pb.go player_xdb.pb.go
```

或者：
```bash
cd lucky/server
go run ./pkg/xdb/example
```

## VS Code 配置

在 `.vscode/settings.json` 中添加：
```json
{
  "go.buildTags": "",
  "go.testFlags": ["-v"],
  "go.buildFlags": []
}
```

运行方式：
- 使用终端：`go run main.go config.go player.pb.go player_xdb.pb.go`
- 或使用包模式：`go run .`

