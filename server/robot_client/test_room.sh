#!/bin/bash

# Room 功能测试脚本

# 定义输出文件名
OUTPUT_NAME="robot_test_room"

# 编译机器人客户端
echo "Building robot client for room test..."
cd "$(dirname "$0")/.." || exit 1

# 临时修改 main.go 中的测试标志
sed -i.bak 's/testRoom.*=.*false/testRoom                = true                   \/\/ 是否运行 Room 功能测试/' robot_client/main.go
sed -i.bak 's/printLog.*=.*false/printLog                = true                    \/\/ 是否输出详细日志/' robot_client/main.go || sed -i.bak 's/printLog.*=.*true/printLog                = true                    \/\/ 是否输出详细日志/' robot_client/main.go

go build -o bin/${OUTPUT_NAME} ./robot_client 2>&1

# 恢复 main.go
if [ -f robot_client/main.go.bak ]; then
    mv robot_client/main.go.bak robot_client/main.go
fi

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful. Running test..."

# 运行机器人客户端
./bin/${OUTPUT_NAME} 2>&1 | tee /tmp/${OUTPUT_NAME}.log

echo ""
echo "Test finished. Check /tmp/${OUTPUT_NAME}.log for full logs."

