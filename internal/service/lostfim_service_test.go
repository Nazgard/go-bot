package service

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/pkg/lostfilm"
)

type httpClientMock struct{}

func (c *httpClientMock) Do(req *http.Request) (*http.Response, error) {
	var file *os.File
	if strings.HasPrefix(req.URL.Path, "/series") {
		file, _ = os.Open("../../pkg/lostfilm/episode_page.thtml")
	}
	if strings.HasPrefix(req.URL.Path, "/movies") {
		file, _ = os.Open("../../pkg/lostfilm/movie_page.thtml")
	}
	switch req.URL.Path {
	case "/new":
		file, _ = os.Open("../../pkg/lostfilm/root_page.thtml")
	case "/v_search.php":
		file, _ = os.Open("../../pkg/lostfilm/torrent_ref1.thtml")
	case "/v3/index.php":
		file, _ = os.Open("../../pkg/lostfilm/torrent_ref2.thtml")
	case "/td.php":
		file, _ = os.Open("../../pkg/lostfilm/Heels.S01E04.1080p.rus.LostFilm.TV.mkv.torrent")
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bufio.NewReader(file)),
	}, nil
}

type bucketMock struct{}

func (b *bucketMock) OpenDownloadStream(fileID any) (*gridfs.DownloadStream, error) {
	return nil, nil
}

func (b bucketMock) UploadFromStream(filename string, source io.Reader, opts ...*options.UploadOptions) (primitive.ObjectID, error) {
	return primitive.NewObjectID(), nil
}

var client = &lostfilm.Client{Config: lostfilm.ClientConfig{
	HttpClient:  &httpClientMock{},
	MainPageUrl: "",
	Cookie:      http.Cookie{},
}}

type repositoryMock struct{}

var items = make([]repository.Item, 0)

func (r repositoryMock) FindLatest(ctx context.Context) ([]repository.Item, error) {
	return items, nil
}

func (r repositoryMock) Exists(page string) (bool, error) {
	for _, item := range items {
		if item.Page == page {
			return true, nil
		}
	}
	return false, nil
}

func (r repositoryMock) Insert(item *repository.Item) error {
	items = append(items, *item)
	return nil
}

func (r repositoryMock) Update(item *repository.Item) error {
	updated := false
	for idx, i := range items {
		if i.Id == item.Id {
			items[idx] = *item
			updated = true
		}
	}
	if !updated {
		return mongo.ErrNoDocuments
	} else {
		return nil
	}
}

func (r repositoryMock) GetByPage(page string) (*repository.Item, error) {
	for _, item := range items {
		if item.Page == page {
			return &item, nil
		}
	}
	return nil, mongo.ErrNoDocuments
}

var serviceMock = LostFilmServiceImpl{
	Client:     client,
	Repository: repositoryMock{},
	Bucket:     &bucketMock{},
	Telegram:   &telegramServiceMock{},
}

type telegramServiceMock struct {
}

func (t *telegramServiceMock) Start() {
}

func (t *telegramServiceMock) SendMessageLostFilmChannel(lfItem *repository.Item) error {
	return nil
}

var fakeRootElement = lostfilm.RootElement{
	Page:          "/series/Heels/season_1/episode_4/",
	Name:          "????????",
	NameEN:        "Heels",
	EpisodeName:   "?????????????????? ??????????",
	EpisodeNameEn: "Cutting Promos",
	Date:          time.Now(),
	Poster:        "//static.lostfilm.win/Images/611/Posters/image_s1.jpg",
}

func after() {
	items = make([]repository.Item, 0)
}

func TestService_LastEpisodes(t *testing.T) {
	serviceMock.StoreElement(fakeRootElement)
	episodes, err := serviceMock.LastEpisodes(context.Background())
	if err != nil {
		t.Error(err)
	}
	if len(episodes) != 1 {
		t.Errorf("wrong len %d", len(episodes))
	}
	after()
}

func TestService_storeElement(t *testing.T) {
	serviceMock.StoreElement(fakeRootElement)
	if len(items) != 1 {
		t.Error("wrong len")
	}
	if items[0].Poster == "" {
		t.Error("Empty poster")
	}
	after()
}

func TestService_storeMovieElement(t *testing.T) {
	fakeRootElement.Page = "/movies/JurassicWorldDominion"
	serviceMock.StoreElement(fakeRootElement)
	if len(items) != 1 {
		t.Error("wrong len")
	}
	if items[0].Poster == "" {
		t.Error("Empty poster")
	}
	if items[0].EpisodeNameFull != "??????????" {
		t.Error("Wrong episode name full")
	}
	after()
}
