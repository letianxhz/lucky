package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"lucky/server/pkg/xdb"

	"github.com/pkg/errors"
)

// toSnakeCase 将驼峰命名转换为蛇形命名
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Table MySQL 表实现
type Table struct {
	src       *xdb.Source
	dao       *Dao
	sqlFields string
}

func (tm *TableMgr) CreateTable(dao *Dao, src *xdb.Source) *Table {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tableName := src.TableName
	if table, ok := tm.tables[tableName]; ok {
		return table
	}

	// 构建 SQL 字段列表
	var buf strings.Builder
	names := src.FieldNames(false)
	// 如果 Fields 为空，使用反射获取字段名
	if len(names) == 0 {
		// 从 ProtoType 获取字段名
		protoType := src.ProtoType
		if protoType != nil {
			for i := 0; i < protoType.NumField(); i++ {
				field := protoType.Field(i)
				// 跳过未导出的字段和特殊字段（如 state, sizeCache, unknownFields, XVersion）
				if !field.IsExported() ||
					field.Name == "state" ||
					field.Name == "sizeCache" ||
					field.Name == "unknownFields" ||
					field.Name == "XVersion" {
					continue
				}
				if buf.Len() > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString("`")
				// 转换字段名为数据库字段名（snake_case）
				dbName := toSnakeCase(field.Name)
				buf.WriteString(dbName)
				buf.WriteString("`")
			}
		}
	} else {
		for i, name := range names {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString("`")
			buf.WriteString(name)
			buf.WriteString("`")
		}
	}

	table := &Table{
		src:       src,
		dao:       dao,
		sqlFields: buf.String(),
	}
	tm.tables[tableName] = table
	return table
}

func (t *Table) Recover(ctx context.Context, commitments []xdb.Commitment) error {
	// 恢复逻辑：从重做日志恢复数据
	// 这里可以添加恢复逻辑
	_ = t.Save(ctx, commitments, 5*time.Second, 100*time.Millisecond, func() bool { return true })
	return nil
}

func (t *Table) Save(ctx context.Context, commitments []xdb.Commitment, writeTimeout time.Duration, retryInterval time.Duration, running func() bool) bool {
	if len(commitments) == 0 {
		return true
	}

	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()

	for _, commitment := range commitments {
		if !running() {
			return false
		}

		data, _ := commitment.PrepareWrite()
		lifecycle := commitment.Lifecycle()

		// 构建 SQL 语句
		var sql string
		var args []interface{}

		switch lifecycle {
		case xdb.LifecycleNew:
			// INSERT
			sql, args = t.buildInsertSQL(data)
		case xdb.LifecycleNormal:
			// UPDATE
			sql, args = t.buildUpdateSQL(commitment, data)
		case xdb.LifecycleDeleted:
			// DELETE
			sql, args = t.buildDeleteSQL(commitment)
		default:
			continue
		}

		_, err := t.dao.client.ExecContext(ctx, sql, args...)
		if err != nil {
			// 记录错误，但继续处理其他记录
			fmt.Printf("Failed to save commitment: %v, SQL: %s, Args: %v\n", err, sql, args)
			continue
		}
	}

	return true
}

func (t *Table) buildInsertSQL(data interface{}) (string, []interface{}) {
	// 从 data 中提取字段值
	protoType := t.src.ProtoType
	if protoType == nil {
		return "", nil
	}

	// 使用反射获取字段值
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var fields []string
	var placeholders []string
	var args []interface{}

	for i := 0; i < protoType.NumField(); i++ {
		field := protoType.Field(i)
		// 跳过特殊字段
		if !field.IsExported() ||
			field.Name == "state" ||
			field.Name == "sizeCache" ||
			field.Name == "unknownFields" ||
			field.Name == "XVersion" {
			continue
		}

		dbName := toSnakeCase(field.Name)
		fields = append(fields, "`"+dbName+"`")
		placeholders = append(placeholders, "?")

		// 获取字段值
		fieldVal := val.Field(i)
		if fieldVal.IsValid() {
			args = append(args, fieldVal.Interface())
		} else {
			args = append(args, nil)
		}
	}

	sql := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		t.src.TableName,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "))
	return sql, args
}

