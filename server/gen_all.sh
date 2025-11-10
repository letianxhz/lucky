#!/bin/bash

# 生成所有代码的脚本
# 包括：pb 代码、msg 代码、db 脚本、config 代码

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

GEN_DIR="$SCRIPT_DIR/gen"
DB_DIR="$GEN_DIR/db"        # pb 代码输出目录（改为 db）
MSG_DIR="$GEN_DIR/msg"
SQL_DIR="$GEN_DIR/sql"      # SQL 脚本输出目录（改为 sql）
CONFIG_DIR="$GEN_DIR/config"

XDB_DIR="$SCRIPT_DIR/pkg/xdb"
PROTOC_GEN_XDB="$XDB_DIR/protoc-gen-xdb/protoc-gen-xdb"

echo "=== 生成所有代码 ==="
echo "Server 目录: $SCRIPT_DIR"
echo "生成目录: $GEN_DIR"
echo ""

# 创建所有输出目录
mkdir -p "$DB_DIR" "$MSG_DIR" "$SQL_DIR" "$CONFIG_DIR"

# 检查 protoc-gen-xdb 是否存在
if [ ! -f "$PROTOC_GEN_XDB" ]; then
    echo "构建 protoc-gen-xdb..."
    cd "$XDB_DIR/protoc-gen-xdb"
    ./build.sh
    cd "$SCRIPT_DIR"
fi

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo "错误: protoc 未安装"
    echo "请安装 Protocol Buffers 编译器:"
    echo "  brew install protobuf  # macOS"
    exit 1
fi

# 查找所有 .proto 文件（在 db/proto 目录下，包括子目录）
PROTO_DIR="$SCRIPT_DIR/db/proto"
PROTO_FILES=$(find "$PROTO_DIR" -name "*.proto" -type f 2>/dev/null | grep -v extension.proto)

# 如果没有找到，尝试从 example 目录查找（兼容旧代码）
if [ -z "$PROTO_FILES" ]; then
    PROTO_DIR="$SCRIPT_DIR/pkg/xdb/example"
    PROTO_FILES=$(find "$PROTO_DIR" -maxdepth 1 -name "*.proto" -type f 2>/dev/null)
fi

if [ -z "$PROTO_FILES" ]; then
    echo "⚠ 未找到 .proto 文件在 $PROTO_DIR"
    exit 0
fi

echo "1. 生成 db 代码（pb 代码）..."
for proto_file in $PROTO_FILES; do
    proto_name=$(basename "$proto_file" .proto)
    proto_dir=$(dirname "$proto_file")
    
    echo "   处理: $proto_file"
    
    # 计算相对于 PROTO_DIR 的路径
    proto_rel_path=$(python3 -c "import os; print(os.path.relpath('$proto_file', '$PROTO_DIR'))" 2>/dev/null || echo "$(basename "$proto_file")")
    proto_rel_dir=$(dirname "$proto_rel_path")
    
    # 如果 proto_rel_dir 是 "."，说明在根目录，否则在子目录
    if [ "$proto_rel_dir" = "." ]; then
        output_subdir=""
        package_name="db"
    else
        # 保持子目录结构，例如 center/uuid.proto -> gen/db/center/
        output_subdir="$proto_rel_dir"
        package_name="$(basename "$proto_rel_dir")"  # 使用目录名作为包名（如 center）
        mkdir -p "$DB_DIR/$output_subdir"
    fi
    
    # 生成 proto Go 代码（使用 source_relative 会自动保持目录结构）
    protoc \
      --go_out="$DB_DIR" \
      --go_opt=paths=source_relative \
      --proto_path="$PROTO_DIR" \
      --proto_path="$XDB_DIR" \
      "$proto_file"
    
    # 生成 xdb 代码（也需要保持目录结构）
    protoc \
      --proto_path="$PROTO_DIR" \
      --proto_path="$XDB_DIR" \
      --plugin=protoc-gen-xdb="$PROTOC_GEN_XDB" \
      --xdb_out="$DB_DIR" \
      "$proto_file"
    
    echo "   ✓ $proto_name.pb.go 和 ${proto_name}_xdb.pb.go 生成成功"
done

# 移动生成的文件到正确的目录（如果生成到了子目录，如 lucky/server/gen/db）
# 处理各种可能的路径格式
if [ -d "$DB_DIR/lucky" ]; then
    find "$DB_DIR/lucky" -name "*.pb.go" -type f | while read file; do
        # 尝试提取相对路径（从 lucky/server/gen/db/ 之后）
        rel_path=$(echo "$file" | sed "s|^$DB_DIR/lucky/server/gen/db/||" 2>/dev/null)
        if [ "$rel_path" = "$file" ] || [ -z "$rel_path" ]; then
            # 尝试其他格式（从 lucky/ 之后）
            rel_path=$(echo "$file" | sed "s|^$DB_DIR/lucky/||" 2>/dev/null)
        fi
        if [ "$rel_path" != "$file" ] && [ -n "$rel_path" ]; then
            # 提取子目录部分（如 center/uuid.xdb.pb.go -> center/）
            target_dir="$DB_DIR/$(dirname "$rel_path")"
            # 如果是根目录，target_dir 就是 DB_DIR
            if [ "$(dirname "$rel_path")" = "." ]; then
                target_dir="$DB_DIR"
            fi
            mkdir -p "$target_dir"
            mv "$file" "$target_dir/" 2>/dev/null && echo "   移动: $(basename "$file") -> $target_dir/"
        fi
    done
    rm -rf "$DB_DIR/lucky" 2>/dev/null || true
