package commands

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

	var sb strings.Builder
	sb.WriteString("Последние релизы LostFilm:\n\n")

	count := 0
	for _, item := range items {
		if count >= 10 {
			break
		}
		sb.WriteString(fmt.Sprintf("%s - %s\n", item.Id.Hex(), item.Name))
		count++
	}

	return sb.String()
}

func resendLostFilm(id string) string {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "Неверный ID"
	}

	item, err := lostfilm.GetByID(objID)
	if err != nil {
		return fmt.Sprintf("Ошибка: %s", err.Error())
	}

	if item == nil {
		return "Релиз не найден"
	}

	err = lostfilm.ResendToTelegram(item)
	if err != nil {
		return fmt.Sprintf("Ошибка: %s", err.Error())
	}

	return fmt.Sprintf("Релиз переотправлен: %s - %s", item.Name, item.EpisodeNameFull)
}