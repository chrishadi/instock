package tbot

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/chrishadi/instock/common"
)

const (
	chatId  = 123
	message = "message"
	token   = "api:token"
)

func TestNewGivenEmptyTokenShouldReturnNil(t *testing.T) {
	bot := New("", chatId, nil)

	if bot != nil {
		t.Error("Expect bot to be nil, got", bot)
	}
}

func TestNewGivenNonEmptyTokenShouldReturnBot(t *testing.T) {
	opts := BotOptions{http.Post}
	expected := &Bot{token, chatId, &opts}

	bot := New(token, chatId, &opts)

	if bot.token != expected.token ||
		bot.chatId != expected.chatId ||
		bot.opts.HttpPost == nil {
		t.Errorf("Expect %v, got %v", expected, bot)
	}
}

func TestSendMessageGivenEmptyMessageShouldReturnError(t *testing.T) {
	bot := New(token, chatId, &BotOptions{})
	err := bot.SendMessage("")

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestSendMessageWhenHttpPostFailShouldReturnError(t *testing.T) {
	httpPost := func(url, contentType string, body io.Reader) (*http.Response, error) {
		return nil, errors.New("http-post-error")
	}
	opts := BotOptions{httpPost}

	bot := New(token, chatId, &opts)
	err := bot.SendMessage(message)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestSendMessageWhenHttpPostOkShouldReturnNil(t *testing.T) {
	httpPost := func(url, contentType string, body io.Reader) (*http.Response, error) {
		resp := http.Response{
			StatusCode: 200,
			Body:       common.MockRespBody{Content: []byte(`{"ok":true,"result":[]}`)}}
		return &resp, nil
	}
	opts := BotOptions{httpPost}

	bot := New(token, chatId, &opts)
	err := bot.SendMessage(message)

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
}
