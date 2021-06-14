package tbot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/chrishadi/instock/common"
)

const (
	dummyToken  = "api:token"
	dummyChatId = 123
)

type fakeBody struct {
	content []byte
}

func (m fakeBody) Read(p []byte) (int, error) {
	copy(p, m.content)
	return len(m.content), io.EOF
}

func (m fakeBody) Close() error {
	return nil
}

func TestNewGivenNonEmptyTokenThenReturnBot(t *testing.T) {
	opts := BotOptions{http.Post, json.Marshal, common.ReadResponse}
	expected := &Bot{dummyToken, dummyChatId, &opts}

	bot := New(dummyToken, dummyChatId, &opts)

	if bot.token != expected.token ||
		bot.chatId != expected.chatId ||
		bot.opts.HttpPost == nil ||
		bot.opts.JsonMarshal == nil ||
		bot.opts.ReadResponse == nil {
		t.Errorf("Expect %v, got %v", expected, bot)
	}
}

func TestNewGivenEmptyTokenThenReturnNil(t *testing.T) {
	bot := New("", dummyChatId, nil)

	if bot != nil {
		t.Error("Expect bot to be nil, got", bot)
	}
}

func TestApiUrlFor(t *testing.T) {
	command := "getAnything"
	expected := fmt.Sprintf(apiUrlFmt, dummyToken, command)

	bot := New(dummyToken, dummyChatId, &BotOptions{})
	actual := bot.ApiUrlFor(command)

	if actual != expected {
		t.Errorf("Expect %s, got %s", expected, actual)
	}
}

func TestSendMessageGivenHttpPostReturnOk(t *testing.T) {
	resp := http.Response{StatusCode: 200, Body: fakeBody{content: []byte(`{"ok":true,"result":[]}`)}}
	httpPost := func(url, contentType string, body io.Reader) (*http.Response, error) {
		return &resp, nil
	}
	opts := BotOptions{httpPost, json.Marshal, common.ReadResponse}

	bot := New(dummyToken, dummyChatId, &opts)
	err := bot.SendMessage("")

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
}

func TestSendMessageGivenStatusCodeIsNot200(t *testing.T) {
	resp := http.Response{StatusCode: 400, Body: fakeBody{content: []byte(`{"ok":false,"result":[]}`)}}
	httpPost := func(url, contentType string, body io.Reader) (*http.Response, error) {
		return &resp, nil
	}
	opts := BotOptions{httpPost, json.Marshal, common.ReadResponse}

	bot := New(dummyToken, dummyChatId, &opts)
	err := bot.SendMessage("")

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestSendMessageGivenHttpPostReturnError(t *testing.T) {
	httpPost := func(url, contentType string, body io.Reader) (*http.Response, error) {
		return nil, errors.New("")
	}
	opts := BotOptions{httpPost, json.Marshal, nil}

	bot := New(dummyToken, dummyChatId, &opts)
	err := bot.SendMessage("")

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestSendMessageGivenJsonMarshalReturnError(t *testing.T) {
	jsonMarshal := func(interface{}) ([]byte, error) {
		return nil, errors.New("")
	}
	opts := BotOptions{nil, jsonMarshal, nil}

	bot := New(dummyToken, dummyChatId, &opts)
	err := bot.SendMessage("")

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}
