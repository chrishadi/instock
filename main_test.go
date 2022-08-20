package ingest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/chrishadi/instock/tbot"
	"github.com/go-pg/pg/v10"
	"github.com/kelseyhightower/envconfig"
)

const (
	url        = "url"
	oops       = "oops"
	tbotHost   = "bot-host"
	tbotToken  = "bot:token"
	tbotChatId = 9
	numOfGL    = 5
)

var stockJson = []byte(`[ { "Id": 135, "Name": "A Inc", "Code": "A", "StockSubSectorId": 12, "SubSectorName": "Building Construction", "StockSectorId": 5, "SectorName": "PROPERTY, REAL ESTATE AND BUILDING CONSTRUCTION", "NewSubIndustryId": 114, "NewSubIndustryName": "Building Construction", "NewIndustryId": 58, "NewIndustryName": "Building Construction", "NewSubSectorId": 29, "NewSubSectorName": "Building Construction", "NewSectorId": 10, "NewSectorName": "Infrasstructure", "Last": 1295.0, "PrevClosingPrice": 1285.0, "AdjustedClosingPrice": 1295.0, "AdjustedOpenPrice": 1295.0, "AdjustedHighPrice": 1320.0, "AdjustedLowPrice": 1280.0, "Volume": 31797900.0, "Frequency": 4830.0, "Value": 41137150500.0, "OneDay": 0.00778210, "OneWeek": -0.02631579, "OneMonth": 0.22169811, "ThreeMonth": 0.41530055, "SixMonth": 0.06584362, "OneYear": 0.40760870, "ThreeYear": -0.10380623, "FiveYear": -0.67165314, "TenYear": 3.07232704, "Mtd": 0.18807339, "Ytd": -0.30563003, "Per": 40.37681000, "Pbr": 0.56868000, "Capitalization": 8028867073430.0, "BetaOneYear": 1.85076090, "StdevOneYear": 0.56303772, "PerAnnualized": 46.65755000, "PsrAnnualized": 0.62168000, "PcfrAnnualized": -3.49928000, "LastDate": "2021-10-25T00:00:00", "LastUpdate": "2021-10-25T00:00:00", "Roe": 0.0121884212438363 } ]`)

func TestGetStockJsonFromApiWhenFailShouldHaveError(t *testing.T) {
	ts := httptest.NewUnstartedServer(nil)
	defer ts.Close()

	_, err := getStockJsonFromApi(ts.URL)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestGetStockJsonFromApiWhenSuccessShouldReturnJsonAndNilError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(stockJson)
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	json, err := getStockJsonFromApi(ts.URL)

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
	if string(json) != string(stockJson) {
		t.Errorf("Expect %s, got %s", stockJson, json)
	}
}

type mockStockRepository struct {
	inserted    int
	insertError error
}

func (mock mockStockRepository) Insert(stocks []Stock) (int, error) {
	return mock.inserted, mock.insertError
}

type mockStockLastUpdateRepository struct {
	getResult    []StockLastUpdate
	getError     error
	refreshError error
}

func (mock mockStockLastUpdateRepository) Get() ([]StockLastUpdate, error) {
	return mock.getResult, mock.getError
}

func (mock mockStockLastUpdateRepository) Refresh() error {
	return mock.refreshError
}

func TestIngestJsonGivenBadJsonShouldReturnError(t *testing.T) {
	badJson := []byte("bad-json")

	_, err := ingestJson(badJson, mockStockRepository{}, mockStockLastUpdateRepository{}, numOfGL)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestIngestJsonWhenGetStockLastUpdatesFailShouldReturnError(t *testing.T) {
	stockLastUpdateRepo := mockStockLastUpdateRepository{getError: errors.New("get-last-updates-error")}

	_, err := ingestJson(stockJson, mockStockRepository{}, stockLastUpdateRepo, numOfGL)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestIngestJsonWhenInsertStocksFailShouldReturnError(t *testing.T) {
	stockRepo := mockStockRepository{insertError: errors.New("insert-stocks-error")}

	_, err := ingestJson(stockJson, stockRepo, mockStockLastUpdateRepository{}, numOfGL)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestIngestJsonWhenRefreshStockLastUpdatesFailShouldReturnError(t *testing.T) {
	stockLastUpdateRepo := mockStockLastUpdateRepository{refreshError: errors.New("refresh-mv-error")}

	_, err := ingestJson(stockJson, mockStockRepository{}, stockLastUpdateRepo, numOfGL)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestIngestJsonWhenSuccessShouldReturnNilError(t *testing.T) {
	stockRepo := mockStockRepository{inserted: 1}
	_, err := ingestJson(stockJson, stockRepo, mockStockLastUpdateRepository{}, numOfGL)

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
}

func TestExtractCodesGivenStocksShouldReturnStockCodes(t *testing.T) {
	stocks := []Stock{{Code: "A"}, {Code: "B"}}
	expected := []string{"A", "B"}

	actual := extractCodes(stocks)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %#v, got %#v", expected, actual)
	}
}

func TestLogwbGivenAnErrorShouldWriteItsMessageToBuffer(t *testing.T) {
	sb := &strings.Builder{}

	logwb(errors.New(oops), sb)

	expected := oops + "\n"
	actual := sb.String()

	if actual != expected {
		t.Errorf("Expect %s, got %s", expected, actual)
	}
}

func TestSendMessageGivenEmptyMessageShouldNotSendMessage(t *testing.T) {
	sent := false

	handler := func(w http.ResponseWriter, r *http.Request) {
		sent = true
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	bot := tbot.New(ts.URL, tbotToken, tbotChatId)

	sendMessage("", bot)

	if sent == true {
		t.Error("Expect sent to be false, got true")
	}
}

func TestSendMessageWhenFailShouldNotReturnNilError(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	bot := tbot.New(ts.URL, tbotToken, tbotChatId)

	err := sendMessage("bot-message", bot)

	if err == nil {
		t.Error("Expect error not to be nil, got nil")
	}
}

func TestSendMessageWhenSuccessShouldReturnNilError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	bot := tbot.New(ts.URL, tbotToken, tbotChatId)

	err := sendMessage("bot-message", bot)

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
}

func TestIngest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(stockJson)
	}))

	const stockApiUrlKey = "STOCK_API_URL"
	stockApiUrl := os.Getenv(stockApiUrlKey)
	os.Setenv(stockApiUrlKey, ts.URL)
	defer os.Setenv(stockApiUrlKey, stockApiUrl)

	const dbNameKey = "PG_DATABASE"
	testDbName := os.Getenv("PG_TEST_DATABASE")
	if len(testDbName) > 0 {
		dbName := os.Getenv(dbNameKey)
		os.Setenv(dbNameKey, testDbName)
		defer os.Setenv(dbNameKey, dbName)
		defer cleanUpDb()
	}

	err := Ingest(context.Background(), PubSubMessage{})
	if err != nil {
		t.Fatal(err)
	}
}

func cleanUpDb() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	db := pg.Connect(&pg.Options{
		Addr:     cfg.Pg.Addr,
		Database: cfg.Pg.Database,
		User:     cfg.Pg.User,
		Password: cfg.Pg.Password,
	})
	defer db.Close()

	db.Model((*Stock)(nil)).Exec("TRUNCATE ?TableName RESTART IDENTITY")
	db.Model((*StockLastUpdate)(nil)).Exec("REFRESH MATERIALIZED VIEW ?TableName")
}
