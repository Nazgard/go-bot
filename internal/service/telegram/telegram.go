package telegram

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/internal/service/kinozal"
)

const dateParseLayout = "2006-01-02"
const day = time.Minute * 60 * 24

type Service interface {
	Init()
	SendMessageLostFilmChannel(lfItem *repository.Item) error
}

type ServiceImpl struct {
	HttpClient HttpClient
	mrBot      *tgbotapi.BotAPI
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewTelegramService(httpClient HttpClient) *ServiceImpl {
	return &ServiceImpl{
		HttpClient: httpClient,
	}
}

type telegramLogger struct {
}

func (t *telegramLogger) Println(v ...interface{}) {
	config.GetLogger().Debug(v...)
}

func (t *telegramLogger) Printf(format string, v ...interface{}) {
	config.GetLogger().Debug(v...)
}

func (s *ServiceImpl) Init() {
	log := config.GetLogger()
	cfg := config.GetConfig().Telegram
	if !cfg.Enable {
		return
	}
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Error("Error while connect to telegram ", err.Error(), " retrying in 15 sec")
		time.Sleep(15 * time.Second)
		s.Init()
	}
	s.mrBot = bot
	err = tgbotapi.SetLogger(&telegramLogger{})
	if err != nil {
		log.Error(err)
	}

	bot.Debug = cfg.Debug

	log.Info("Authorized on telegram account ", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Error("Error while get telegram updates", err)
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
			log.Error("Error while send telegram message", err.Error())
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
	beautifulDay := time.Date(2019, time.April, 5, 19, 30, 0, 0, location)
	if txt == "" {
		return s.duration(time.Now(), beautifulDay)
	}
	split := strings.Split(txt, " ")
	if len(split) == 1 {
		parse, err := time.ParseInLocation(dateParseLayout, split[0], location)
		if err != nil {
			return err.Error()
		}
		return s.duration(parse, beautifulDay)
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
		return s.duration(parse1, parse2)
	}
	return "Фигню прислал"
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
	kinozal.AddFavoriteCh <- id
	return "Ok"
}

func (s *ServiceImpl) kinozalDeleteFavorite(txt string) string {
	id, err := strconv.ParseInt(txt, 10, 64)
	if err != nil {
		return err.Error()
	}
	kinozal.DeleteFavoriteCh <- id
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
	log := config.GetLogger()
	cfg := config.GetConfig()
	domain := cfg.Web.Domain

	posterRequest, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https:%s", lfItem.Poster), nil)
	if err != nil {
		return err
	}

	response, err := s.HttpClient.Do(posterRequest)

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
		log.Error(err)
		return err
	}
	return nil
}
