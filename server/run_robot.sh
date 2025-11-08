#!/bin/bash

# 运行 Robot 测试脚本
# 使用方法: ./run_robot.sh [robot数量]

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查 robot_client 是否存在
if [ ! -f "bin/robot_client" ]; then
    echo -e "${YELLOW}编译 robot_client...${NC}"
    go build -o "bin/robot_client" "./robot_client" || {
        echo -e "${RED}✗ robot_client 编译失败${NC}"
        exit 1
    }
fi

# 检查服务是否运行
echo -e "${YELLOW}检查服务状态...${NC}"
if ! pgrep -f "./bin/web" > /dev/null; then
    echo -e "${RED}✗ Web 服务未运行，请先运行 ./start_all.sh${NC}"
    exit 1
fi

if ! pgrep -f "./bin/gate" > /dev/null; then
    echo -e "${RED}✗ Gate 服务未运行，请先运行 ./start_all.sh${NC}"
    exit 1
fi

if ! pgrep -f "./bin/game" > /dev/null; then
    echo -e "${RED}✗ Game 服务未运行，请先运行 ./start_all.sh${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 所有必需服务正在运行${NC}"
echo ""

# 运行 robot_client
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  启动 Robot 测试${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

./bin/robot_client

