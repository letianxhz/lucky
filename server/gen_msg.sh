#!/bin/bash

# 生成 msg 脚本
# 从 pkg/protocol 目录的 .proto 文件生成 Go 代码到 gen/msg 目录

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== 生成 msg 代码（从 pkg/protocol）==="
echo ""

# 创建输出目录
MSG_DIR="$SCRIPT_DIR/gen/msg"
mkdir -p "$MSG_DIR"

# 查找所有 .proto 文件（在 pkg/protocol 目录下）
PROTO_DIR="$SCRIPT_DIR/pkg/protocol"
PROTO_FILES=$(find "$PROTO_DIR" -maxdepth 1 -name "*.proto" -type f 2>/dev/null)

if [ -z "$PROTO_FILES" ]; then
    echo "⚠ 未找到 .proto 文件在 $PROTO_DIR"
    exit 0
fi

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo "错误: protoc 未安装"
    echo "请安装 Protocol Buffers 编译器:"
    echo "  brew install protobuf  # macOS"
    exit 1
fi

# 检查 protoc-gen-go 插件
if ! command -v protoc-gen-go &> /dev/null; then
    echo "错误: protoc-gen-go 未安装"
    echo "请安装: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

echo "处理协议文件..."
for proto_file in $PROTO_FILES; do
    proto_name=$(basename "$proto_file" .proto)
    
    echo "  处理: $proto_name.proto"
    
    # 生成 proto Go 代码到 gen/msg 目录
    protoc \
      --go_out="$MSG_DIR" \
      --go_opt=paths=source_relative \
      --proto_path="$PROTO_DIR" \
      "$proto_file"
    
    echo "  ✓ ${proto_name}.pb.go 生成成功"
done

# 移动生成的文件到 msg 目录（如果生成到了子目录）
if [ -d "$MSG_DIR/lucky" ]; then
    find "$MSG_DIR/lucky" -name "*.pb.go" -exec mv {} "$MSG_DIR/" \; 2>/dev/null || true
    rm -rf "$MSG_DIR/lucky" 2>/dev/null || true
fi

# 修改生成的代码包名为 msg
for f in "$MSG_DIR"/*.pb.go; do
    if [ -f "$f" ]; then
        # 替换 package 为 msg（包括 pb、protocol、main 等）
        sed -i '' -e 's/^package pb$/package msg/' -e 's/^package protocol$/package msg/' -e 's/^package main$/package msg/' "$f" 2>/dev/null || \
        sed -i -e 's/^package pb$/package msg/' -e 's/^package protocol$/package msg/' -e 's/^package main$/package msg/' "$f"
    fi
done

echo ""
echo "✓ msg 代码生成完成"
echo "  输出目录: $MSG_DIR"
echo ""
echo "生成的文件:"
ls -lh "$MSG_DIR"/*.pb.go 2>/dev/null | awk '{print "  " $9}' || echo "  无 .pb.go 文件"
echo ""
