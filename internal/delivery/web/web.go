package web

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"makarov.dev/bot/docs"
	"makarov.dev/bot/internal/config"
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

func StartWeb(ctx context.Context) {
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

	lfGroup := r.Group("/lostfilm")
	{
		ctr := LostFilmController{}
		ctr.Add(lfGroup)
	}

	kinozalGroup := r.Group("/kinozal")
	{
		ctr := KinozalController{}
		ctr.Add(kinozalGroup)
	}

	fileGroup := r.Group("/dl")
	{
		ctr := FileController{}
		ctr.Add(fileGroup)
	}

	twitchGroup := r.Group("/twitch")
	{
		ctr := TwitchController{}
		ctr.Add(twitchGroup)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Fatalf("Error while shutdown web %s", err.Error())
	}

	log.Infof("Web stopped")
}
