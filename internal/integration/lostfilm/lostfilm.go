package lostfilm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mattn/go-mastodon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/integration/telegram"
	"makarov.dev/bot/pkg"
	"makarov.dev/bot/pkg/lostfilm"
	"net/http"
	"strings"
	"time"
)

type Item struct {
	Id              primitive.ObjectID `bson:"_id"`
	Page            string             `bson:"page"`
	Name            string             `bson:"name"`
	EpisodeName     string             `bson:"episode_name"`
	EpisodeNameFull string             `bson:"episode_name_full"`
	Date            time.Time          `bson:"date"`
	Created         time.Time          `bson:"created"`
	ItemFiles       []ItemFile         `bson:"item_files"`
	Poster          string             `bson:"poster"`
	RetryCount      int                `bson:"retry_count"`
}

type ItemFile struct {
	Quality     string             `bson:"quality"`
	Description string             `bson:"description"`
	GridFsId    primitive.ObjectID `bson:"grid_fs_id"`
}

var Client = lostfilm.Client{
	Config: lostfilm.ClientConfig{
		HttpClient:  pkg.DefaultHttpClient,
		MainPageUrl: config.GetConfig().LostFilm.Domain,
		Cookie:      http.Cookie{Name: config.GetConfig().LostFilm.CookieName, Value: config.GetConfig().LostFilm.CookieVal},
	},
	Logger: config.GetLogger(),
}

func StoreElement(element lostfilm.RootElement) {
	cfg := config.GetConfig()
	lfCfg := cfg.LostFilm
	log := config.GetLogger()
	item, err := getByPage(element.Page)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Errorf("Error while get item by page %s", element.Page)
		return
	}
	if item == nil {
		log.Infof("Store LF item %s", element.Page)
	} else {
		log.Infof("Try append torrent %s", element.Page)
	}
	episode, err := Client.GetEpisode(element.Page)
	if err != nil {
		log.Errorf("Error while get episode %s", err.Error())
		return
	}

	refs, err := Client.GetTorrentRefs(episode.Id)
	if err != nil {
		log.Errorf("Error while get episode refs %s", err.Error())
		return
	}

	nameFull := ""
	if strings.HasPrefix(element.Page, "/movies") {
		nameFull = "Фильм"
	}
	itemFiles := make([]ItemFile, 0, 3)

	for _, ref := range refs {
		alreadyExist := false
		if item != nil {
			for _, file := range item.ItemFiles {
				if file.Quality == ref.Quality {
					alreadyExist = true
					break
				}
			}
		}
		if alreadyExist {
			continue
		}

		if cfg.Redis.Enable {
			config.GetRedis().Del(context.Background(), "lf-"+ref.Quality)
		}

		if nameFull == "" {
			nameFull = ref.NameFull
		}
		torrent, err := Client.GetTorrent(ref.TorrentUrl)
		if err != nil {
			log.Errorf("Error while get torrent %s", err.Error())
			return
		}

		objectID, err := config.
			GetBucket().
			UploadFromStream(element.Name+". "+nameFull+".torrent", bytes.NewReader(torrent))
		if err != nil {
			log.Errorf("Error while store torrent %s", err.Error())
			return
		}

		itemFiles = append(itemFiles, ItemFile{
			Quality:     ref.Quality,
			Description: ref.Description,
			GridFsId:    objectID,
		})
	}

	if item != nil {
		item.RetryCount++
		item.ItemFiles = append(item.ItemFiles, itemFiles...)
		err := update(item)
		if err != nil {
			log.Errorf("Error while update item %s %s", item.Id.Hex(), err.Error())
			return
		}
	} else {
		item = &Item{
			Id:              primitive.NewObjectID(),
			Page:            element.Page,
			Name:            element.Name,
			EpisodeName:     element.EpisodeName,
			EpisodeNameFull: nameFull,
			Date:            element.Date,
			Created:         time.Now(),
			ItemFiles:       itemFiles,
			Poster:          element.Poster,
		}
		err = insert(item)
		if err != nil {
			log.Errorf("Error while save item %s", err.Error())
			return
		}
	}
	if len(item.ItemFiles) == 3 || (len(item.ItemFiles) > 0 && item.RetryCount >= lfCfg.MaxRetries) {
		go sendToTelegram(item)
		go sendToMastodon(*item)
	}
}

