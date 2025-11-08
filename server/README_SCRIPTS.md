# 服务管理脚本说明

## 脚本列表

### 1. start_all.sh
启动所有服务（master, center, gate, game, web）

**使用方法:**
```bash
./start_all.sh
```

**功能:**
- 检查并启动 NATS 服务器
- 检查并编译所有服务（如果可执行文件不存在）
- 按顺序启动所有服务
- 保存服务 PID 到 `.service_pids` 文件
- 检查服务启动状态

### 2. stop_all.sh
停止所有服务

**使用方法:**
```bash
./stop_all.sh
```

**功能:**
- 从 PID 文件停止服务
- 查找并停止所有运行中的服务进程
- 强制停止仍在运行的进程
- 清理 PID 文件

### 3. status.sh
查看服务运行状态

**使用方法:**
```bash
./status.sh
```

**功能:**
- 显示所有服务的运行状态
- 显示端口监听状态
- 显示 PID 文件中的进程状态

## 服务端口

- **Master**: 40011
- **Center**: 30011
- **Gate**: 10011 (TCP), 10010 (WebSocket)
- **Game**: 20011
- **Web**: 8081
- **NATS**: 4222

## 日志文件

所有服务的日志保存在 `logs/` 目录：
- `logs/master.log`
- `logs/center.log`
- `logs/gate.log`
- `logs/game.log`
- `logs/web.log`

## 使用示例

### 启动所有服务
```bash
./start_all.sh
```

### 查看服务状态
```bash
./status.sh
```

### 停止所有服务
```bash
./stop_all.sh
```

### 查看实时日志
```bash
# 查看所有日志
tail -f logs/*.log

# 查看特定服务日志
tail -f logs/gate.log
```

## 注意事项

1. 确保在 `server` 目录下运行脚本
2. 确保已安装并配置好 NATS 服务器
3. 确保端口未被其他程序占用
4. 停止服务时会先尝试正常停止，如果失败会强制停止（kill -9）

