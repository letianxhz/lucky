#!/bin/bash

# 购买道具测试脚本
# 使用方法: ./run_buy_item_test.sh

cd "$(dirname "$0")/.."

echo "========== 编译购买道具测试程序 =========="
go build -o bin/robot_buy_item_test -ldflags "-X main.testBuyItem=true -X main.printLog=true" ./robot_client 2>&1

if [ $? -ne 0 ]; then
    echo "编译失败"
    exit 1
fi

echo "========== 运行购买道具测试 =========="
echo "请确保以下服务已启动："
echo "  - Web 服务 (端口 8081)"
echo "  - Gate 服务 (端口 10011)"
echo "  - Game 服务 (节点 10001)"
echo ""
echo "按 Enter 继续，或 Ctrl+C 取消..."
read

./bin/robot_buy_item_test