func (t *Table) buildUpdateSQL(commitment xdb.Commitment, data interface{}) (string, []interface{}) {
	// 构建 UPDATE SQL
	pk := commitment.Source().PKOf(commitment)
	if pk == nil {
		return "", nil
	}

	// 获取变更的字段
	changes := commitment.Changes()
	if changes == 0 {
		return "", nil
	}

	// 从 data 中提取字段值
	protoType := t.src.ProtoType
	if protoType == nil {
		return "", nil
	}

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var setParts []string
	var args []interface{}

	// 根据 Source 的 Fields 信息构建 SET 子句
	for i := 0; i < protoType.NumField(); i++ {
		field := protoType.Field(i)
		// 跳过特殊字段
		if !field.IsExported() ||
			field.Name == "state" ||
			field.Name == "sizeCache" ||
			field.Name == "unknownFields" ||
			field.Name == "XVersion" {
			continue
		}

		// 检查字段是否在变更集合中（简化处理：假设所有字段都变更）
		// TODO: 根据 FieldSet 精确判断哪些字段变更了
		dbName := toSnakeCase(field.Name)
		setParts = append(setParts, "`"+dbName+"` = ?")

		fieldVal := val.Field(i)
		if fieldVal.IsValid() {
			args = append(args, fieldVal.Interface())
		} else {
			args = append(args, nil)
		}
	}

	// 构建 WHERE 子句（主键）
	whereClause, pkArgs := t.buildWhereClause(pk)
	args = append(args, pkArgs...)

	sql := fmt.Sprintf("UPDATE `%s` SET %s WHERE %s",
		t.src.TableName,
		strings.Join(setParts, ", "),
		whereClause)
	return sql, args
}

func (t *Table) buildWhereClause(pk xdb.PK) (string, []interface{}) {
	// 使用反射获取主键值
	pkType := reflect.TypeOf(pk)
	if pkType.Kind() == reflect.Ptr {
		pkType = pkType.Elem()
	}
	pkVal := reflect.ValueOf(pk)
	if pkVal.Kind() == reflect.Ptr {
		pkVal = pkVal.Elem()
	}

	// 获取第一个字段作为主键值（简化处理）
	if pkType.NumField() > 0 {
		field := pkType.Field(0)
		fieldVal := pkVal.Field(0)
		if fieldVal.IsValid() {
			dbName := toSnakeCase(field.Name)
			return "`" + dbName + "` = ?", []interface{}{fieldVal.Interface()}
		}
	}

	return "", nil
}

func (t *Table) buildDeleteSQL(commitment xdb.Commitment) (string, []interface{}) {
	// 构建 DELETE SQL
	pk := commitment.Source().PKOf(commitment)
	if pk == nil {
		return "", nil
	}

	whereClause, args := t.buildWhereClause(pk)
	if whereClause == "" {
		return "", nil
	}

	sql := fmt.Sprintf("DELETE FROM `%s` WHERE %s", t.src.TableName, whereClause)
	return sql, args
}

func (t *Table) Fetch(ctx context.Context, onlyOne bool, pk xdb.PK) (xdb.RecordCursor, error) {
	whereClause, args := t.buildWhereClause(pk)
	if whereClause == "" {
		return nil, errors.New("failed to build where clause")
	}

	sql := fmt.Sprintf("SELECT %s FROM `%s` WHERE %s", t.sqlFields, t.src.TableName, whereClause)
	if onlyOne {
		sql += " LIMIT 1"
	}

	rows, err := t.dao.client.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch from MySQL")
	}

	return newRecordCursor(rows, t.src), nil
}

func (t *Table) FetchMulti(ctx context.Context, pks []xdb.PK) (xdb.RecordCursor, error) {
	if len(pks) == 0 {
		return &RecordCursor{rows: nil, src: t.src}, nil
	}

	// 构建 IN 查询
	placeholders := make([]string, len(pks))
	args := make([]interface{}, len(pks))
	for i, pk := range pks {
		placeholders[i] = "?"
		args[i] = pk.Full()
	}

	sql := fmt.Sprintf("SELECT %s FROM `%s` WHERE `player_id` IN (%s)",
		t.sqlFields, t.src.TableName, strings.Join(placeholders, ", "))

	rows, err := t.dao.client.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch multi from MySQL")
	}

	return newRecordCursor(rows, t.src), nil
}

func (t *Table) Find(ctx context.Context, filter interface{}) (xdb.RecordCursor, error) {
	// 简化实现：假设 filter 是 SQL WHERE 子句
	where := "1=1"
	args := []interface{}{}

	if filter != nil {
		// TODO: 根据 filter 类型构建 WHERE 子句
	}

	sql := fmt.Sprintf("SELECT %s FROM `%s` WHERE %s", t.sqlFields, t.src.TableName, where)
	rows, err := t.dao.client.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find from MySQL")
	}

	return newRecordCursor(rows, t.src), nil
}

// RecordCursor MySQL 记录游标
type RecordCursor struct {
	rows *sql.Rows
	src  *xdb.Source
}

func newRecordCursor(rows *sql.Rows, src *xdb.Source) *RecordCursor {
	return &RecordCursor{
		rows: rows,
		src:  src,
	}
}

func (r *RecordCursor) Next(ctx context.Context) bool {
	if r.rows == nil {
		return false
	}
	return r.rows.Next()
}

