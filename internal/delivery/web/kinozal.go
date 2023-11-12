package web

import (
	"github.com/gin-gonic/gin"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/integration/kinozal"
	"time"
)

type KinozalController struct {
}

func (c *KinozalController) Add(g *gin.RouterGroup) {
	cacheMiddleware := CacheMiddleware(
		200,
		"application/xml; charset=utf-8",
		func(c *gin.Context) string {
			return "kz"
		},
		30*time.Minute,
	)
	g.GET("rss", cacheMiddleware, c.rss())
}

// @Tags Kinozal controller
// @Produce xml
// @Produce json
// @Success 200 {object} Rss
// @Failure 400,500 {object} HTTPError
// @Router /kinozal/rss [get]
func (c *KinozalController) rss() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		episodes, err := kinozal.LastEpisodes(ctx)
		if err != nil {
			NewError(ctx, 500, err)
			return
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
