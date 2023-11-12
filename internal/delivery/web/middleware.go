package web

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"makarov.dev/bot/internal/config"
	"time"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func CacheMiddleware(okCode int, contentType string, keyGen func(c *gin.Context) string, ex time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig().Redis
		if !cfg.Enable {
			return
		}

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		key := keyGen(c)
		b, err := config.GetRedis().Get(c, key).Bytes()
		if err == nil || err != redis.Nil {
			config.GetLogger().Tracef("Used redis cache for %s", key)
			c.Data(okCode, contentType, b)
			c.Abort()
			return
		}

		c.Next()

		go func() {
			config.GetLogger().Tracef("Set cache %s", key)
			config.GetRedis().Set(context.Background(), key, blw.body.String(), ex)
		}()

	}
}
