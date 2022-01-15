package kinozal

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/charmap"
	"makarov.dev/bot/pkg/log"
)

const defaultMainPageUrl = "http://kinozal.tv"

type Client struct {
	Config ClientConfig
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

func NewClient(cookie string) *Client {
	return &Client{ClientConfig{
		HttpClient:  &http.Client{Timeout: 30 * time.Second},
		MainPageUrl: defaultMainPageUrl,
		Cookie:      cookie,
	}}
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
	idStr := strconv.FormatInt(id, 10)
	name, err := c.GetName(id)
	if err != nil {
		return nil, err
	}
	r, err := c.getRequest("http://dl.kinozal.tv/download.php?id=" + idStr)
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
	for {
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
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	req.Header.Set("Cookie", c.Config.Cookie)

	res, err := c.Config.HttpClient.Do(req)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return res.Body, nil
}
