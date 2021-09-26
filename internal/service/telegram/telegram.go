package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/pkg/log"
	"strings"
	"time"
)

func Init() {
	cfg := config.GetConfig().Telegram
	if !cfg.Enable {
		return
	}
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		panic(err)
	}

	bot.Debug = cfg.Debug

	log.Info("Authorized on account", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID
		route(&msg)

		_, err := bot.Send(msg)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func route(msg *tgbotapi.MessageConfig) {
	txt := strings.ReplaceAll(msg.Text, "/dd", "")
	txt = strings.TrimSpace(txt)
	if strings.Contains(msg.Text, "/dd") {
		msg.Text = dd(txt)
	}
}

const dateParseLayout = "2006-01-02"

func dd(txt string) string {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return "Фигня с Location"
	}
	beautifulDay := time.Date(2019, time.April, 5, 19, 30, 0, 0, location)
	if txt == "" {
		return duration(time.Now(), beautifulDay)
	}
	split := strings.Split(txt, " ")
	if len(split) == 1 {
		parse, err := time.ParseInLocation(dateParseLayout, split[0], location)
		if err != nil {
			return err.Error()
		}
		return duration(parse, beautifulDay)
	}
	if len(split) == 2 {
		parse1, err := time.Parse(dateParseLayout, split[0])
		if err != nil {
			return err.Error()
		}
		parse2, err := time.Parse(dateParseLayout, split[1])
		if err != nil {
			return err.Error()
		}
		return duration(parse1, parse2)
	}
	return "Фигню прислал"
}

const day = time.Minute * 60 * 24

func duration(a, b time.Time) string {
	d := b.Sub(a)

	if d < 0 {
		d *= -1
	}

	if d < day {
		return d.String()
	}

	n := d / day
	d -= n * day

	if d == 0 {
		return fmt.Sprintf("%dd", n)
	}

	return fmt.Sprintf("%dd%s", n, d)
}
