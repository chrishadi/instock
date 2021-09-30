package ingest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/chrishadi/instock/tbot"
)

const (
	url        = "url"
	oops       = "oops"
	tbotHost   = "bot-host"
	tbotToken  = "bot:token"
	tbotChatId = 9
)

var stockJson = []byte(`[{"Code":"A","LastUpdate":"2021-07-10T00:00:00"}]`)

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

	_, err := ingestJson(badJson, mockStockRepository{}, mockStockLastUpdateRepository{})

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestIngestJsonWhenGetStockLastUpdatesFailShouldReturnError(t *testing.T) {
	stockLastUpdateRepo := mockStockLastUpdateRepository{getError: errors.New("get-last-updates-error")}

	_, err := ingestJson(stockJson, mockStockRepository{}, stockLastUpdateRepo)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestIngestJsonWhenInsertStocksFailShouldReturnError(t *testing.T) {
	stockRepo := mockStockRepository{insertError: errors.New("insert-stocks-error")}

	_, err := ingestJson(stockJson, stockRepo, mockStockLastUpdateRepository{})

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestIngestJsonWhenRefreshStockLastUpdatesFailShouldReturnError(t *testing.T) {
	stockLastUpdateRepo := mockStockLastUpdateRepository{refreshError: errors.New("refresh-mv-error")}

	_, err := ingestJson(stockJson, mockStockRepository{}, stockLastUpdateRepo)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestIngestJsonWhenSuccessShouldReturnNilError(t *testing.T) {
	stockRepo := mockStockRepository{inserted: 1}
	_, err := ingestJson(stockJson, stockRepo, mockStockLastUpdateRepository{})

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

func TestLogWsGivenAnErrorShouldWriteItsMessageToBuffer(t *testing.T) {
	sb := strings.Builder{}

	logws(&sb, errors.New(oops))

	expected := oops + "\n"
	actual := sb.String()

	if actual != expected {
		t.Errorf("Expect %s, got %s", expected, actual)
	}
}

func TestPanicWsGivenAnErrorShouldWriteItsMessageToBuffer(t *testing.T) {
	sb := strings.Builder{}

	defer func() {
		recover()
	}()

	panicws(&sb, errors.New(oops))

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

	sendMessage(bot, "")

	if sent == true {
		t.Error("Expect sent to be false, got true")
	}
}

func TestSendMessageWhenFailShouldNotReturnNilError(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	bot := tbot.New(ts.URL, tbotToken, tbotChatId)

	err := sendMessage(bot, "bot-message")

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

	err := sendMessage(bot, "bot-message")

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
}
