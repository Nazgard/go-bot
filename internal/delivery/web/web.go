package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/pkg/log"
	"time"
)

func Init() {
	cfg := config.GetConfig().Web

	w := &log.Writer{}
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
				param.ErrorMessage,
			)
		},
		Output: gin.DefaultWriter,
	})

	gin.SetMode(cfg.Mode)
	r := gin.New()
	r.Use(logger, gin.Recovery())

	g := r.Group("/lostfilm")
	go addLostfilm(g)

	err := r.Run(cfg.Addr)
	if err != nil {
		panic(err)
	}
}
