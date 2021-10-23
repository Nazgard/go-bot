package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"makarov.dev/bot/internal/config"
	"makarov.dev/bot/internal/service"
	"makarov.dev/bot/internal/service/kinozal"
	"makarov.dev/bot/internal/service/lostfilm"
	"makarov.dev/bot/pkg/log"
	"time"
)

type Registry struct {
	FileService service.FileService
	LFService   lostfilm.Service
	KZService   kinozal.Service
}

func (reg *Registry) Init() {
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
				param.ErrorMessage)
		},
		Output: gin.DefaultWriter,
	})

	gin.SetMode(cfg.Mode)
	r := gin.New()
	r.Use(logger, gin.Recovery())

	lfGroup := r.Group("/lostfilm")
	go addLostfilm(lfGroup, reg.LFService)
	kinozalGroup := r.Group("/kinozal")
	go addKinozal(kinozalGroup, reg.KZService)
	fileGroup := r.Group("/dl")
	go addFile(fileGroup, reg.FileService)

	err := r.Run(cfg.Addr)
	if err != nil {
		panic(err)
	}
}
