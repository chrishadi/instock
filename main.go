package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/chrishadi/instock/common"
	"github.com/chrishadi/instock/tbot"
	"github.com/go-pg/pg/v10"
)

func main() {
	bot := tbot.New(botToken, chatId, &tbot.BotOptions{
		HttpPost:     http.Post,
		JsonMarshal:  json.Marshal,
		ReadResponse: common.ReadResponse,
	})
	sb := &strings.Builder{}
	defer sendMessage(bot, sb)

	newStocks, err := getStocks(stockApiUrl, http.Get, common.ReadResponse, json.Unmarshal)
	if err != nil {
		panicws(sb, err)
	}

	db := pg.Connect(&pgOpts)
	defer db.Close()

	savedStocks, err := queryAllStockCodeAndLastUpdate(db)
	if err != nil {
		panicws(sb, err)
	}

	facets, err := filter(newStocks, savedStocks)
	if err != nil {
		panicws(sb, err)
	}

	inserted := 0
	ormRes, err := db.Model(&facets.Active).Insert()
	if err != nil {
		logws(sb, err)
	} else {
		inserted = ormRes.RowsReturned()
	}

	logwsf(sb, "Received: %d, Active: %d, Inserted: %d\n", len(newStocks), len(facets.Active), inserted)
	logwsf(sb, "New: %d\n", len(facets.New))
	logws(sb, extractCode(facets.New))
	logwsf(sb, "Stale: %d\n", len(facets.Stale))
	logws(sb, extractCode(facets.Stale))
}

func getStocks(url string, httpGet common.HttpGetFn, readResponse common.ReadResponseFn, jsonUnmarshal common.JsonUnmarshalFn) (stocks []Stock, err error) {
	resp, err := httpGet(url)
	if err != nil {
		return stocks, err
	}
	defer resp.Body.Close()

	bytes, err := readResponse(resp)
	if err != nil {
		return stocks, err
	}

	err = jsonUnmarshal(bytes, &stocks)

	return stocks, err
}

func queryAllStockCodeAndLastUpdate(db *pg.DB) (stocks []Stock, err error) {
	err = db.Model(&stocks).
		Column("code").
		ColumnExpr("max(last_update) AS last_update").
		Group("code").
		Select()

	return stocks, err
}

func extractCode(stocks []Stock) []string {
	res := make([]string, len(stocks))
	for i, stock := range stocks {
		res[i] = stock.Code
	}
	return res
}

func logws(b *strings.Builder, v interface{}) {
	s := fmt.Sprintln(v)
	log.Print(s)
	b.WriteString(s)
}

func logwsf(b *strings.Builder, format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	log.Print(s)
	b.WriteString(s)
}

func panicws(b *strings.Builder, v interface{}) {
	s := fmt.Sprint(v)
	b.WriteString(s)
	log.Panic(s)
}

func sendMessage(bot *tbot.Bot, b *strings.Builder) {
	if bot != nil && b.Len() > 0 {
		err := bot.SendMessage(b.String())
		if err != nil {
			log.Print(err)
		}
	}
}
