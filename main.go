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
	stocks, err := getStocks()
	if err != nil {
		log.Fatalln(err)
	}

	if len(stocks) == 0 {
		fmt.Println("No data.")
	} else {
		ormRes, err := save(&stocks)
		if err != nil {
			log.Fatalln(err)
		} else {
			fmt.Printf("%d stocks read.\n%d stocks inserted.\n", len(stocks), ormRes.RowsReturned())
		}
	}
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

func save(stocks *[]Stock) (orm.Result, error) {
	db := pg.Connect(&pgOpts)
	defer db.Close()

	return db.Model(stocks).Insert()
}
