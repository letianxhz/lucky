#!/bin/bash

# 测试 Player Actor 和 Room Actor 之间的通信

cd "$(dirname "$0")/.."

echo "========== 编译测试程序 =========="
go build -o bin/test_actor_comm ./robot_client/test_actor_communication.go ./robot_client/*.go 2>&1 | head -20

if [ $? -ne 0 ]; then
    echo "编译失败"
    exit 1
fi

echo "========== 运行 Actor 通信测试 =========="
./bin/test_actor_comm 2>&1 | tee /tmp/test_actor_comm.log

echo ""
echo "========== 测试完成 =========="
echo "日志文件: /tmp/test_actor_comm.log"



