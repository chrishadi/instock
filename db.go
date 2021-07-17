package ingest

import (
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type db interface {
	Select(model interface{}) error
	Insert(model interface{}) (orm.Result, error)
	Exec(model interface{}, query string) (orm.Result, error)
	Close()
}

type pgDb struct {
	pgDb *pg.DB
}

func (db *pgDb) Select(model interface{}) error {
	return db.pgDb.Model(model).Select()
}

func (db *pgDb) Insert(model interface{}) (orm.Result, error) {
	return db.pgDb.Model(model).Insert()
}

func (db *pgDb) Exec(model interface{}, query string) (orm.Result, error) {
	return db.pgDb.Model(model).Exec(query)
}

func (db *pgDb) Close() {
	db.pgDb.Close()
}

func ConnectDb(opts *pg.Options) db {
	return &pgDb{pgDb: pg.Connect(opts)}
}
