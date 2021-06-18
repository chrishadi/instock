package tbot

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/chrishadi/instock/common"
)

const (
	apiUrlFmt   = "https://api.telegram.org/bot%s/%s"
	contentType = "application/json"
	sendMessage = "sendMessage"
)

type BotOptions struct {
	HttpPost     common.HttpPostFn
	JsonMarshal  common.JsonMarshalFn
	ReadResponse common.ReadResponseFn
}

type Bot struct {
	token  string
	chatId int
	opts   *BotOptions
}

type SendMessageParams struct {
	ChatId int    `json:"chat_id"`
	Text   string `json:"text"`
}

func New(token string, chatId int, options *BotOptions) *Bot {
	if len(token) == 0 {
		return nil
	}
	return &Bot{token, chatId, options}
}

func (bot Bot) ApiUrlFor(command string) string {
	return fmt.Sprintf(apiUrlFmt, bot.token, command)
}

func (bot Bot) SendMessage(text string) error {
	if len(text) == 0 {
		return errors.New("Not sending empty message.")
	}

	url := bot.ApiUrlFor(sendMessage)
	msg := SendMessageParams{bot.chatId, text}

	jsonb, err := bot.opts.JsonMarshal(msg)
	if err != nil {
		return err
	}

	body := bytes.NewBuffer(jsonb)
	resp, err := bot.opts.HttpPost(url, contentType, body)
	if err == nil {
		defer resp.Body.Close()
		_, err = bot.opts.ReadResponse(resp)
	}

	return err
}
