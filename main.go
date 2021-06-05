package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-pg/pg/v10"
)

func main() {
	newStocks, err := getStocks()
	if err != nil {
		log.Panic(err)
	}

	db := pg.Connect(&pgOpts)
	defer db.Close()
	savedStocks, err := queryAllStockCodeAndLastUpdate(db)
	if err != nil {
		log.Panic(err)
	}

	facets, err := filter(newStocks, savedStocks)
	if err != nil {
		log.Panic(err)
	}

	inserted := 0
	ormRes, err := db.Model(&facets.Active).Insert()
	if err != nil {
		log.Print(err)
	} else {
		inserted = ormRes.RowsReturned()
	}

	log.Printf("Received: %d, Active: %d, Inserted: %d\n", len(newStocks), len(facets.Active), inserted)
	log.Printf("New: %d\n", len(facets.New))
	logJson(facets.New)
	log.Printf("Stale: %d\n", len(facets.Stale))
	logJson(facets.Stale)
}

func getStocks() ([]Stock, error) {
	var stocks []Stock

	resp, err := http.Get(stockApiUrl)
	if err != nil {
		return stocks, err
	}
	defer resp.Body.Close()

	bytes, err := readResponse(resp)
	if err != nil {
		return stocks, err
	}

	err = json.Unmarshal(bytes, &stocks)

	return stocks, err
}

func readResponse(resp *http.Response) ([]byte, error) {
	buffer, err := ioutil.ReadAll(resp.Body)
	if err == nil && resp.StatusCode != 200 {
		err = fmt.Errorf("Status: %d, Body: %s", resp.StatusCode, buffer)
	}

	return buffer, err
}

func queryAllStockCodeAndLastUpdate(db *pg.DB) ([]Stock, error) {
	var stocks []Stock

	err := db.Model(&stocks).
		Column("code").
		ColumnExpr("max(last_update) AS last_update").
		Group("code").
		Select()

	return stocks, err
}

func logJson(stocks []Stock) {
	if len(stocks) == 0 {
		return
	}

	marshaled, err := json.Marshal(stocks)
	if err != nil {
		log.Print(err)
	} else {
		log.Print(string(marshaled))
	}
}
