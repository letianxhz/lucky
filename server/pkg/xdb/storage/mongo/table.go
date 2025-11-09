package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"lucky/server/pkg/xdb"
)

// Table MongoDB 表实现
type Table struct {
	src      *xdb.Source
	executor *mongo.Collection
}

func (t *Table) Recover(ctx context.Context, commitments []xdb.Commitment) error {
	// 恢复逻辑：从重做日志恢复数据
	// 这里可以添加恢复逻辑
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

		switch lifecycle {
		case xdb.LifecycleNew, xdb.LifecycleNormal:
			// 插入或更新
			filter := bson.M{"_id": t.getDocumentID(commitment)}
			update := bson.M{"$set": data}
			opts := options.Update().SetUpsert(true)
			_, err := t.executor.UpdateOne(ctx, filter, update, opts)
			if err != nil {
				// 记录错误，但继续处理其他记录
				fmt.Printf("Failed to save commitment: %v\n", err)
				continue
			}

		case xdb.LifecycleDeleted:
			// 删除
			filter := bson.M{"_id": t.getDocumentID(commitment)}
			_, err := t.executor.DeleteOne(ctx, filter)
			if err != nil {
				fmt.Printf("Failed to delete commitment: %v\n", err)
				continue
			}
		}
	}

	return true
}

func (t *Table) getDocumentID(commitment xdb.Commitment) string {
	// 使用 Source 的 PKOf 方法获取主键
	pk := commitment.Source().PKOf(commitment)
	if pk != nil {
		return pk.String()
	}
	return ""
}

func (t *Table) Fetch(ctx context.Context, onlyOne bool, pk xdb.PK) (xdb.RecordCursor, error) {
	filter := bson.M{"_id": pk.String()}
	cursor, err := t.executor.Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch from MongoDB")
	}

	return newRecordCursor(cursor, t.src), nil
}

func (t *Table) FetchMulti(ctx context.Context, pks []xdb.PK) (xdb.RecordCursor, error) {
	if len(pks) == 0 {
		return &RecordCursor{cursor: nil, src: t.src}, nil
	}

	ids := make([]string, len(pks))
	for i, pk := range pks {
		ids[i] = pk.String()
	}

	filter := bson.M{"_id": bson.M{"$in": ids}}
	cursor, err := t.executor.Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch multi from MongoDB")
	}

	return newRecordCursor(cursor, t.src), nil
}

func (t *Table) Find(ctx context.Context, filter interface{}) (xdb.RecordCursor, error) {
	cursor, err := t.executor.Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find from MongoDB")
	}

	return newRecordCursor(cursor, t.src), nil
}

// RecordCursor MongoDB 记录游标
type RecordCursor struct {
	cursor *mongo.Cursor
	src    *xdb.Source
}

func newRecordCursor(cursor *mongo.Cursor, src *xdb.Source) *RecordCursor {
	return &RecordCursor{
		cursor: cursor,
		src:    src,
	}
}

func (r *RecordCursor) Next(ctx context.Context) bool {
	if r.cursor == nil {
		return false
	}
	return r.cursor.Next(ctx)
}

func (r *RecordCursor) Decode(val interface{}) error {
	if r.cursor == nil {
		return errors.New("cursor is nil")
	}
	return r.cursor.Decode(val)
}

func (r *RecordCursor) All(ctx context.Context, results interface{}) error {
	if r.cursor == nil {
		return nil
	}
	return r.cursor.All(ctx, results)
}

func (r *RecordCursor) Close(ctx context.Context) error {
	if r.cursor == nil {
		return nil
	}
	return r.cursor.Close(ctx)
}
