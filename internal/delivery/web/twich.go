package web

import (
	"github.com/gin-gonic/gin"
	"makarov.dev/bot/internal/repository"
)

type TwichController struct {
	Repository *repository.TwitchChatRepository
}

func (c *TwichController) Add(g *gin.RouterGroup) {
	g.GET("/messages", func(ctx *gin.Context) {

		data, err := c.Repository.GetLastMessages(ctx.Query("channel"), ctx.Query("limit"))
		if err != nil {
			NewError(ctx, 500, err)
			return
		}
		ctx.JSON(200, &data)

	})
}