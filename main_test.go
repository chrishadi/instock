package ingest

import (
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/chrishadi/instock/common"
	"github.com/chrishadi/instock/tbot"
	"github.com/go-pg/pg/v10/orm"
)

const (
	url       = "url"
	token     = "bot:token"
	botChatId = 9
	oops      = "oops"
)

var stockJson = []byte(`[{"Code":"A","LastUpdate":"2021-07-10T00:00:00"}]`)

func TestGetStockJsonFromApiWhenHttpGetFailShouldReturnErrorAndNilJson(t *testing.T) {
	httpGetStub := func(url string) (*http.Response, error) {
		return nil, errors.New("http-get-error")
	}

	json, err := getStockJsonFromApi(url, httpGetStub)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
	if json != nil {
		t.Error("Expect json to be nil, got", json)
	}
}

func TestGetStockJsonFromApiWhenSuccessShouldReturnJsonAndNilError(t *testing.T) {
	httpGetStub := func(url string) (*http.Response, error) {
		resp := http.Response{
			StatusCode: 200,
			Body:       common.MockRespBody{Content: []byte(stockJson)},
		}
		return &resp, nil
	}

	json, err := getStockJsonFromApi(url, httpGetStub)

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
	if string(json) != string(stockJson) {
		t.Errorf("Expect %s, got %s", stockJson, json)
	}
}

type mockDB struct {
	selectError  error
	insertResult orm.Result
	insertError  error
	execResult   orm.Result
	execError    error
}

func (db *mockDB) Select(model interface{}) error {
	return db.selectError
}

func (db *mockDB) Insert(model interface{}) (orm.Result, error) {
	return db.insertResult, db.insertError
}

func (db *mockDB) Exec(model interface{}, query string) (orm.Result, error) {
	return db.execResult, db.execError
}

func (db *mockDB) Close() {
}

type mockOrmResult struct{}

func (r mockOrmResult) Model() orm.Model {
	return nil
}

func (r mockOrmResult) RowsAffected() int {
	return 0
}

func (r mockOrmResult) RowsReturned() int {
	return 1
}

func TestIngestJsonGivenBadJsonShouldReturnError(t *testing.T) {
	badJson := []byte("bad-json")

	_, err := ingestJson(badJson, nil)

	if err == nil {
		t.Error("Expect error to be nil")
	}
}

func TestIngestJsonWhenDbSelectFailShouldReturnError(t *testing.T) {
	db := mockDB{selectError: errors.New("db-select-error")}

	_, err := ingestJson(stockJson, &db)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestIngestJsonWhenDbInsertFailShouldReturnError(t *testing.T) {
	db := mockDB{insertError: errors.New("db-insert-error")}

	_, err := ingestJson(stockJson, &db)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestIngestJsonWhenDbExecFailShouldReturnError(t *testing.T) {
	db := mockDB{insertError: errors.New("db-exec-error")}

	_, err := ingestJson(stockJson, &db)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestIngestJsonWhenSuccessShouldReturnNilError(t *testing.T) {
	insertResult := mockOrmResult{}
	_, err := ingestJson(stockJson, &mockDB{insertResult: insertResult})

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
	httpPost := func(url, contentType string, body io.Reader) (*http.Response, error) {
		sent = true
		return nil, nil
	}
	bot := tbot.New(token, botChatId, &tbot.BotOptions{HttpPost: httpPost})

	sendMessage(bot, "")

	if sent == true {
		t.Error("Expect sent to be false, got true")
	}
}

func TestSendMessageWhenHttpPostFailShouldLogError(t *testing.T) {
	httpPost := func(url, contentType string, body io.Reader) (*http.Response, error) {
		return nil, errors.New("http-post-error")
	}
	bot := tbot.New(token, botChatId, &tbot.BotOptions{HttpPost: httpPost})

	sendMessage(bot, "bot-message")
}