fi

# 修改生成的代码包名
# 对于根目录的文件，包名为 db
for f in "$DB_DIR"/*.pb.go; do
    if [ -f "$f" ]; then
        sed -i '' -e 's/^package example$/package db/' -e 's/^package main$/package db/' -e 's/^package pb$/package db/' "$f" 2>/dev/null || \
        sed -i -e 's/^package example$/package db/' -e 's/^package main$/package db/' -e 's/^package pb$/package db/' "$f"
    fi
done

# 对于子目录的文件，包名为目录名
find "$DB_DIR" -type d -mindepth 1 | while read subdir; do
    package_name=$(basename "$subdir")
    for f in "$subdir"/*.pb.go; do
        if [ -f "$f" ]; then
            sed -i '' -e "s/^package example$/package $package_name/" -e "s/^package main$/package $package_name/" -e "s/^package pb$/package $package_name/" -e "s/^package db$/package $package_name/" "$f" 2>/dev/null || \
            sed -i -e "s/^package example$/package $package_name/" -e "s/^package main$/package $package_name/" -e "s/^package pb$/package $package_name/" -e "s/^package db$/package $package_name/" "$f"
        fi
    done
done

# 移动 SQL 文件到 sql 目录（保持子目录结构）
# 从 DB_DIR 移动 SQL 文件到 SQL_DIR
find "$DB_DIR" -name "*.sql" -type f | while read sql_file; do
    # 计算相对于 DB_DIR 的路径
    rel_path=$(echo "$sql_file" | sed "s|^$DB_DIR/||")
    target_dir="$SQL_DIR/$(dirname "$rel_path")"
    # 如果是根目录的文件，target_dir 就是 SQL_DIR
    if [ "$(dirname "$rel_path")" = "." ]; then
        target_dir="$SQL_DIR"
    fi
    mkdir -p "$target_dir"
    mv "$sql_file" "$target_dir/" 2>/dev/null || true
done

# 兼容：从 PROTO_DIR 复制 SQL 文件（如果存在）
if [ -f "$PROTO_DIR/player.sql" ]; then
    cp "$PROTO_DIR/player.sql" "$SQL_DIR/" 2>/dev/null || true
fi
if [ -f "$PROTO_DIR/item.sql" ]; then
    cp "$PROTO_DIR/item.sql" "$SQL_DIR/" 2>/dev/null || true
fi

echo ""
echo "2. 生成 msg 代码（从 pkg/protocol）..."
# 从 pkg/protocol 目录的 .proto 文件生成 Go 代码
PROTOCOL_DIR="$SCRIPT_DIR/pkg/protocol"
PROTOCOL_FILES=$(find "$PROTOCOL_DIR" -maxdepth 1 -name "*.proto" -type f 2>/dev/null)

if [ -n "$PROTOCOL_FILES" ]; then
    # 检查 protoc-gen-go 插件
    if ! command -v protoc-gen-go &> /dev/null; then
        echo "   ⚠ protoc-gen-go 未安装，跳过 msg 代码生成"
    else
        for proto_file in $PROTOCOL_FILES; do
            proto_name=$(basename "$proto_file" .proto)
            
            echo "   处理: $proto_name.proto"
            
            # 生成 proto Go 代码到 gen/msg 目录
            protoc \
              --go_out="$MSG_DIR" \
              --go_opt=paths=source_relative \
              --proto_path="$PROTOCOL_DIR" \
              "$proto_file"
            
            echo "   ✓ ${proto_name}.pb.go 生成成功"
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
    fi
else
    echo "   ⚠ 未找到协议文件在 $PROTOCOL_DIR"
fi

echo ""
echo "3. 生成 config 代码..."
# 生成配置代码（数据库配置等）
cat > "$CONFIG_DIR/db_config.go" <<EOF
// Code generated by gen_all.sh. DO NOT EDIT.

package config

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	// TODO: 从 proto 文件生成数据库配置
}

// GetDBConfig 获取数据库配置
func GetDBConfig() *DatabaseConfig {
	return &DatabaseConfig{}
}
EOF

echo "   ✓ db_config.go 生成成功"

echo ""
echo "=== 生成完成 ==="
echo ""
echo "生成的文件:"
echo "  DB 代码（PB）: $DB_DIR"
ls -lh "$DB_DIR"/*.pb.go 2>/dev/null | awk '{print "    " $9}' || echo "    无 .pb.go 文件"
echo ""
echo "  MSG 代码: $MSG_DIR"
ls -lh "$MSG_DIR"/*_msg.go 2>/dev/null | awk '{print "    " $9}' || echo "    无 _msg.go 文件"
echo ""
echo "  SQL 脚本: $SQL_DIR"
ls -lh "$SQL_DIR"/*.sql 2>/dev/null | awk '{print "    " $9}' || echo "    无 .sql 文件"
echo ""
echo "  Config 代码: $CONFIG_DIR"
ls -lh "$CONFIG_DIR"/*.go 2>/dev/null | awk '{print "    " $9}' || echo "    无 .go 文件"
echo ""

