package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/integration/lostfilm"
	"makarov.dev/bot/internal/integration/telegram"
)

func init() {
	err := telegram.AddRouterFunc("/lf", lfCmd)
	if err != nil {
		config.GetLogger().Errorf("Error while add telegram LF cmd %s", err.Error())
		return
	}
	err = telegram.AddRouterFunc("/lostfilm", lfCmd)
	if err != nil {
		config.GetLogger().Errorf("Error while add telegram lostfilm cmd %s", err.Error())
		return
	}
}

func lfCmd(txt string) string {
	txt = strings.TrimSpace(txt)
	if txt == "" {
		return lostfilmHelp()
	}

	parts := strings.Fields(txt)
	cmd := parts[0]

	switch cmd {
	case "list":
		return sendLostFilmList()
	case "resend":
		if len(parts) < 2 {
			return "Использование: /lf resend <id>"
		}
		return resendLostFilm(parts[1])
	default:
		return lostfilmHelp()
	}
}

func lostfilmHelp() string {
	return `Команды LostFilm:
/lf list - показать последние релизы
/lf resend <id> - переотправить релиз в канал`
}

func sendLostFilmList() string {
	items, err := lostfilm.FindLatest(context.Background())
	if err != nil {
		return fmt.Sprintf("Ошибка: %s", err.Error())
	}

	if len(items) == 0 {
		return "Нет релизов"
	}

	cfg := config.GetConfig()
	markups := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0),
	}

	count := 0
	for _, item := range items {
		if count >= 10 {
			break
		}
		url := cfg.Web.Domain + "/dl/" + item.ItemFiles[0].GridFsId.Hex()
		btn := tgbotapi.InlineKeyboardButton{
			Text: fmt.Sprintf("%s - %s", item.Name, item.EpisodeNameFull),
			URL:  &url,
		}
		markups.InlineKeyboard = append(markups.InlineKeyboard, []tgbotapi.InlineKeyboardButton{btn})
		count++
	}

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      0,
			ReplyMarkup: markups,
		},
		Text: "Последние релизы LostFilm:",
	}

	_, err = telegram.SendMessage(msg)
	if err != nil {
		return fmt.Sprintf("Ошибка отправки: %s", err.Error())
	}

	return "Отправлено"
}

func resendLostFilm(id string) string {
	objectID, err := strconv.Atoi(id)
	if err != nil {
		return "Неверный ID"
	}

	items, err := lostfilm.FindLatest(context.Background())
	if err != nil {
		return fmt.Sprintf("Ошибка: %s", err.Error())
	}

	if objectID < 0 || objectID >= len(items) {
		return "Релиз не найден"
	}

	item := items[objectID]
	err = lostfilm.ResendToTelegram(&item)
	if err != nil {
		return fmt.Sprintf("Ошибка: %s", err.Error())
	}

	return fmt.Sprintf("Релиз переотправлен: %s - %s", item.Name, item.EpisodeNameFull)
}