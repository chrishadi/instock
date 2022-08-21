package ingest

import "github.com/go-pg/pg/v10"

type StockRepository interface {
	Insert([]Stock) (int, error)
}

type StockLastUpdateRepository interface {
	Get() ([]StockLastUpdate, error)
	Refresh() error
}

type PGStockRepository struct {
	db *pg.DB
}

func (repo PGStockRepository) Insert(stocks []Stock) (int, error) {
	ormResult, err := repo.db.Model(&stocks).Insert()
	if err != nil {
		return 0, err
	}

	return ormResult.RowsAffected(), nil
}

type PGStockLastUpdateRepository struct {
	db *pg.DB
}

func (repo PGStockLastUpdateRepository) Get() (updates []StockLastUpdate, err error) {
	err = repo.db.Model(&updates).Select()
	return updates, err
}

func (repo PGStockLastUpdateRepository) Refresh() error {
	_, err := repo.db.Model((*StockLastUpdate)(nil)).Exec("REFRESH MATERIALIZED VIEW ?TableName")
	return err
}
