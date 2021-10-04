package web

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/service"
	"makarov.dev/bot/internal/service/lostfilm"
	"net/http"
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

func addLostfilm(group *gin.RouterGroup, service lostfilm.Service, fileService service.FileService) {
	group.GET("/rss", func(ctx *gin.Context) {
		quality := ctx.Query("quality")
		episodes, err := service.LastEpisodes()
		if err != nil {
			ctx.AbortWithError(500, err)
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
					Link:         config.GetConfig().Web.Domain + "/lostfilm/dl/" + file.GridFsId.Hex(),
					PubDate:      episode.Created.Format(dateLayout),
					Description:  file.Description,
					OriginalDate: episode.Date.Format(dateLayout),
					OriginalUrl:  episode.Page,
					Uid:          episode.Id.Hex(),
				})
			}
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
			ctx.AbortWithError(500, err)
		}
		ctx.DataFromReader(http.StatusOK, reader.GetFile().Length, ".torrent", reader, extraHeaders)
	})
}
