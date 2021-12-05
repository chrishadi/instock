package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/chrishadi/instock/reader"
	"github.com/chrishadi/instock/tbot"
	"github.com/go-pg/pg/v10"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	StockApiUrl string `required:"true" split_words:"true"`
	Pg          struct {
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
	NumOfGL int `required:"true" envconfig:"num_of_gl" split_words:"true"`
}

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type IngestionResult struct {
	received int
	*AggregateResult
}

func Ingest(ctx context.Context, m PubSubMessage) error {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Panic(err)
	}

	bot := tbot.New(cfg.Bot.Host, cfg.Bot.Token, cfg.Bot.ChatId)
	sb := &strings.Builder{}
	defer sendBufferedMessage(sb, bot)

	pgOpts := pg.Options{
		Network:  cfg.Pg.Network,
		Addr:     cfg.Pg.Addr,
		Database: cfg.Pg.Database,
		User:     cfg.Pg.User,
		Password: cfg.Pg.Password,
	}
	db := pg.Connect(&pgOpts)
	defer db.Close()

	buf, err := getStockJsonFromApi(cfg.StockApiUrl)
	if err != nil {
		panicwb(err, sb)
	}

	stockRepo := PgStockRepository{db: db}
	stockLastUpdateRepo := PgStockLastUpdateRepository{db: db}
	res, err := ingestJson(buf, stockRepo, stockLastUpdateRepo, cfg.NumOfGL)
	if err != nil {
		if res == nil {
			panicwb(err, sb)
		} else {
			logwb(err, sb)
		}
	}

	logResult(res, sb)

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

func ingestJson(buf []byte, stockRepo StockRepository, stockLastUpdateRepo StockLastUpdateRepository, numOfGL int) (*IngestionResult, error) {
	var newStocks []Stock

	err := json.Unmarshal(buf, &newStocks)
	if err != nil {
		return nil, err
	}

	stockLastUpdates, err := stockLastUpdateRepo.Get()
	if err != nil {
		return nil, err
	}

	facets, err := aggregate(newStocks, stockLastUpdates, numOfGL)
	if err != nil {
		return nil, err
	}

	res := IngestionResult{received: len(newStocks), AggregateResult: facets}
	if len(res.Active) > 0 {
		err = insertStocks(res.Active, stockRepo)
		if err == nil {
			err = stockLastUpdateRepo.Refresh()
		}
	}

	return &res, err
}

func insertStocks(stocks []Stock, repo StockRepository) error {
	inserted, err := repo.Insert(stocks)
	if err != nil {
		return err
	}

	if inserted < len(stocks) {
		return fmt.Errorf("Inserted %d is less than active stocks", inserted)
	}

	return nil
}

func logResult(res *IngestionResult, sb *strings.Builder) {
	lnew := len(res.New)
	msg := fmt.Sprintf("Received: %d, Active: %d, Stale: %d, New: %d", res.received, len(res.Active), len(res.Stale), lnew)
	logwb(msg, sb)
	if lnew > 0 {
		codes := extractCodes(res.New)
		logwb(strings.Join(codes, " "), sb)
	}

	if len(res.TopGainers) > 0 {
		logwb("Gainers: "+strings.Join(res.TopGainers, " "), sb)
	}
	if len(res.TopLosers) > 0 {
		logwb("Losers: "+strings.Join(res.TopLosers, " "), sb)
	}
}

func extractCodes(stocks []Stock) []string {
	res := make([]string, len(stocks))
	for i, stock := range stocks {
		res[i] = stock.Code
	}
	return res
}

func logwb(v interface{}, b *strings.Builder) {
	s := fmt.Sprintln(v)
	log.Print(s)
	b.WriteString(s)
}

func panicwb(v interface{}, b *strings.Builder) {
	s := fmt.Sprint(v)
	b.WriteString(s)
	log.Panic(s)
}

func sendBufferedMessage(sb *strings.Builder, bot *tbot.Bot) {
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
