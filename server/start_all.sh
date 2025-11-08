#!/bin/bash

# 启动所有服务脚本
# 使用方法: ./start_all.sh

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 创建必要的目录
mkdir -p logs
mkdir -p bin

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  启动 Demo Cluster 所有服务${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查 NATS 服务器是否运行
if ! pgrep -f "nats-server" > /dev/null; then
    echo -e "${YELLOW}启动 NATS 服务器...${NC}"
    nats-server > /dev/null 2>&1 &
    sleep 2
    if pgrep -f "nats-server" > /dev/null; then
        echo -e "${GREEN}✓ NATS 服务器已启动${NC}"
    else
        echo -e "${RED}✗ NATS 服务器启动失败${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✓ NATS 服务器已在运行${NC}"
fi

# 检查并编译服务
echo -e "${YELLOW}检查服务可执行文件...${NC}"
services=("master" "center" "gate" "game" "web")
for service in "${services[@]}"; do
    if [ ! -f "bin/$service" ]; then
        echo -e "${YELLOW}编译 $service 服务...${NC}"
        go build -o "bin/$service" "./cmd/$service" || {
            echo -e "${RED}✗ $service 编译失败${NC}"
            exit 1
        }
    fi
done

# 检查并编译 robot_client
if [ ! -f "bin/robot_client" ]; then
    echo -e "${YELLOW}编译 robot_client...${NC}"
    go build -o "bin/robot_client" "./robot_client" || {
        echo -e "${RED}✗ robot_client 编译失败${NC}"
        exit 1
    }
fi

# 保存 PID 的文件
PID_FILE="$SCRIPT_DIR/.service_pids"

# 启动 master 服务
echo -e "${YELLOW}启动 master 服务...${NC}"
./bin/master > logs/master.log 2>&1 &
MASTER_PID=$!
echo "$MASTER_PID" > "$PID_FILE"
sleep 2
if ps -p $MASTER_PID > /dev/null; then
    echo -e "${GREEN}✓ Master 服务已启动 (PID: $MASTER_PID)${NC}"
else
    echo -e "${RED}✗ Master 服务启动失败${NC}"
    tail -10 logs/master.log
    exit 1
fi

# 启动 center 服务
echo -e "${YELLOW}启动 center 服务...${NC}"
./bin/center > logs/center.log 2>&1 &
CENTER_PID=$!
echo "$CENTER_PID" >> "$PID_FILE"
sleep 2
if ps -p $CENTER_PID > /dev/null; then
    echo -e "${GREEN}✓ Center 服务已启动 (PID: $CENTER_PID)${NC}"
else
    echo -e "${RED}✗ Center 服务启动失败${NC}"
    tail -10 logs/center.log
    exit 1
fi

# 启动 gate 服务
echo -e "${YELLOW}启动 gate 服务...${NC}"
./bin/gate > logs/gate.log 2>&1 &
GATE_PID=$!
echo "$GATE_PID" >> "$PID_FILE"
sleep 2
if ps -p $GATE_PID > /dev/null; then
    echo -e "${GREEN}✓ Gate 服务已启动 (PID: $GATE_PID)${NC}"
else
    echo -e "${RED}✗ Gate 服务启动失败${NC}"
    tail -10 logs/gate.log
    exit 1
fi

# 启动 game 服务
echo -e "${YELLOW}启动 game 服务...${NC}"
NODE_ID=10001 ./bin/game > logs/game.log 2>&1 &
GAME_PID=$!
echo "$GAME_PID" >> "$PID_FILE"
sleep 2
if ps -p $GAME_PID > /dev/null; then
    echo -e "${GREEN}✓ Game 服务已启动 (PID: $GAME_PID)${NC}"
else
    echo -e "${RED}✗ Game 服务启动失败${NC}"
    tail -10 logs/game.log
    exit 1
fi

# 启动 web 服务
echo -e "${YELLOW}启动 web 服务...${NC}"
./bin/web > logs/web.log 2>&1 &
WEB_PID=$!
echo "$WEB_PID" >> "$PID_FILE"
sleep 2
if ps -p $WEB_PID > /dev/null; then
    echo -e "${GREEN}✓ Web 服务已启动 (PID: $WEB_PID)${NC}"
else
    echo -e "${RED}✗ Web 服务启动失败${NC}"
    tail -10 logs/web.log
    exit 1
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  所有服务启动完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "服务 PID 列表:"
echo "  Master: $MASTER_PID"
echo "  Center: $CENTER_PID"
echo "  Gate:   $GATE_PID"
echo "  Game:   $GAME_PID"
echo "  Web:    $WEB_PID"
echo ""
echo "PID 文件保存在: $PID_FILE"
echo ""
echo "查看日志:"
echo "  tail -f logs/master.log"
echo "  tail -f logs/center.log"
echo "  tail -f logs/gate.log"
echo "  tail -f logs/game.log"
echo "  tail -f logs/web.log"
echo ""
echo "停止所有服务: ./stop_all.sh"

