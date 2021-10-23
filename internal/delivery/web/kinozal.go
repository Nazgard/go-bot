package web

import (
	"github.com/gin-gonic/gin"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/service/kinozal"
	"time"
)

type KinozalController struct {
	Service kinozal.Service
}

func (c *KinozalController) Add(g *gin.RouterGroup) {
	g.GET("rss", c.rss())
}

func (c KinozalController) rss() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		episodes, err := c.Service.LastKinozalEpisodes(ctx)
		if err != nil {
			_ = ctx.AbortWithError(500, err)
		}
		var lastBuildDate string
		if len(episodes) == 0 {
			lastBuildDate = time.Now().Format(dateLayout)
		} else {
			lastBuildDate = episodes[0].Created.Format(dateLayout)
		}
		rss := Rss{
			Version: "1.0",
			Channel: RssChannel{
				Title:         "Свежачок от Kinozal.TV",
				Link:          "http://kinozal.tv/",
				LastBuildDate: lastBuildDate,
				Items:         make([]RssChannelItem, 0),
			},
		}

		for _, episode := range episodes {
			rss.Channel.Items = append(rss.Channel.Items, RssChannelItem{
				Title:   episode.Name,
				Link:    config.GetConfig().Web.Domain + "/dl/" + episode.GridFsId.Hex(),
				PubDate: episode.Created.Format(dateLayout),
				Uid:     episode.Id.Hex(),
			})
		}

		ctx.XML(200, rss)
	}
}
