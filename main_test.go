package ingest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/chrishadi/instock/tbot"
	"github.com/go-pg/pg/v10/orm"
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

func TestSendMessageWhenFail(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	bot := tbot.New(ts.URL, tbotToken, tbotChatId)

	sendMessage(bot, "bot-message")
}

func TestSendMessageWhenSuccess(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	bot := tbot.New(ts.URL, tbotToken, tbotChatId)

	sendMessage(bot, "bot-message")
}
