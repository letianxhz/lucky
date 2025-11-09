package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"

	"lucky/server/pkg/xdb"
)

// Dao MongoDB 数据访问对象
type Dao struct {
	client          *mongo.Client
	transformDBName func(string) string
	dbMap           map[string]*Database
}

func newDao(client *mongo.Client, dbNameTransformer func(string) string) *Dao {
	return &Dao{
		client:          client,
		transformDBName: dbNameTransformer,
		dbMap:           make(map[string]*Database),
	}
}

func (d *Dao) database(dbName string) *Database {
	db, ok := d.dbMap[dbName]
	if !ok {
		actualDbName := dbName
		if d.transformDBName != nil {
			actualDbName = d.transformDBName(dbName)
		}
		db = newDatabase(d.client.Database(actualDbName))
		d.dbMap[dbName] = db
	}
	return db
}

func (d *Dao) Table(src *xdb.Source) xdb.Table {
	// 使用 Namespace 作为数据库名，TableName 作为集合名
	return d.database(src.Namespace).Table(src)
}

// Database MongoDB 数据库
type Database struct {
	mdb      *mongo.Database
	collMap  map[string]*mongo.Collection
	tableMap map[string]*Table
}

func newDatabase(mdb *mongo.Database) *Database {
	return &Database{
		mdb:      mdb,
		collMap:  make(map[string]*mongo.Collection),
		tableMap: make(map[string]*Table),
	}
}

func (db *Database) Collection(tableName string) *mongo.Collection {
	coll, ok := db.collMap[tableName]
	if !ok {
		coll = db.mdb.Collection(tableName)
		db.collMap[tableName] = coll
	}
	return coll
}

func (db *Database) Table(src *xdb.Source) *Table {
	return &Table{
		src:      src,
		executor: db.Collection(src.TableName),
	}
}
