#!/bin/bash
# 运行 xdb 示例程序

cd "$(dirname "$0")"

PB_DIR="pb"

# 检查生成的文件
if [ ! -f "$PB_DIR/player.pb.go" ] || [ ! -f "$PB_DIR/player_xdb.pb.go" ]; then
    echo "错误: 缺少生成的代码文件"
    echo "请先运行 ./generate.sh"
    exit 1
fi

# 运行
go run main.go config.go "$PB_DIR/player.pb.go" "$PB_DIR/player_xdb.pb.go"
