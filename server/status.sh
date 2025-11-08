#!/bin/bash

# 查看服务状态脚本
# 使用方法: ./status.sh

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Demo Cluster 服务状态${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查服务进程
services=("master" "center" "gate" "game" "web")
total_running=0

for service in "${services[@]}"; do
    pids=$(pgrep -f "./bin/$service" || true)
    if [ -n "$pids" ]; then
        for pid in $pids; do
            status=$(ps -p "$pid" -o stat= 2>/dev/null || echo "")
            if [ -n "$status" ]; then
                echo -e "${GREEN}✓${NC} $service (PID: $pid) - 运行中"
                total_running=$((total_running + 1))
            fi
        done
    else
        echo -e "${RED}✗${NC} $service - 未运行"
    fi
done

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "运行中的服务: ${total_running}/${#services[@]}"
echo -e "${BLUE}========================================${NC}"

# 检查端口监听
echo ""
echo -e "${BLUE}端口监听状态:${NC}"
ports=("8081:Web" "10011:Gate" "20011:Game" "30011:Center" "40011:Master" "4222:NATS")
for port_info in "${ports[@]}"; do
    port="${port_info%%:*}"
    name="${port_info##*:}"
    if lsof -i ":$port" > /dev/null 2>&1; then
        listener=$(lsof -i ":$port" | grep LISTEN | head -1 | awk '{print $1, $2}' || echo "")
        echo -e "${GREEN}✓${NC} $name (端口 $port) - $listener"
    else
        echo -e "${RED}✗${NC} $name (端口 $port) - 未监听"
    fi
done

# 检查 PID 文件
PID_FILE="$SCRIPT_DIR/.service_pids"
if [ -f "$PID_FILE" ]; then
    echo ""
    echo -e "${BLUE}PID 文件: $PID_FILE${NC}"
    cat "$PID_FILE" | while read pid; do
        if [ -n "$pid" ]; then
            if ps -p "$pid" > /dev/null 2>&1; then
                echo -e "  PID $pid - ${GREEN}运行中${NC}"
            else
                echo -e "  PID $pid - ${RED}已停止${NC}"
            fi
        fi
    done
fi

echo ""

