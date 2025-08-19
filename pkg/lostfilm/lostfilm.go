package lostfilm

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/PuerkitoBio/goquery"
)

type ClientConfig struct {
	HttpClient  HttpClient
	MainPageUrl string
	Cookie      http.Cookie
}

type Client struct {
	Config ClientConfig
	Logger *logrus.Logger
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
	Poster        string
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

func (c Client) GetRoot() ([]RootElement, error) {
	doc, err := c.getDoc(c.Config.MainPageUrl + "/new")
	if err != nil {
		c.Logger.Error(err.Error())
		return nil, err
	}
	rows := doc.Find(".row")
	r := make([]RootElement, 0, 15)
	parseRow := func(i int, row *goquery.Selection) {
		link, foundLink := row.Find("a").Eq(0).Attr("href")
		if !foundLink {
			c.Logger.Debug("Not found link")
			return
		}
		_, episodeNumber := ExtractSeasonEpisode(link)
		posterLink, foundPoster := row.Find(".thumb").Eq(0).Attr("src")
		if !foundPoster {
			posterLink = ""
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

		seriesName := row.Find(".name-ru").Eq(0).Text()
		seriesNameEn := row.Find(".name-en").Eq(0).Text()
		if seriesName == "" || seriesNameEn == "" {
			c.Logger.Debug("Not found series name")
			return
		}

		episodeName := row.Find(".alpha").Eq(0).Text()
		episodeNameEn := row.Find(".beta").Eq(0).Text()

		// episodeNumber > 0 тогда, когда link является сериалом см. ExtractSeasonEpisode
		if episodeNumber > 0 && (episodeName == "" || episodeNameEn == "") {
			c.Logger.Debug("Not found episode name")
			return
		}

		r = append(r, RootElement{
			Page:          link,
			Name:          seriesName,
			NameEN:        seriesNameEn,
			EpisodeName:   episodeName,
			EpisodeNameEn: episodeNameEn,
			Date:          date,
			order:         i,
			Poster:        posterLink,
		})
	}
	rows.Each(parseRow)

	return r, nil
}

func ExtractSeasonEpisode(path string) (int, int) {
	// Регулярка для series: season_X/episode_Y
	reSeries := regexp.MustCompile(`season_(\d+)/episode_(\d+)/`)

	if reSeries.MatchString(path) {
		matches := reSeries.FindStringSubmatch(path)
		if len(matches) == 3 {
			var season, episode int
			fmt.Sscanf(matches[1], "%d", &season)
			fmt.Sscanf(matches[2], "%d", &episode)
			return season, episode
		}
	}

	// Если это movies
	reMovies := regexp.MustCompile(`^/movies/`)
	if reMovies.MatchString(path) {
		return -1, -1
	}

	// Если ничего не подошло
	return -2, -2
}

func (c Client) GetEpisode(page string) (*Episode, error) {
	doc, err := c.getDoc(c.Config.MainPageUrl + page)
	if err != nil {
		c.Logger.Error(err.Error())
		return nil, err
	}
	onClick, found := doc.Find(".external-btn").Attr("onclick")
	if !found {
		return nil, nil
	}

	rawId := strings.Replace(strings.Replace(onClick, "PlayEpisode('", "", -1), "')", "", -1)
	id, err := strconv.ParseInt(rawId, 10, 64)
	if err != nil {
		c.Logger.Error(err.Error())
		return nil, err
	}

	return &Episode{Id: id}, nil
}

func (c Client) GetTorrentRefs(episodeId int64) ([]TorrentRef, error) {
	doc, err := c.getDoc(c.Config.MainPageUrl + "/v_search.php?a=" + strconv.FormatInt(episodeId, 10))
	if err != nil {
		c.Logger.Error(err.Error())
		return nil, err
	}

	trackUrl, exists := doc.Find("meta").Attr("content")
	if !exists {
		return nil, errors.New("track url not exists")
	}
	trackUrl = strings.Replace(trackUrl, "0; url=", "", -1)

	doc, err = c.getDoc(trackUrl)
	if err != nil {
		return nil, err
	}

	r := make([]TorrentRef, 0, 3)

	nameFull := strings.TrimSpace(doc.Find(".inner-box--text").Text())
	nameFull = strings.ReplaceAll(nameFull, "\t\t\t", " ")

	doc.Find(".inner-box--item").Each(func(i int, s *goquery.Selection) {
		quality := strings.TrimSpace(s.Find(".inner-box--label").Text())
		url, exists := s.Find("a").Eq(0).Attr("href")
		if !exists {
			c.Logger.Error("url not exist")
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
		c.Logger.Error(err.Error())
		return nil, err
	}
	return io.ReadAll(body)
}

func (c Client) Listing(ch chan RootElement, interval time.Duration) {
	for {
		c.Logger.Debugf("Read updates from LostFilm")
		elements, _ := c.GetRoot()
		for _, element := range elements {
			ch <- element
		}
		time.Sleep(interval)
	}
}

func (c Client) getRequest(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.Logger.Error(err.Error())
		return nil, err
	}

	if strings.HasPrefix(url, c.Config.MainPageUrl) {
		req.Header.Set("referer", c.Config.MainPageUrl)
	}
	req.Header.Set("Cookie", c.Config.Cookie.Name+"="+c.Config.Cookie.Value)

	res, err := c.Config.HttpClient.Do(req)
	if err != nil {
		c.Logger.Error(err.Error())
		return nil, err
	}

	return res.Body, nil
}

func (c Client) getDoc(url string) (*goquery.Document, error) {
	body, err := c.getRequest(url)
	if err != nil {
		c.Logger.Error(err.Error())
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.Logger.Error(err.Error())
		}
	}(body)

	return goquery.NewDocumentFromReader(body)
}
