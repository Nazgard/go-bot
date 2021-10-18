package web

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/service"
	"makarov.dev/bot/internal/service/kinozal"
	"net/http"
	"time"
)

func addKinozal(group *gin.RouterGroup, service kinozal.Service, fileService service.FileService) {
	group.GET("/rss", func(ctx *gin.Context) {
		episodes, err := service.LastKinozalEpisodes(ctx)
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
				Link:    config.GetConfig().Web.Domain + "/kinozal/dl/" + episode.GridFsId.Hex(),
				PubDate: episode.Created.Format(dateLayout),
				Uid:     episode.Id.Hex(),
			})
		}

		ctx.XML(200, rss)
	})

	group.GET("/dl/:fileId", func(ctx *gin.Context) {
		fileId := ctx.Param("fileId")
		if fileId == "" {
			ctx.AbortWithStatus(400)
		}

		objectID, err := primitive.ObjectIDFromHex(fileId)
		if err != nil {
			ctx.AbortWithStatus(400)
		}

		extraHeaders := map[string]string{
			"Content-Disposition": "attachment; filename=" + objectID.Hex() + ".torrent",
		}

		reader, err := fileService.GetFile(&objectID)
		if err != nil {
			_ = ctx.AbortWithError(500, err)
		}
		ctx.DataFromReader(http.StatusOK, reader.GetFile().Length, ".torrent", reader, extraHeaders)
	})
}
