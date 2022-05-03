package web

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"makarov.dev/bot/docs"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/repository"
	"makarov.dev/bot/internal/service"
)

type Controller interface {
	Add(g *gin.RouterGroup)
}

// NewError example
func NewError(ctx *gin.Context, status int, err error) {
	er := HTTPError{
		Code:    status,
		Message: err.Error(),
	}
	ctx.AbortWithStatusJSON(status, er)
}

// HTTPError example
type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}

func StartWeb() {
	cfg := config.GetConfig()
	webCfg := cfg.Web
	log := config.GetLogger()

	docs.SwaggerInfo.BasePath = "/"

	w := log.WriterLevel(logrus.DebugLevel)
	gin.DefaultErrorWriter = w
	gin.DefaultWriter = w
	logger := gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
				param.ClientIP,
				param.TimeStamp.Format(time.RFC1123),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage)
		},
		Output: gin.DefaultWriter,
	})

	gin.SetMode(webCfg.Mode)
	r := gin.New()
	r.Use(logger, gin.Recovery())

	if cfg.Debug {
		pprof.Register(r, "dev/pprof")
	}

	lfService := service.GetLostFilmService()
	kzService := service.GetKinozalService()
	fsService := service.GetFileService()
	tRepository := repository.GetTwitchChatRepository()

	lfGroup := r.Group("/lostfilm")
	{
		ctr := LostFilmController{Service: lfService}
		ctr.Add(lfGroup)
	}

	kinozalGroup := r.Group("/kinozal")
	{
		ctr := KinozalController{Service: kzService}
		ctr.Add(kinozalGroup)
	}

	fileGroup := r.Group("/dl")
	{
		ctr := FileController{FileService: fsService}
		ctr.Add(fileGroup)
	}

	twitchGroup := r.Group("/twitch")
	{
		ctr := TwitchController{Repository: tRepository}
		ctr.Add(twitchGroup)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	err := r.Run(webCfg.Addr)
	if err != nil {
		panic(err)
	}
}