func FindLatest(ctx context.Context) ([]Item, error) {
	log := config.GetLogger()
	limit := int64(50)
	cursor, err := getCollection().Find(ctx, bson.D{}, &options.FindOptions{
		Sort:  bson.D{{"date", -1}, {"created", -1}},
		Limit: &limit,
	})
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	items := make([]Item, 0)
	err = cursor.All(ctx, &items)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return items, nil
}

func Exists(page string) (bool, error) {
	cfg := config.GetConfig().LostFilm
	item, err := getByPage(page)
	if err != nil {
		return false, err
	}
	if item == nil {
		return false, nil
	}
	return len(item.ItemFiles) >= 3 || item.RetryCount >= cfg.MaxRetries, nil
}

func insert(item *Item) error {
	ctx, cancel := getContext()
	defer cancel()

	_, err := getCollection().InsertOne(ctx, item)
	if err != nil {
		return err
	}

	return nil
}

func update(item *Item) error {
	ctx, cancel := getContext()
	defer cancel()

	_, err := getCollection().UpdateOne(ctx, bson.D{{"_id", item.Id}}, bson.M{"$set": item})
	if err != nil {
		return err
	}

	return nil
}

func getByPage(page string) (*Item, error) {
	ctx, cancel := getContext()
	defer cancel()

	result := getCollection().FindOne(ctx, bson.D{{"page", page}})
	if result.Err() != nil {
		if result.Err() != mongo.ErrNoDocuments {
			return nil, result.Err()
		} else {
			return nil, nil
		}
	}
	item := Item{}
	err := result.Decode(&item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func getCollection() *mongo.Collection {
	return config.GetDatabase().Collection("lostfilm_items")
}

func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func sendToTelegram(item *Item) {
	cfg := config.GetConfig()
	domain := cfg.Web.Domain

	posterRequest, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https:%s", item.Poster), nil)
	if err != nil {
		return
	}

	response, err := pkg.DefaultHttpClient.Do(posterRequest)
	if err != nil {
		return
	}
	defer response.Body.Close()

	markups := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0),
	}
	buttons := make([]tgbotapi.InlineKeyboardButton, 0)
	for _, file := range item.ItemFiles {
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
		Caption: fmt.Sprintf("%s. %s", item.Name, item.EpisodeNameFull),
	}

	_, err = telegram.SendMessage(msg)
	if err != nil {
		config.GetLogger().Errorf("%s (channel id %d) %s",
			"Error while send lostfilm item to telegram channel",
			config.GetConfig().Telegram.LostFilmUpdateChannel,
			err.Error(),
		)
		return
	}
}

func sendToMastodon(item Item) {
	if len(item.ItemFiles) <= 0 {
		return
	}
	client := config.GetMastodonClient()
	posterRequest, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https:%s", item.Poster), nil)
	if err != nil {
		config.GetLogger().Errorf("Error while download poster for Mastodon status %s", err.Error())
		return
	}
	response, err := client.Client.Do(posterRequest)
	if err != nil {
		config.GetLogger().Errorf("Error while download poster for Mastodon status %s", err.Error())
		return
	}
	defer response.Body.Close()
	imgBytes, err := io.ReadAll(response.Body)
	if err != nil {
		config.GetLogger().Errorf("Error while download poster for Mastodon status %s", err.Error())
		return
	}
	attachment, err := client.UploadMediaFromBytes(context.Background(), imgBytes)
	if err != nil {
		config.GetLogger().Errorf("Error while upload media for Mastodon status %s", err.Error())
		return
	}

	status := fmt.Sprintf("%s. %s\n", item.Name, item.EpisodeNameFull)
	domain := config.GetConfig().Web.Domain
	for _, f := range item.ItemFiles {
		status += fmt.Sprintf("\n%s %s/dl/%s", f.Quality, domain, f.GridFsId.Hex())
	}
	_, err = client.PostStatus(context.Background(), &mastodon.Toot{
		Status:     status,
		Visibility: mastodon.VisibilityPublic,
		Language:   "ru",
		MediaIDs:   []mastodon.ID{attachment.ID},
	})
	if err != nil {
		config.GetLogger().Errorf("Error while send Mastodon status %s", err.Error())
	} else {
		config.GetLogger().Debugf("Posted status to Mastodon for item %s", item.Page)
	}
}
