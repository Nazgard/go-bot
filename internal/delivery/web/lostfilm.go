package web

import (
	"github.com/gin-gonic/gin"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/service/lostfilm"
)

const dateLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

type Rss struct {
	XMLName struct{}   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel RssChannel `xml:"channel"`
}

type RssChannel struct {
	Title         string           `xml:"title"`
	Link          string           `xml:"link"`
	LastBuildDate string           `xml:"lastBuildDate"`
	Items         []RssChannelItem `xml:"item"`
}

type RssChannelItem struct {
	Title        string `xml:"title"`
	Link         string `xml:"link"`
	PubDate      string `xml:"pubDate"`
	Description  string `xml:"description"`
	OriginalDate string `xml:"originalDate"`
	OriginalUrl  string `xml:"originalUrl"`
	Uid          string `xml:"uid"`
}

type LostFilmController struct {
	Service lostfilm.Service
}

func (c *LostFilmController) Add(g *gin.RouterGroup) {
	g.GET("/rss", c.rss())
}

func (c LostFilmController) rss() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		quality := ctx.Query("quality")
		episodes, err := c.Service.LastEpisodes(ctx)
		if err != nil {
			_ = ctx.AbortWithError(500, err)
		}
		rss := Rss{
			Version: "1.0",
			Channel: RssChannel{
				Title:         "Свежачок от LostFilm.TV",
				Link:          "https://www.lostfilm.tv/",
				LastBuildDate: episodes[0].Created.Format(dateLayout),
				Items:         make([]RssChannelItem, 0),
			},
		}

		for _, episode := range episodes {
			for _, file := range episode.ItemFiles {
				if quality != "" {
					if file.Quality != quality {
						continue
					}
				}
				rss.Channel.Items = append(rss.Channel.Items, RssChannelItem{
					Title:        episode.Name + ". " + episode.EpisodeNameFull,
					Link:         config.GetConfig().Web.Domain + "/dl/" + file.GridFsId.Hex(),
					PubDate:      episode.Created.Format(dateLayout),
					Description:  file.Description,
					OriginalDate: episode.Date.Format(dateLayout),
					OriginalUrl:  episode.Page,
					Uid:          episode.Id.Hex(),
				})
			}
		}

		ctx.XML(200, rss)
	}
}
