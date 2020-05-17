package core

import (
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func notification(msg, url, token string, group int64) error {
	bot, err := tb.NewBot(tb.Settings{
		URL:    url,
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 5 * time.Second},
	})
	if err != nil {
		return err
	}

	_, err = bot.Send(&tb.Chat{ID: group}, "```\n"+msg+"\n```", &tb.SendOptions{ParseMode: tb.ModeMarkdownV2})
	return err
}
