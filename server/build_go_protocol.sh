#!/bin/bash

# 编译 Go protocol 文件脚本
# 使用方法: ./build_go_protocol.sh

cd "$(dirname "$0")"

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo "错误: protoc 未安装"
    echo "请运行: brew install protobuf"
    exit 1
fi

# 检查 protoc-gen-go 是否安装
if ! command -v protoc-gen-go &> /dev/null; then
    echo "错误: protoc-gen-go 未安装"
    echo "请运行: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    echo "并确保 $HOME/go/bin 在 PATH 中"
    exit 1
fi

echo "开始编译 protocol 文件..."
echo "⚠ 注意: 现在 protocol 文件应该使用 gen_msg.sh 生成到 gen/msg 目录"
echo "   此脚本已废弃，请使用: ./gen_msg.sh"
echo ""
echo "如需继续使用此脚本，将生成到 gen/msg 目录..."

# 确保输出目录存在
mkdir -p gen/msg

# 编译所有 proto 文件
# 注意：使用 source_relative 时，输出文件会生成在 proto 文件所在目录
# 需要手动移动到 gen/msg 目录
for proto_file in ./pkg/protocol/*.proto; do
    if [ -f "$proto_file" ]; then
        echo "编译: $proto_file"
        protoc --go_out=gen/msg --go_opt=paths=source_relative "$proto_file"
        # 移动生成的文件到 gen/msg（如果生成到了子目录）
        proto_name=$(basename "$proto_file" .proto)
        if [ -d "gen/msg/lucky" ]; then
            find gen/msg/lucky -name "${proto_name}.pb.go" -exec mv {} gen/msg/ \; 2>/dev/null || true
            rm -rf gen/msg/lucky 2>/dev/null || true
        fi
        # 修改包名为 msg
        if [ -f "gen/msg/${proto_name}.pb.go" ]; then
            sed -i '' -e 's/^package pb$/package msg/' -e 's/^package protocol$/package msg/' gen/msg/${proto_name}.pb.go 2>/dev/null || \
            sed -i -e 's/^package pb$/package msg/' -e 's/^package protocol$/package msg/' gen/msg/${proto_name}.pb.go
            echo "  ✓ 已生成到 gen/msg/${proto_name}.pb.go"
        fi
    fi
done

if [ $? -eq 0 ]; then
    echo "✓ Protocol 文件编译成功"
    echo "生成的文件:"
    ls -lh gen/msg/*.pb.go 2>/dev/null | grep -v "_test.go" || echo "  无文件生成"
else
    echo "✗ Protocol 文件编译失败"
    exit 1
fi

