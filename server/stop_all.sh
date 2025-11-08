#!/bin/bash

# 停止所有服务脚本
# 使用方法: ./stop_all.sh

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  停止 Demo Cluster 所有服务${NC}"
echo -e "${YELLOW}========================================${NC}"

# PID 文件
PID_FILE="$SCRIPT_DIR/.service_pids"

# 停止通过 PID 文件记录的服务
if [ -f "$PID_FILE" ]; then
    echo -e "${YELLOW}从 PID 文件停止服务...${NC}"
    while read pid; do
        if [ -n "$pid" ] && ps -p "$pid" > /dev/null 2>&1; then
            echo -e "${YELLOW}停止进程 PID: $pid${NC}"
            kill "$pid" 2>/dev/null || true
        fi
    done < "$PID_FILE"
    rm -f "$PID_FILE"
    sleep 1
fi

# 停止所有通过 bin/ 启动的服务
echo -e "${YELLOW}查找并停止所有服务进程...${NC}"
services=("master" "center" "gate" "game" "web")
stopped_count=0

for service in "${services[@]}"; do
    pids=$(pgrep -f "./bin/$service" || true)
    if [ -n "$pids" ]; then
        for pid in $pids; do
            echo -e "${YELLOW}停止 $service 服务 (PID: $pid)...${NC}"
            kill "$pid" 2>/dev/null || true
            stopped_count=$((stopped_count + 1))
        done
    fi
done

# 等待进程退出
if [ $stopped_count -gt 0 ]; then
    echo -e "${YELLOW}等待进程退出...${NC}"
    sleep 2
    
    # 强制杀死仍在运行的进程
    for service in "${services[@]}"; do
        pids=$(pgrep -f "./bin/$service" || true)
        if [ -n "$pids" ]; then
            for pid in $pids; do
                echo -e "${RED}强制停止 $service 服务 (PID: $pid)...${NC}"
                kill -9 "$pid" 2>/dev/null || true
            done
        fi
    done
fi

# 检查是否还有服务在运行
remaining=$(pgrep -f "./bin/(master|center|gate|game|web)" | wc -l | tr -d ' ')
if [ "$remaining" -gt 0 ]; then
    echo -e "${RED}警告: 仍有 $remaining 个服务进程在运行${NC}"
    pgrep -f "./bin/(master|center|gate|game|web)"
else
    echo -e "${GREEN}✓ 所有服务已停止${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  停止完成${NC}"
echo -e "${GREEN}========================================${NC}"

