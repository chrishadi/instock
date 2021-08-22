package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/chrishadi/instock/common"
	"github.com/chrishadi/instock/tbot"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type IngestionResult struct {
	received int
	*FilterResult
}

func Ingest(ctx context.Context, m PubSubMessage) error {
	bot := tbot.New(botHost, botToken, chatId)
	sb := &strings.Builder{}
	defer func() {
		log.Print("Sending bot message...")
		err := sendMessage(bot, sb.String())
		if err != nil {
			log.Print(err)
		} else {
			log.Print("Done")
		}
	}()

	db := ConnectDb(&pgOpts)
	defer db.Close()

	buf, err := getStockJsonFromApi(stockApiUrl)
	if err != nil {
		panicws(sb, err)
	}

	res, err := ingestJson(buf, db)
	if err != nil {
		if res == nil {
			panicws(sb, err)
		} else {
			logws(sb, err)
		}
	}

	lnew := len(res.New)
	msg := fmt.Sprintf("Received: %d, Active: %d, Stale: %d, New: %d", res.received, len(res.Active), len(res.Stale), lnew)
	logws(sb, msg)
	if lnew > 0 {
		codes := extractCodes(res.New)
		logws(sb, strings.Join(codes, " "))
	}

	return err
}

func getStockJsonFromApi(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return common.ReadResponse(resp)
}

func ingestJson(buf []byte, db db) (*IngestionResult, error) {
	var newStocks []Stock

	err := json.Unmarshal(buf, &newStocks)
	if err != nil {
		return nil, err
	}

	stockLastUpdates, err := getStockLastUpdates(db)
	if err != nil {
		return nil, err
	}

	facets, err := filter(newStocks, stockLastUpdates)
	if err != nil {
		return nil, err
	}

	res := IngestionResult{received: len(newStocks), FilterResult: facets}
	if len(res.Active) > 0 {
		err = insertStocks(res.Active, db)
		if err == nil {
			_, err = db.Exec((*StockLastUpdate)(nil), "REFRESH MATERIALIZED VIEW ?TableName")
		}
	}

	return &res, err
}

func getStockLastUpdates(db db) (stockLastUpdates []StockLastUpdate, err error) {
	err = db.Select(&stockLastUpdates)
	return stockLastUpdates, err
}

func insertStocks(stocks []Stock, db db) error {
	ormRes, err := db.Insert(&stocks)
	if err != nil {
		return err
	}

	inserted := ormRes.RowsReturned()
	if inserted < len(stocks) {
		return fmt.Errorf("Inserted %d is less than active stocks", inserted)
	}

	return nil
}

func extractCodes(stocks []Stock) []string {
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

func panicws(b *strings.Builder, v interface{}) {
	s := fmt.Sprint(v)
	b.WriteString(s)
	log.Panic(s)
}

func sendMessage(bot *tbot.Bot, msg string) error {
	if bot == nil || len(msg) == 0 {
		return nil
	}

	return bot.SendMessage(msg)
}
