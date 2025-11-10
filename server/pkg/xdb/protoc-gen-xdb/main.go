package main

import (
	"bytes"
	"flag"
	"fmt"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "v1.0.0"

var (
	showVersion = flag.Bool("version", false, "show version")

	// SQL schema buffers for MySQL
	// key: relativePath/tableName, value: buffer
	sqlSchemaBuffers = make(map[string]*bytes.Buffer)

	// SQL file paths mapping: tableName -> relativePath
	sqlFilePaths = make(map[string]string)
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		return
	}

	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			// 检查文件是否包含 xdb 选项
			if !hasXdbMessage(f) {
				continue
			}

			// 生成 xdb 代码
			if err := generateXdbFile(gen, f); err != nil {
				return err
			}
		}

		// 生成 SQL 文件
		if err := generateSQLFiles(gen); err != nil {
			return err
		}

		return nil
	})
}

// hasXdbMessage 检查文件是否包含需要生成 xdb 代码的 message
func hasXdbMessage(f *protogen.File) bool {
	for _, msg := range f.Messages {
		if getTableName(msg) != "" {
			return true
		}
	}
	return false
}

// generateXdbFile 生成 xdb 文件
func generateXdbFile(gen *protogen.Plugin, f *protogen.File) error {
	// 计算相对路径，保持与 proto 文件相同的目录结构
	// 例如: db/proto/center/uuid.proto -> center/uuid.xdb.pb.go
	protoPath := f.Desc.Path()
	relativePath := getRelativePathForXdb(protoPath)
	var filename string
	if relativePath != "" {
		filename = relativePath + ".xdb.pb.go"
	} else {
		// 根目录文件
		filename = f.GeneratedFilenamePrefix + ".xdb.pb.go"
		// 如果 GeneratedFilenamePrefix 包含路径，只取文件名部分
		if idx := strings.LastIndex(filename, "/"); idx >= 0 {
			filename = filename[idx+1:]
		}
	}
	// 使用空字符串作为 import path，让 protoc 根据 filename 路径自动处理
	// 这样可以避免生成到 lucky/server/gen/db 这样的完整路径
	g := gen.NewGeneratedFile(filename, "")

	// 收集所有需要生成的 message
	messagesToGenerate := []*protogen.Message{}
	for _, msg := range f.Messages {
		table := getTableName(msg)
		if table != "" {
			messagesToGenerate = append(messagesToGenerate, msg)
		}
	}

	if len(messagesToGenerate) == 0 {
		return nil
	}

	// 生成文件头
	data := &TemplateData{
		PackageName: string(f.GoPackageName),
		Version:     version,
		Imports: []ImportInfo{
			{Path: "context"},
			{Path: "encoding/json"},
			{Path: "fmt"},
			{Path: "reflect"},
			{Path: "lucky/server/pkg/xdb"},
			{Path: "google.golang.org/protobuf/proto", Alias: "proto"},
		},
	}

	// 解析并执行文件头模板
	headerTmpl, err := parseTemplate("fileHeader", fileHeaderTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse file header template: %w", err)
	}
	headerCode, err := executeTemplate(headerTmpl, data)
	if err != nil {
		return fmt.Errorf("failed to execute file header template: %w", err)
	}
	g.P(headerCode)

	// 为每个 message 生成代码
	for _, msg := range messagesToGenerate {
		if err := generateMessageCode(g, gen, f, msg); err != nil {
			return err
		}
	}

	return nil
}

