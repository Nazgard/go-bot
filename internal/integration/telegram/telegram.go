package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"makarov.dev/bot/internal/config"
)

const (
	dateParseLayout = "2006-01-02"
	day             = time.Hour * 24
)

var mrBot *tgbotapi.BotAPI
var router = make(map[string]func(txt string) string)

type telegramLogger struct {
}

func (t *telegramLogger) Println(v ...any) {
	config.GetLogger().Debug(v...)
}

func (t *telegramLogger) Printf(format string, v ...any) {
	config.GetLogger().Debug(v...)
}

func Start(ctx context.Context) {
	log := config.GetLogger()
	cfg := config.GetConfig().Telegram
	if !cfg.Enable {
		log.Info("Telegram integration disabled")
		return
	}
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Errorf("Error while connect to telegram %s %s", err.Error(), " retrying in 15 sec")
		time.Sleep(15 * time.Second)
		Start(ctx)
	}
	mrBot = bot
	err = tgbotapi.SetLogger(&telegramLogger{})
	if err != nil {
		log.Errorf("Error while set looger %s", err.Error())
	}

	bot.Debug = cfg.Debug

	log.Infof("Authorized on telegram account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Errorf("Error while get telegram updates %s", err.Error())
	}

	for update := range updates {
		select {
		case <-ctx.Done():
			log.Infof("Telegram background job stopped")
			return
		default:
			if update.Message == nil {
				continue
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID
			route(&msg)

			_, err := bot.Send(msg)
			if err != nil {
				log.Errorf("Error while send telegram message %s", err.Error())
			}
		}

	}
}

func route(msg *tgbotapi.MessageConfig) {
	txt := strings.TrimSpace(msg.Text)
	wordSplit := strings.Split(txt, " ")
	if len(wordSplit) < 1 {
		msg.Text = "empty cmd"
		return
	}
	cmdWithSlash := strings.TrimSpace(wordSplit[0])
	fnc, e := router[cmdWithSlash]
	if !e {
		return
	}

	txt = strings.ReplaceAll(msg.Text, cmdWithSlash, "")
	txt = strings.TrimSpace(txt)
	msg.Text = fnc(txt)
}

func AddRouterFunc(cmd string, fnc func(txt string) string) error {
	_, e := router[cmd]
	if e {
		return fmt.Errorf("router cmd already exist")
	}

	router[cmd] = fnc

	return nil
}

func SendMessage(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	cfg := config.GetConfig()
	if !cfg.Telegram.Enable {
		return tgbotapi.Message{}, nil
	}
	return mrBot.Send(c)
}
