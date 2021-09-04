package ingest

import (
	"github.com/go-pg/pg/v10"
)

type StockRepository interface {
	Insert([]Stock) (int, error)
}

type StockLastUpdateRepository interface {
	Get() ([]StockLastUpdate, error)
	Refresh() error
}

type PgStockRepository struct {
	db *pg.DB
}

func (repo PgStockRepository) Insert(stocks []Stock) (int, error) {
	ormResult, err := repo.db.Model(&stocks).Insert()
	if err != nil {
		return 0, err
	}

	return ormResult.RowsReturned(), nil
}

type PgStockLastUpdateRepository struct {
	db *pg.DB
}

func (repo PgStockLastUpdateRepository) Get() (updates []StockLastUpdate, err error) {
	err = repo.db.Model(&updates).Select()
	return updates, err
}

func (repo PgStockLastUpdateRepository) Refresh() error {
	_, err := repo.db.Exec((*StockLastUpdate)(nil), "REFRESH MATERIALIZED VIEW ?TableName")
	return err
}
