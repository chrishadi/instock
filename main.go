package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

func main() {
	newStocks, err := getStocks()
	if err != nil {
		log.Fatalln(err)
	}

	if len(newStocks) == 0 {
		fmt.Println("No data.")
		return
	}

	savedStocks, err := queryAllStockCodeAndLastUpdate()
	if err != nil {
		log.Fatalln(err)
	}

	facets, err := filter(newStocks, savedStocks)
	if err != nil {
		log.Fatalln(err)
	}

	inserted := 0
	ormRes, err := insert(facets.Active)
	if err != nil {
		log.Println(err)
	} else {
		inserted = ormRes.RowsReturned()
	}

	fmt.Printf("Received: %d, Active: %d, Inserted: %d\n", len(newStocks), len(facets.Active), inserted)
	fmt.Printf("New: %d\n", len(facets.New))
	printJson(facets.New)
	fmt.Printf("Stale: %d\n", len(facets.Stale))
	printJson(facets.Stale)
}

func getStocks() ([]Stock, error) {
	var stocks []Stock

	resp, err := http.Get(stockApiUrl)
	if err != nil {
		return stocks, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return stocks, err
	}

	err = json.Unmarshal(bytes, &stocks)

	return stocks, err
}

func queryAllStockCodeAndLastUpdate() ([]Stock, error) {
	var stocks []Stock

	db := pg.Connect(&pgOpts)
	defer db.Close()

	err := db.Model(&stocks).
		Column("code").
		ColumnExpr("max(last_update) AS last_update").
		Group("code").
		Select()

	return stocks, err
}

func insert(stocks []Stock) (orm.Result, error) {
	db := pg.Connect(&pgOpts)
	defer db.Close()

	return db.Model(&stocks).Insert()
}

func printJson(stocks []Stock) {
	if len(stocks) == 0 {
		return
	}

	marshaled, err := json.Marshal(stocks)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(string(marshaled))
	}
}
