package ingest

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/chrishadi/instock/reader"
	"github.com/chrishadi/instock/tbot"
	"github.com/chrishadi/instock/toplist"
	"github.com/go-pg/pg/v10"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	StockApiUrl string `required:"true" split_words:"true"`
	PG          struct {
		Network      string
		Addr         string
		Database     string `required:"true"`
		TestDatabase string `split_words:"true"`
		User         string `required:"true"`
		Password     string `required:"true"`
	}
	Bot struct {
		Host   string
		Token  string
		ChatId int `split_words:"true"`
	}
	NumOfTopRank int `required:"true" split_words:"true"`
}

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type StockGain struct {
	Code string
	Gain float64
}

type report struct {
	received int
	active   int
	stale    int
	new      []string
	gainers  []string
	losers   []string
}

func Ingest(ctx context.Context, m PubSubMessage) error {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Panic(err)
	}

	bot := tbot.New(cfg.Bot.Host, cfg.Bot.Token, cfg.Bot.ChatId)
	sb := &strings.Builder{}
	defer sendBufferToBot(sb, bot)

	pgOpts := pg.Options{
		Network:  cfg.PG.Network,
		Addr:     cfg.PG.Addr,
		Database: cfg.PG.Database,
		User:     cfg.PG.User,
		Password: cfg.PG.Password,
	}
	db := pg.Connect(&pgOpts)
	defer db.Close()

	buf, err := getStockJsonFromApi(cfg.StockApiUrl)
	if err != nil {
		logwb(err, sb)
	}

	var stocks []Stock
	if err = json.Unmarshal(buf, &stocks); err != nil {
		logwb(err, sb)
		return err
	}

	stockLastUpdateRepo := PGStockLastUpdateRepository{db: db}
	stockLastUpdates, err := stockLastUpdateRepo.Get()
	if err != nil {
		logwb(err, sb)
		return err
	}

	facets, err := aggregate(stocks, stockLastUpdates)
	if err != nil {
		return err
	}

	res := &report{
		received: len(stocks),
		active:   len(facets.Active),
		stale:    len(facets.Stale),
		new:      extractCodes(facets.New),
	}

	if len(facets.Active) > 0 {
		stockRepo := PGStockRepository{db: db}
		if err = ingestStocks(facets.Active, stockRepo, stockLastUpdateRepo); err != nil {
			logwb(err, sb)
		}

		res.gainers, res.losers = getTopStockCodes(facets.Active, cfg.NumOfTopRank)
	}

	logReport(res, sb)

	return err
}

func getStockJsonFromApi(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return reader.ReadResponse(resp)
}

func ingestStocks(stocks []Stock, repo StockRepository, mv StockLastUpdateRepository) error {
	if err := insertStocks(stocks, repo); err != nil {
		return err
	}

	return mv.Refresh()
}

func insertStocks(stocks []Stock, repo StockRepository) error {
	inserted, err := repo.Insert(stocks)
	if err != nil {
		return err
	}

	if inserted < len(stocks) {
		return fmt.Errorf("number of inserted %d is less than active stocks", inserted)
	}

	return nil
}

func getTopStockCodes(stocks []Stock, n int) ([]string, []string) {
	gainers := toplist.New(n, func(a, b interface{}) bool {
		return a.(StockGain).Gain > b.(StockGain).Gain
	})
	losers := toplist.New(n, func(a, b interface{}) bool {
		return a.(StockGain).Gain < b.(StockGain).Gain
	})

	for _, stock := range stocks {
		gain := stock.OneDay
		if gain == 0.0 {
			continue
		}

		var list *toplist.TopList
		if gain > 0.0 {
			list = gainers
		} else {
			list = losers
		}
		list.Add(StockGain{stock.Code, stock.OneDay})
	}

	gainerCodes := extractTopRankCodes(gainers.Elements())
	loserCodes := extractTopRankCodes(losers.Elements())

	return gainerCodes, loserCodes
}

func extractTopRankCodes(ls *list.List) []string {
	codes := make([]string, 0, ls.Len())
	for e := ls.Front(); e != nil; e = e.Next() {
		codes = append(codes, e.Value.(StockGain).Code)
	}
	return codes
}

func extractCodes(stocks []Stock) []string {
	res := make([]string, len(stocks))
	for i, stock := range stocks {
		res[i] = stock.Code
	}
	return res
}

func logReport(res *report, sb *strings.Builder) {
	new := len(res.new)
	msg := fmt.Sprintf("Received: %d, Active: %d, Stale: %d, New: %d", res.received, res.active, res.stale, new)
	logwb(msg, sb)
	if new > 0 {
		codes := res.new
		logwb(strings.Join(codes, " "), sb)
	}

	if len(res.gainers) > 0 {
		logwb("Gainers: "+strings.Join(res.gainers, " "), sb)
	}
	if len(res.losers) > 0 {
		logwb("Losers: "+strings.Join(res.losers, " "), sb)
	}
}

func logwb(v interface{}, b *strings.Builder) {
	s := fmt.Sprintln(v)
	log.Print(s)
	b.WriteString(s)
}

func sendBufferToBot(sb *strings.Builder, bot *tbot.Bot) {
	log.Print("Sending bot message...")
	err := sendMessage(sb.String(), bot)
	if err != nil {
		log.Print(err)
	} else {
		log.Print("Done")
	}
}

func sendMessage(msg string, bot *tbot.Bot) error {
	if bot == nil || len(msg) == 0 {
		return nil
	}

	return bot.SendMessage(msg)
}
