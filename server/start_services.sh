#!/bin/bash

# 启动 NATS 服务器（如果未运行）
if ! pgrep -f "nats-server" > /dev/null; then
    echo "启动 NATS 服务器..."
    nats-server > /dev/null 2>&1 &
    sleep 2
fi

# 启动 master 服务
echo "启动 master 服务..."
./bin/master > logs/master.log 2>&1 &
MASTER_PID=$!
sleep 2

# 启动 center 服务
echo "启动 center 服务..."
./bin/center > logs/center.log 2>&1 &
CENTER_PID=$!
sleep 2

# 启动 gate 服务
echo "启动 gate 服务..."
./bin/gate > logs/gate.log 2>&1 &
GATE_PID=$!
sleep 2

# 启动 game 服务
echo "启动 game 服务..."
NODE_ID=10001 ./bin/game > logs/game.log 2>&1 &
GAME_PID=$!
sleep 2

# 启动 web 服务（robot 需要）
echo "启动 web 服务..."
./bin/web > logs/web.log 2>&1 &
WEB_PID=$!
sleep 2

echo "所有服务已启动"
echo "Master PID: $MASTER_PID"
echo "Center PID: $CENTER_PID"
echo "Gate PID: $GATE_PID"
echo "Game PID: $GAME_PID"
echo "Web PID: $WEB_PID"
echo ""
echo "停止服务: kill $MASTER_PID $CENTER_PID $GATE_PID $GAME_PID $WEB_PID"
