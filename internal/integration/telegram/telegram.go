package telegram

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
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

func configureHttpClient() http.Client {
	c := config.GetConfig()
	if !c.Proxy.Enable {
		return http.Client{}
	}

	var auth *proxy.Auth
	if c.Proxy.Socks5User != "" && c.Proxy.Socks5Password != "" {
		auth = &proxy.Auth{
			User:     c.Proxy.Socks5User,
			Password: c.Proxy.Socks5Password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", c.Proxy.Socks5Addr, auth, proxy.Direct)
	if err != nil {
		log.Errorf("Can't connect to the proxy: %s (continuing without proxy)", err)
		return http.Client{}
	}

	return http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
		},
	}
}

func Start(ctx context.Context) {
	log := config.GetLogger()
	cfg := config.GetConfig().Telegram
	if !cfg.Enable {
		log.Info("Telegram integration disabled")
		return
	}

	httpClient := configureHttpClient()
	bot, err := tgbotapi.NewBotAPIWithClient(cfg.BotToken, &httpClient)
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