// generateMessageCode 为单个 message 生成代码
func generateMessageCode(g *protogen.GeneratedFile, gen *protogen.Plugin, f *protogen.File, msg *protogen.Message) error {
	// 构建模板数据
	data, err := buildTemplateData(g, f, msg)
	if err != nil {
		return fmt.Errorf("failed to build template data for %s: %w", msg.GoIdent.GoName, err)
	}

	// 解析所有模板
	templates := map[string]string{
		"fieldConstants": fieldConstantsTemplate,
		"pk":             pkTemplate,
		"record":         recordTemplate,
		"mutableRecord":  mutableRecordTemplate,
		"source":         sourceTemplate,
		"commitment":     commitmentTemplate,
		"init":           initTemplate,
	}

	parsedTemplates := make(map[string]*template.Template)
	for name, text := range templates {
		tmpl, err := parseTemplate(name, text)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", name, err)
		}
		parsedTemplates[name] = tmpl
	}

	// 按顺序生成代码
	generateOrder := []string{"fieldConstants", "pk", "record", "mutableRecord", "source", "commitment", "init"}
	for _, templateName := range generateOrder {
		tmpl := parsedTemplates[templateName]
		code, err := executeTemplate(tmpl, data)
		if err != nil {
			return fmt.Errorf("failed to execute template %s: %w", templateName, err)
		}
		g.P(code)
	}

	// 生成 SQL schema（如果驱动是 MySQL）
	tableName := getTableName(msg)
	if tableName != "" {
		driver := getDriver(msg)
		comments := msg.Comments.Leading.String()
		commentsLower := strings.ToLower(comments)

		shouldGenerate := driver == "mysql" ||
			strings.Contains(commentsLower, "driver_mysql") ||
			strings.Contains(commentsLower, "mysql") ||
			strings.Contains(comments, "MySQL")

		// 临时测试：强制为所有有 table 的 message 生成 SQL
		shouldGenerate = true

		if shouldGenerate {
			generateSQLSchema(gen, f, msg)
		}
	}

	return nil
}

// buildTemplateData 构建模板数据
func buildTemplateData(g *protogen.GeneratedFile, f *protogen.File, msg *protogen.Message) (*TemplateData, error) {
	tableName := getTableName(msg)
	driverName := getDriver(msg)
	pkFields := getPKFields(msg)

	// 构建字段信息
	fields := []FieldInfo{}
	pkFieldInfos := []FieldInfo{}
	fieldPrefix := msg.GoIdent.GoName + "Field"

	for _, field := range msg.Fields {
		isRuntime := isRuntimeField(field)
		if isRuntime {
			continue
		}

		fieldName := toUpperCamelCase(strings.TrimPrefix(field.GoName, "X"))
		constName := fieldPrefix + fieldName
		goType := fieldGoType(g, field)

		fieldInfo := FieldInfo{
			GoName:    field.GoName,
			ProtoName: string(field.Desc.Name()),
			GoType:    goType,
			IsPK:      false,
			IsRuntime: isRuntime,
			ConstName: constName,
			Comment:   getFieldComment(field),
		}

		// 检查是否为主键字段
		for _, pkField := range pkFields {
			if pkField == field {
				fieldInfo.IsPK = true
				pkFieldInfos = append(pkFieldInfos, fieldInfo)
				break
			}
		}

		fields = append(fields, fieldInfo)
	}

	// 构建模板数据
	data := &TemplateData{
		PackageName:    string(f.GoPackageName),
		Version:        version,
		MessageName:    msg.GoIdent.GoName,
		Fields:         fields,
		PKFields:       pkFieldInfos,
		FieldPrefix:    fieldPrefix,
		RecordName:     msg.GoIdent.GoName + "Record",
		PKName:         msg.GoIdent.GoName + "PK",
		CommitmentName: msg.GoIdent.GoName + "Commitment",
		SourceName:     "_" + msg.GoIdent.GoName + "Source",
		TableName:      tableName,
		DriverName:     driverName,
		Namespace:      tableName,
		KeySize:        len(pkFields),
	}

	return data, nil
}

// 辅助函数

