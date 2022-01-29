package web

import (
	"github.com/gin-gonic/gin"
	"makarov.dev/bot/internal/repository"
)

type TwitchController struct {
	Repository *repository.TwitchChatRepository
}

func (c *TwitchController) Add(g *gin.RouterGroup) {
	g.GET("/messages", c.messages())
}

// @Tags Twitch controller
// @Param channel query string false "Channel filter"
// @Param limit query int false "Message list limit"
// @Produce json
// @Success 200 {array} repository.TwitchChatMessage
// @Failure 400,500 {object} HTTPError
// @Router /twitch/messages [get]
func (c *TwitchController) messages() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		data, err := c.Repository.GetLastMessages(ctx.Query("channel"), ctx.Query("limit"))
		if err != nil {
			NewError(ctx, 500, err)
			return
		}
		ctx.JSON(200, &data)
	}
}
