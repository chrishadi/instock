package tbot

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	host    = "host"
	chatId  = 123
	message = "message"
	token   = "api:token"
)

func TestNewGivenEmptyTokenShouldReturnNil(t *testing.T) {
	bot := New(host, "", chatId)

	if bot != nil {
		t.Error("Expect bot to be nil, got", bot)
	}
}

func TestNewGivenNonEmptyTokenShouldReturnBot(t *testing.T) {
	bot := New(host, token, chatId)

	if bot.host != host ||
		bot.token != token ||
		bot.chatId != chatId {
		expected := &Bot{host, token, chatId}
		t.Errorf("Expect %v, got %v", expected, bot)
	}
}

func TestNewGivenHostWithTrailingSlashShouldRemoveIt(t *testing.T) {
	bot := New("host/", token, chatId)

	if bot.host != "host" {
		t.Errorf("Expect host, got %s", bot.host)
	}
}

func TestSendMessageGivenEmptyMessageShouldReturnError(t *testing.T) {
	bot := New(host, token, chatId)
	err := bot.SendMessage("")

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestSendMessageWhenHttpPostFailShouldReturnError(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	bot := New(ts.URL, token, chatId)
	err := bot.SendMessage(message)

	if err == nil {
		t.Error("Expect error not to be nil")
	}
}

func TestSendMessageWhenHttpPostOkShouldReturnNil(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":true,"result":[]}`))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	bot := New(ts.URL, token, chatId)
	err := bot.SendMessage(message)

	if err != nil {
		t.Error("Expect error to be nil, got", err)
	}
}