func getTableName(msg *protogen.Message) string {
	// 从 message options 中获取 table 名称
	// 注意：由于 extension 读取需要编译后的 extension.pb.go，这里先使用简化方法
	// 实际应该从 proto extension 中读取 (xdb.table = 71002)
	// 暂时从注释或其他方式推断，或者使用默认值

	opts := msg.Desc.Options()
	if opts == nil {
		// 如果没有选项，使用 message 名称的小写形式作为默认值
		return strings.ToLower(msg.GoIdent.GoName)
	}

	// 尝试通过 protoreflect 读取扩展字段
	// 字段号 71002 对应 xdb.table
	// 由于需要 extension descriptor，这里先返回默认值
	// TODO: 实现正确的 extension 读取

	// 使用 message 名称的小写形式作为默认值
	// 用户可以在生成的代码中手动修改，或者后续实现正确的 extension 读取
	return strings.ToLower(msg.GoIdent.GoName)
}

func getDriver(msg *protogen.Message) string {
	// 从 message options 中获取 driver
	opts := msg.Desc.Options()
	if opts == nil {
		return "none"
	}

	// 尝试读取 xdb.driver 扩展（字段号 71003）
	// DriverType 枚举值：DRIVER_NONE=0, DRIVER_MYSQL=1, DRIVER_MONGODB=2
	// 由于需要 extension descriptor，这里先使用简化方法
	// 检查消息的注释中是否有 driver 提示
	comments := msg.Comments.Leading.String()
	if strings.Contains(comments, "DRIVER_MYSQL") || strings.Contains(comments, "mysql") {
		return "mysql"
	}
	if strings.Contains(comments, "DRIVER_MONGODB") || strings.Contains(comments, "mongo") {
		return "mongo"
	}

	// 默认返回 "none"
	// 用户可以通过修改生成的代码或使用 fix_driver.sh 来设置正确的驱动
	return "none"
}

func getPKFields(msg *protogen.Message) []*protogen.Field {
	var pkFields []*protogen.Field
	for _, field := range msg.Fields {
		// 检查字段是否有 pk 选项
		// 这里简化处理，实际应该从 proto extension 中读取
		// 暂时使用启发式方法：字段名包含 Id 的作为主键
		if strings.HasPrefix(field.GoName, "Id") ||
			strings.HasSuffix(field.GoName, "Id") ||
			strings.HasPrefix(field.GoName, "ID") ||
			strings.HasSuffix(field.GoName, "ID") {
			pkFields = append(pkFields, field)
		}
	}

	// 如果没有找到主键字段，使用第一个非运行时字段作为主键
	if len(pkFields) == 0 && len(msg.Fields) > 0 {
		for _, field := range msg.Fields {
			if !isRuntimeField(field) {
				pkFields = append(pkFields, field)
				break
			}
		}
	}

	return pkFields
}

func isRuntimeField(field *protogen.Field) bool {
	// 检查是否是运行时字段（以 _ 开头）
	return strings.HasPrefix(string(field.Desc.Name()), "_")
}

func findField(msg *protogen.Message, name string) *protogen.Field {
	for _, field := range msg.Fields {
		if string(field.Desc.Name()) == name {
			return field
		}
	}
	return nil
}

func fieldGoType(g *protogen.GeneratedFile, field *protogen.Field) string {
	// 对于基本类型，直接返回 Go 类型
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.MessageKind:
		// 对于 message 类型，使用完整的类型名
		return g.QualifiedGoIdent(field.Message.GoIdent)
	case protoreflect.EnumKind:
		// 对于 enum 类型，使用完整的类型名
		return g.QualifiedGoIdent(field.Enum.GoIdent)
	default:
		return g.QualifiedGoIdent(field.GoIdent)
	}
}

func toUpperCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// getSQLSchemaBuffer 获取 SQL schema buffer
func getSQLSchemaBuffer(srcFileName, tableName string) *bytes.Buffer {
	buff, ok := sqlSchemaBuffers[tableName]
	if !ok {
		buff = &bytes.Buffer{}
		sqlSchemaBuffers[tableName] = buff
		buff.WriteString("-- Code generated by protoc-gen-xdb. DO NOT EDIT.\n")
		buff.WriteString("-- source: " + srcFileName + "\n")
	}
	return buff
}

