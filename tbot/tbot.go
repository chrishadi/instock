package tbot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/chrishadi/instock/common"
)

type Bot struct {
	host   string
	token  string
	chatId int
}

type SendMessageParams struct {
	ChatId int    `json:"chat_id"`
	Text   string `json:"text"`
}

func New(host, token string, chatId int) *Bot {
	if len(token) == 0 {
		return nil
	}

	_host := strings.Trim(host, "/")
	return &Bot{_host, token, chatId}
}

func (bot Bot) apiUrlFor(command string) string {
	return fmt.Sprintf("%s/bot%s/%s", bot.host, bot.token, command)
}

func (bot Bot) SendMessage(text string) error {
	if len(text) == 0 {
		return errors.New("Not sending empty message.")
	}

	url := bot.apiUrlFor("sendMessage")
	params := SendMessageParams{bot.chatId, text}

	json, _ := json.Marshal(params)
	body := bytes.NewBuffer(json)
	resp, err := http.Post(url, "application/json", body)
	if err == nil {
		defer resp.Body.Close()
		_, err = common.ReadResponse(resp)
	}

	return err
}
