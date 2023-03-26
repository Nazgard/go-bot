package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"makarov.dev/bot/pkg"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nleeper/goment"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/repository"
)

const (
	dateParseLayout = "2006-01-02"
	day             = time.Hour * 24
)

type ServiceImpl struct {
	HttpClient HttpClient
	mrBot      *tgbotapi.BotAPI
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var onceTelegramService = sync.Once{}
var telegramService TelegramService

func NewTelegramService() *ServiceImpl {
	return &ServiceImpl{
		HttpClient: pkg.DefaultHttpClient,
	}
}

func GetTelegramService() TelegramService {
	if telegramService == nil {
		onceTelegramService.Do(func() {
			telegramService = NewTelegramService()
		})
	}
	return telegramService
}

type telegramLogger struct {
}

func (t *telegramLogger) Println(v ...any) {
	config.GetLogger().Debug(v...)
}

func (t *telegramLogger) Printf(format string, v ...any) {
	config.GetLogger().Debug(v...)
}

func (s *ServiceImpl) Start() {
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
		s.Start()
	}
	s.mrBot = bot
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
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID
		s.route(&msg)

		_, err := bot.Send(msg)
		if err != nil {
			log.Errorf("Error while send telegram message %s", err.Error())
		}
	}
}

func (s *ServiceImpl) route(msg *tgbotapi.MessageConfig) {
	if strings.Contains(msg.Text, "/dd") {
		txt := strings.ReplaceAll(msg.Text, "/dd", "")
		txt = strings.TrimSpace(txt)
		msg.Text = s.ddCmd(txt)
	}
	if strings.Contains(msg.Text, "/add") {
		txt := strings.ReplaceAll(msg.Text, "/add", "")
		txt = strings.TrimSpace(txt)
		msg.Text = s.addCmd(txt)
	}
	if strings.Contains(msg.Text, "/delete") {
		txt := strings.ReplaceAll(msg.Text, "/delete", "")
		txt = strings.TrimSpace(txt)
		msg.Text = s.deleteCmd(txt)
	}
}

func (s *ServiceImpl) ddCmd(txt string) string {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return "Фигня с Location"
	}
	beautifulDay, _ := goment.New(goment.DateTime{
		Year:     2019,
		Month:    int(time.April),
		Day:      5,
		Hour:     19,
		Minute:   30,
		Location: location})
	var from goment.Goment
	var to goment.Goment
	if txt == "" {
		from = *beautifulDay
		to1, _ := goment.New()
		to = *to1
	}
	if txt != "" {
		split := strings.Split(txt, " ")
		if len(split) == 1 {
			parse, err := time.ParseInLocation(dateParseLayout, split[0], location)
			if err != nil {
				return err.Error()
			}
			from1, _ := goment.New(parse)
			to1, _ := goment.New()
			from = *from1
			to = *to1
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
			from1, _ := goment.New(parse1)
			to1, _ := goment.New(parse2)
			from = *from1
			to = *to1
		}
	}

	rawCount := s.duration(from.ToTime(), to.ToTime())

	monthCount := to.Diff(from, "months")
	dayCount := to.Diff(from, "days")
	if monthCount != 0 {
		yearCount := float32(dayCount) / 365.0
		return fmt.Sprintf("%s (~%.2f года)", rawCount, yearCount)
	}

	return rawCount
}

func (s *ServiceImpl) addCmd(txt string) string {
	if strings.Contains(txt, "kinozal") {
		txt := strings.ReplaceAll(txt, "kinozal", "")
		txt = strings.TrimSpace(txt)
		return s.kinozalAddFavorite(txt)
	}
	return ""
}

func (s *ServiceImpl) deleteCmd(txt string) string {
	if strings.Contains(txt, "kinozal") {
		txt := strings.ReplaceAll(txt, "kinozal", "")
		txt = strings.TrimSpace(txt)
		return s.kinozalDeleteFavorite(txt)
	}
	return ""
}

func (s *ServiceImpl) kinozalAddFavorite(txt string) string {
	id, err := strconv.ParseInt(txt, 10, 64)
	if err != nil {
		return err.Error()
	}
	err = GetKinozalService().AddFavorite(id)
	if err != nil {
		return err.Error()
	}
	return "Ok"
}

func (s *ServiceImpl) kinozalDeleteFavorite(txt string) string {
	id, err := strconv.ParseInt(txt, 10, 64)
	if err != nil {
		return err.Error()
	}
	err = GetKinozalService().DeleteFavorite(id)
	if err != nil {
		return err.Error()
	}
	return "Ok"
}

func (s *ServiceImpl) duration(a, b time.Time) string {
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

func (s *ServiceImpl) SendMessageLostFilmChannel(lfItem *repository.Item) error {
	cfg := config.GetConfig()
	if !cfg.Telegram.Enable {
		return nil
	}
	domain := cfg.Web.Domain

	posterRequest, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https:%s", lfItem.Poster), nil)
	if err != nil {
		return err
	}

	response, err := s.HttpClient.Do(posterRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	markups := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0),
	}
	buttons := make([]tgbotapi.InlineKeyboardButton, 0)
	for _, file := range lfItem.ItemFiles {
		url := domain + "/dl/" + file.GridFsId.Hex()
		buttons = append(buttons, tgbotapi.InlineKeyboardButton{
			Text: file.Quality,
			URL:  &url,
		})
	}
	markups.InlineKeyboard = append(markups.InlineKeyboard, buttons)

	msg := tgbotapi.PhotoConfig{
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{
				ChatID:      cfg.Telegram.LostFilmUpdateChannel,
				ReplyMarkup: markups,
			},
			File: tgbotapi.FileReader{
				Name:   "img",
				Reader: response.Body,
				Size:   response.ContentLength,
			},
		},
		Caption: fmt.Sprintf("%s. %s", lfItem.Name, lfItem.EpisodeNameFull),
	}
	_, err = s.mrBot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
