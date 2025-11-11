#!/bin/bash

# 购买道具压测脚本
# 使用方法: ./run_benchmark.sh [机器人数量] [每个机器人请求数]

cd "$(dirname "$0")/.."

ROBOT_COUNT=${1:-100}      # 默认100个机器人
REQUESTS_PER_ROBOT=${2:-10}  # 默认每个机器人10个请求

echo "========== 购买道具压测 =========="
echo "机器人数量: $ROBOT_COUNT"
echo "每个机器人请求数: $REQUESTS_PER_ROBOT"
echo "总请求数: $((ROBOT_COUNT * REQUESTS_PER_ROBOT))"
echo ""
echo "请确保以下服务已启动："
echo "  - Web 服务 (端口 8081)"
echo "  - Gate 服务 (端口 10011)"
echo "  - Game 服务 (节点 10001)"
echo ""
echo "按 Enter 继续，或 Ctrl+C 取消..."
read

# 修改代码中的压测参数
sed -i '' "s/benchmarkMode     = .*/benchmarkMode     = true/" robot_client/main.go
sed -i '' "s/benchmarkRobots   = .*/benchmarkRobots   = $ROBOT_COUNT/" robot_client/main.go
sed -i '' "s/benchmarkRequests = .*/benchmarkRequests = $REQUESTS_PER_ROBOT/" robot_client/main.go

# 编译
echo "编译压测程序..."
go build -o bin/robot_benchmark ./robot_client 2>&1

if [ $? -ne 0 ]; then
    echo "编译失败"
    exit 1
fi

# 运行压测
echo "开始压测..."
./bin/robot_benchmark

# 恢复代码
sed -i '' "s/benchmarkMode     = .*/benchmarkMode     = false/" robot_client/main.go

echo ""
echo "压测完成！"