// generateSQLSchema 生成 MySQL SQL schema
func generateSQLSchema(gen *protogen.Plugin, f *protogen.File, msg *protogen.Message) {
	srcFileName := f.Desc.Path()
	tableName := getTableName(msg)
	if tableName == "" {
		return
	}

	// 确保 sqlSchemaBuffers 已初始化
	if sqlSchemaBuffers == nil {
		sqlSchemaBuffers = make(map[string]*bytes.Buffer)
		sqlFilePaths = make(map[string]string)
	}

	// 计算相对路径（相对于 db/proto 目录）
	// 例如: center/uuid.proto -> center/uuid.sql
	relativePath := getRelativePathForSQL(srcFileName)
	sqlFilePaths[tableName] = relativePath

	// 使用 tableName 作为 key，但记录路径信息
	schema := getSQLSchemaBuffer(srcFileName, tableName)
	schema.WriteString("\nCREATE TABLE `" + tableName + "` (")

	pks := []string{}
	idx := 0

	for _, field := range msg.Fields {
		if isRuntimeField(field) {
			continue
		}
		idx++

		isPK := false
		for _, pkField := range getPKFields(msg) {
			if pkField == field {
				isPK = true
				break
			}
		}

		schema.WriteString("\n    `")
		schema.WriteString(string(field.Desc.Name()))
		schema.WriteString("` ")

		// 根据字段类型生成 SQL 类型
		switch field.Desc.Kind() {
		case protoreflect.FloatKind:
			schema.WriteString("FLOAT NOT NULL")
		case protoreflect.DoubleKind:
			schema.WriteString("DOUBLE NOT NULL")
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
			schema.WriteString("BIGINT(20) NOT NULL DEFAULT 0")
		case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind, protoreflect.EnumKind:
			schema.WriteString("INT(11) NOT NULL DEFAULT 0")
		case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
			schema.WriteString("BIGINT(20) UNSIGNED NOT NULL DEFAULT 0")
		case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
			schema.WriteString("INT(11) UNSIGNED NOT NULL DEFAULT 0")
		case protoreflect.BoolKind:
			schema.WriteString("TINYINT NOT NULL DEFAULT 0")
		case protoreflect.StringKind:
			// 默认 VARCHAR(255)，可以根据需要调整
			schema.WriteString("VARCHAR(255) NOT NULL DEFAULT ''")
		case protoreflect.BytesKind:
			schema.WriteString("BLOB NOT NULL")
		case protoreflect.MessageKind:
			// 消息类型使用 TEXT 存储 JSON
			schema.WriteString("TEXT NOT NULL")
		default:
			schema.WriteString("TEXT NOT NULL")
		}

		// 添加注释
		if comment := getFieldComment(field); comment != "" {
			schema.WriteString(" COMMENT '")
			schema.WriteString(comment)
			schema.WriteString("'")
		}

		schema.WriteString(",")

		if isPK {
			pks = append(pks, string(field.Desc.Name()))
		}
	}

	// 添加主键
	if len(pks) > 0 {
		schema.WriteString("\n    PRIMARY KEY(`")
		schema.WriteString(strings.Join(pks, "`, `"))
		schema.WriteString("`)\n")
	} else if idx > 0 {
		// 移除最后一个逗号
		schema.Truncate(schema.Len() - 1)
	}

	schema.WriteString(") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;\n")
}

// getFieldComment 获取字段注释
func getFieldComment(field *protogen.Field) string {
	// 简化处理：目前返回空字符串
	// 可以扩展以支持从 xdb.comment 选项读取
	return ""
}

