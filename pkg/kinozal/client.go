package kinozal

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/pkg"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
)

type Client struct {
	Config    ClientConfig
	dlPageUrl string
}

type ClientConfig struct {
	HttpClient  HttpClient
	MainPageUrl string
	Cookie      string
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Element struct {
	Name    string
	Torrent []byte
}

var once = sync.Once{}
var client *Client

func NewClient() *Client {
	cfg := config.GetConfig().Kinozal
	return &Client{Config: ClientConfig{
		HttpClient:  pkg.DefaultHttpClient,
		MainPageUrl: cfg.Domain,
		Cookie:      cfg.Cookie,
	}}
}

func GetDefaultClient() *Client {
	if client == nil {
		once.Do(func() {
			client = NewClient()
		})
	}
	return client
}

func (c Client) GetRoot() ([]int64, error) {
	ids := make([]int64, 0, 50)

	doc, err := c.getDoc(c.Config.MainPageUrl + "/browse.php")
	if err != nil {
		return nil, err
	}
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		attr, exists := s.Find("a").Attr("href")
		if !exists {
			return
		}
		trUrl, err := url.Parse(attr)
		if err != nil {
			return
		}
		rawId := trUrl.Query().Get("id")
		if rawId == "" {
			return
		}
		id, err := strconv.ParseInt(rawId, 10, 64)
		if err != nil {
			return
		}
		ids = append(ids, id)
	})
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (c Client) GetName(id int64) (string, error) {
	idStr := strconv.FormatInt(id, 10)
	doc, err := c.getDoc(c.Config.MainPageUrl + "/details.php?id=" + idStr)
	if err != nil {
		return "", err
	}
	name := doc.Find(".content a").Eq(0).Text()
	decoder := charmap.Windows1251.NewDecoder()
	name, err = decoder.String(name)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (c Client) GetElement(id int64) (*Element, error) {
	if c.dlPageUrl == "" {
		parse, err := url.Parse(c.Config.MainPageUrl)
		if err != nil {
			return nil, err
		}
		c.dlPageUrl = fmt.Sprintf("%s://dl.%s", parse.Scheme, parse.Host)
	}
	idStr := strconv.FormatInt(id, 10)
	name, err := c.GetName(id)
	if err != nil {
		return nil, err
	}
	r, err := c.getRequest(fmt.Sprintf("%s/download.php?id=%s", c.dlPageUrl, idStr))
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return &Element{Name: name, Torrent: bytes}, nil
}

func (c Client) Listing(ch chan int64, interval time.Duration) {
	log := config.GetLogger()
	for {
		log.Debugf("Read updates from Kinozal")
		ids, err := c.GetRoot()
		if err != nil {
			time.Sleep(interval)
			continue
		}
		for _, element := range ids {
			ch <- element
		}
		time.Sleep(interval)
	}
}

func (c Client) getDoc(url string) (*goquery.Document, error) {
	log := config.GetLogger()
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

func (c Client) getRequest(url string) (io.ReadCloser, error) {
	log := config.GetLogger()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	req.Header.Set("cookie", c.Config.Cookie)
	req.Header.Set("referer", c.Config.MainPageUrl)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	res, err := c.Config.HttpClient.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 399 {
		log.Errorf("Error while kinozal GET request. Status code %d. URL %s", res.StatusCode, url)
		return nil, errors.New("wrong status code")
	}

	if strings.Contains(url, "://dl.") {
		ct := res.Header.Get("Content-Type")
		expectedCt := "application/x-bittorrent"
		if ct != expectedCt {
			log.Errorf("Error while kinozal GET request. Wrong content type %s. Expected %s", ct, expectedCt)
			return nil, errors.New("wrong content type")
		}
	}

	return res.Body, nil
}
