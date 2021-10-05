package lostfilm

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"makarov.dev/bot/pkg/log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultMainPageUrl = "https://www.lostfilm.win"

type ClientConfig struct {
	HttpClient  HttpClient
	MainPageUrl string
	Cookie      http.Cookie
}

type Client struct {
	Config ClientConfig
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RootElement struct {
	Page          string    // /series/Heels/season_1/episode_4/
	Name          string    // Хилы
	NameEN        string    // Heels
	EpisodeName   string    // Рекламный ролик
	EpisodeNameEn string    // Cutting Promos
	Date          time.Time // 2021-09-07 00:00:00 +0000
	order         int
}

type Episode struct {
	Id int64
}

type TorrentRef struct {
	NameFull    string
	Quality     string
	Description string
	TorrentUrl  string
}

type FullItem struct {
	Page            string            `json:"page"`
	Name            string            `json:"name"`
	EpisodeName     string            `json:"episode_name"`
	EpisodeNameFull string            `json:"episode_name_full"`
	Date            time.Time         `json:"date"`
	Created         time.Time         `json:"created"`
	Torrents        []FullItemTorrent `json:"torrents"`
}

type FullItemTorrent struct {
	Quality     string `json:"quality"`
	Description string `json:"description"`
	Torrent     []byte `json:"-"`
}

func NewClient(cookieName, cookieVal string) Client {
	return Client{Config: ClientConfig{
		HttpClient:  &http.Client{Timeout: 30 * time.Second},
		MainPageUrl: defaultMainPageUrl,
		Cookie:      http.Cookie{Name: cookieName, Value: cookieVal},
	}}
}

func (c Client) GetRoot() ([]RootElement, error) {
	doc, err := c.getDoc(c.getMainPageUrl() + "/new")
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	rows := doc.Find(".row")
	r := make([]RootElement, 0, 15)
	parseRow := func(i int, row *goquery.Selection) {
		link, found := row.Find("a").Eq(0).Attr("href")
		if !found {
			log.Debug("Not found link")
			return
		}
		rawDate := strings.Replace(
			row.Find(".alpha").Eq(1).Text(),
			"Дата выхода Ru: ",
			"",
			-1)
		date, err := time.Parse("02.01.2006", rawDate)
		if err != nil {
			date = time.Now()
		}
		r = append(r, RootElement{
			Page:          link,
			Name:          row.Find(".name-ru").Eq(0).Text(),
			NameEN:        row.Find(".name-en").Eq(0).Text(),
			EpisodeName:   row.Find(".alpha").Eq(0).Text(),
			EpisodeNameEn: row.Find(".beta").Eq(0).Text(),
			Date:          date,
			order:         i,
		})
	}
	rows.Each(parseRow)

	return r, nil
}

func (c Client) GetEpisode(page string) (*Episode, error) {
	doc, err := c.getDoc(c.getMainPageUrl() + page)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	onClick, found := doc.Find(".external-btn").Attr("onclick")
	if !found {
		return nil, nil
	}

	rawId := strings.Replace(strings.Replace(onClick, "PlayEpisode('", "", -1), "')", "", -1)
	id, err := strconv.ParseInt(rawId, 10, 64)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &Episode{Id: id}, nil
}

func (c Client) GetTorrentRefs(episodeId int64) ([]TorrentRef, error) {
	doc, err := c.getDoc(c.getMainPageUrl() + "/v_search.php?a=" + strconv.FormatInt(episodeId, 10))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	trackUrl, exists := doc.Find("meta").Attr("content")
	if !exists {
		return nil, errors.New("track url not exists")
	}
	trackUrl = strings.Replace(trackUrl, "0; url=", "", -1)

	doc, err = c.getDoc(trackUrl)

	r := make([]TorrentRef, 0, 3)

	nameFull := strings.TrimSpace(doc.Find(".inner-box--text").Text())
	nameFull = strings.ReplaceAll(nameFull, "\t\t\t", " ")

	doc.Find(".inner-box--item").Each(func(i int, s *goquery.Selection) {
		quality := strings.TrimSpace(s.Find(".inner-box--label").Text())
		url, exists := s.Find("a").Eq(0).Attr("href")
		if !exists {
			log.Error("url not exist")
			return
		}
		description := s.Find(".inner-box--desc").Text()
		r = append(r, TorrentRef{
			NameFull:    nameFull,
			Quality:     quality,
			Description: description,
			TorrentUrl:  url,
		})
	})

	return r, nil
}

func (c Client) GetTorrent(url string) ([]byte, error) {
	body, err := c.getRequest(url)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return ioutil.ReadAll(body)
}

func (c Client) Listing(ch chan RootElement, interval time.Duration) {
	for {
		elements, _ := c.GetRoot()
		for _, element := range elements {
			ch <- element
		}
		time.Sleep(interval)
	}
}

func (c Client) getMainPageUrl() string {
	if c.Config.MainPageUrl != "" {
		return c.Config.MainPageUrl
	}
	return defaultMainPageUrl
}

func (c Client) getRequest(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	req.Header.Set("Cookie", c.Config.Cookie.Name+"="+c.Config.Cookie.Value)

	res, err := c.Config.HttpClient.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return res.Body, nil
}

func (c Client) getDoc(url string) (*goquery.Document, error) {
	body, err := c.getRequest(url)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error(err.Error())
		}
	}(body)

	return goquery.NewDocumentFromReader(body)
}