func (r *RecordCursor) Decode(val interface{}) error {
	if r.rows == nil {
		return errors.New("rows is nil")
	}

	// 获取目标类型
	valType := reflect.TypeOf(val)
	if valType.Kind() != reflect.Ptr {
		return errors.New("val must be a pointer")
	}
	valType = valType.Elem()

	// 如果仍然是指针类型（如 *PlayerModel），继续解引用
	if valType.Kind() == reflect.Ptr {
		valType = valType.Elem()
	}

	// 获取列信息
	columns, err := r.rows.Columns()
	if err != nil {
		return errors.Wrap(err, "failed to get columns")
	}

	// 创建值切片
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// 扫描行数据
	if err := r.rows.Scan(valuePtrs...); err != nil {
		return errors.Wrap(err, "failed to scan row")
	}

	// 解码到目标对象
	// 数据库存储的是 Record 数据，需要解码到嵌入的 proto 类型
	// 使用 Source 的 RecordType 来找到正确的解码目标
	// val 是指向目标对象的指针（如 *PlayerModel 或 **PlayerModel）
	valVal := reflect.ValueOf(val)
	if valVal.Kind() != reflect.Ptr {
		return errors.New("val must be a pointer")
	}
	valVal = valVal.Elem() // 解引用一次
	// 如果仍然是指针，继续解引用（处理 **PlayerModel 的情况）
	if valVal.Kind() == reflect.Ptr {
		valVal = valVal.Elem()
	}

	var targetVal reflect.Value
	var protoType reflect.Type

	// 使用 Source 的 RecordType 来确定 proto 类型
	if r.src != nil && r.src.RecordType != nil {
		recordType := r.src.RecordType
		// 在 RecordType 中查找 proto 类型（跳过 Header）
		for i := 0; i < recordType.NumField(); i++ {
			field := recordType.Field(i)
			if field.Anonymous && field.Type.Name() != "Header" {
				if field.Type.Kind() == reflect.Struct {
					protoType = field.Type
					// 在目标对象中找到对应的字段
					if valType.Kind() == reflect.Struct && valVal.IsValid() {
						// 如果是 Model 类型，找到嵌入的 Record 类型
						for j := 0; j < valType.NumField(); j++ {
							modelField := valType.Field(j)
							if modelField.Anonymous {
								// 检查是否是 Record 类型（通过比较类型或检查是否包含 Header）
								isRecordType := false
								if modelField.Type == recordType {
									isRecordType = true
								} else if modelField.Type.Kind() == reflect.Struct {
									// 检查是否包含 Header 字段
									for k := 0; k < modelField.Type.NumField(); k++ {
										subField := modelField.Type.Field(k)
										if subField.Anonymous && subField.Type.Name() == "Header" {
											isRecordType = true
											break
										}
									}
								}

								if isRecordType {
									// 找到了嵌入的 Record
									recordVal := valVal.Field(j)
									if recordVal.IsValid() {
										targetVal = recordVal.Field(i)
										break
									}
								}
							}
						}
						// 如果没找到，可能是直接的 Record 类型
						if !targetVal.IsValid() && valType == recordType {
							targetVal = valVal.Field(i)
						}
					}
					break
				}
			}
		}
	}

	// 如果通过 Source 没找到，尝试直接查找
	if !targetVal.IsValid() && valType.Kind() == reflect.Struct {
		// 查找 proto 类型（跳过 Header）
		for i := 0; i < valType.NumField(); i++ {
			field := valType.Field(i)
			if field.Anonymous && field.Type.Name() != "Header" {
				if field.Type.Kind() == reflect.Struct {
					protoType = field.Type
					targetVal = valVal.Field(i)
					break
				}
			}
		}
	}

	if !targetVal.IsValid() || protoType == nil {
		return errors.New(fmt.Sprintf("failed to find target proto type for decoding, valType: %v, RecordType: %v", valType, r.src.RecordType))
	}

	// 将数据库字段映射到 proto 字段
	columnMap := make(map[string]int)
	for i, col := range columns {
		columnMap[col] = i
	}

	// 填充字段值
	for i := 0; i < protoType.NumField(); i++ {
		field := protoType.Field(i)
		if !field.IsExported() ||
			field.Name == "state" ||
			field.Name == "sizeCache" ||
			field.Name == "unknownFields" ||
			field.Name == "XVersion" {
			continue
		}

		dbName := toSnakeCase(field.Name)
		if colIdx, ok := columnMap[dbName]; ok {
			fieldVal := targetVal.Field(i)
			if fieldVal.IsValid() && fieldVal.CanSet() {
				value := values[colIdx]
				if value != nil {
					// 类型转换
					rv := reflect.ValueOf(value)
					if rv.Type().AssignableTo(fieldVal.Type()) {
						fieldVal.Set(rv)
					} else if rv.Type().ConvertibleTo(fieldVal.Type()) {
						fieldVal.Set(rv.Convert(fieldVal.Type()))
					}
				}
			}
		}
	}

	return nil
}

func (r *RecordCursor) All(ctx context.Context, results interface{}) error {
	if r.rows == nil {
		return nil
	}
	defer r.rows.Close()
	// TODO: 实现批量解码
	return errors.New("All not implemented yet")
}

func (r *RecordCursor) Close(ctx context.Context) error {
	if r.rows == nil {
		return nil
	}
	return r.rows.Close()
}
