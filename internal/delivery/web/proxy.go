package web

import (
	"github.com/gin-gonic/gin"
	intg "makarov.dev/bot/internal/integration/proxy"
)

type ProxyController struct {
}

func (p ProxyController) Add(g *gin.RouterGroup) {
	g.GET("", proxy)
}

//	@Tags		Proxy controller
//	@Param		url							query	string	true	"Url for proxied GET request"
//	@Param		responseHeaderContentType	query	string	false	"override content-type header"
//	@Produce	plain
//	@Router		/proxy [get]
func proxy(ctx *gin.Context) {
	url := ctx.Query("url")
	responseHeaderContentType := ctx.Query("responseHeaderContentType")
	res, header := intg.Get(url)
	if responseHeaderContentType == "" {
		responseHeaderContentType = header.Get("Content-Type")
	}
	ctx.Header("Content-Type", responseHeaderContentType)
	ctx.String(200, res)
}