// generateSQLFiles 生成 SQL 文件
func generateSQLFiles(gen *protogen.Plugin) error {
	if len(sqlSchemaBuffers) == 0 {
		// 没有 SQL schema 需要生成
		return nil
	}

	for tableName, schemaBuffer := range sqlSchemaBuffers {
		// 获取相对路径（如果有）
		relativePath, hasPath := sqlFilePaths[tableName]
		var sqlFileName string
		if hasPath && relativePath != "" {
			// 使用相对路径，例如: center/uuid.sql
			sqlFileName = relativePath + ".sql"
		} else {
			// 默认在根目录
			sqlFileName = tableName + ".sql"
		}

		// 创建输出文件（使用相对路径，protoc 会处理输出目录）
		// 注意：NewGeneratedFile 的第二个参数是 Go import path，对于 SQL 文件可以留空
		g := gen.NewGeneratedFile(sqlFileName, "")
		if g == nil {
			return fmt.Errorf("failed to create generated file for %s", sqlFileName)
		}

		// 写入 SQL 内容
		data := schemaBuffer.Bytes()
		if len(data) == 0 {
			continue
		}

		if _, err := g.Write(data); err != nil {
			return fmt.Errorf("failed to write %s: %w", sqlFileName, err)
		}
	}
	return nil
}

// getRelativePathForXdb 根据 proto 文件路径计算 xdb 文件的相对路径
func getRelativePathForXdb(protoPath string) string {
	// protoPath 格式可能是: db/proto/center/uuid.proto 或 center/uuid.proto
	// 需要提取相对于 db/proto 的路径部分，保留子目录结构

	// 移除 .proto 后缀
	path := strings.TrimSuffix(protoPath, ".proto")

	// 查找 db/proto/ 的位置
	dbProtoIdx := strings.Index(path, "db/proto/")
	if dbProtoIdx >= 0 {
		// 提取 db/proto/ 之后的部分（包括子目录和文件名）
		relativePath := path[dbProtoIdx+len("db/proto/"):]
		// 如果有子目录，返回子目录/文件名，否则返回空字符串
		if idx := strings.LastIndex(relativePath, "/"); idx >= 0 {
			// 有子目录，返回 center/uuid
			return relativePath
		}
		// 没有子目录，返回空字符串（表示根目录）
		return ""
	}

	// 如果没有 db/proto/，尝试查找 proto/ 之后的部分
	protoIdx := strings.Index(path, "proto/")
	if protoIdx >= 0 {
		relativePath := path[protoIdx+len("proto/"):]
		if idx := strings.LastIndex(relativePath, "/"); idx >= 0 {
			// 有子目录
			return relativePath
		}
		return ""
	}

	// 如果都没有，检查是否包含 /，提取目录部分
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		dir := path[:idx]
		file := path[idx+1:]
		if dir != "" {
			return dir + "/" + file
		}
		return file
	}

	return ""
}

// getRelativePathForSQL 根据 proto 文件路径计算 SQL 文件的相对路径
func getRelativePathForSQL(protoPath string) string {
	// protoPath 格式可能是: db/proto/center/uuid.proto 或 center/uuid.proto
	// 需要提取相对于 db/proto 的路径部分

	// 移除 .proto 后缀
	path := strings.TrimSuffix(protoPath, ".proto")

	// 查找 db/proto/ 的位置
	dbProtoIdx := strings.Index(path, "db/proto/")
	if dbProtoIdx >= 0 {
		// 提取 db/proto/ 之后的部分
		relativePath := path[dbProtoIdx+len("db/proto/"):]
		// 移除文件名，只保留目录路径
		if idx := strings.LastIndex(relativePath, "/"); idx >= 0 {
			return relativePath[:idx] + "/" + strings.TrimPrefix(relativePath[idx+1:], "/")
		}
		return ""
	}

	// 如果没有 db/proto/，尝试查找 proto/ 之后的部分
	protoIdx := strings.Index(path, "proto/")
	if protoIdx >= 0 {
		relativePath := path[protoIdx+len("proto/"):]
		if idx := strings.LastIndex(relativePath, "/"); idx >= 0 {
			return relativePath[:idx] + "/" + strings.TrimPrefix(relativePath[idx+1:], "/")
		}
		return ""
	}

	// 如果都没有，检查是否包含 /，提取目录部分
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		dir := path[:idx]
		file := path[idx+1:]
		if dir != "" {
			return dir + "/" + file
		}
		return file
	}

	return ""
}
