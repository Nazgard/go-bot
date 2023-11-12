package config

import (
	"github.com/redis/go-redis/v9"
	"sync"
)

var onceRdb sync.Once
var rdb *redis.Client

func NewRedis() {
	cfg := GetConfig().Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       0,
	})
}

func GetRedis() *redis.Client {
	if rdb == nil {
		onceRdb.Do(func() {
			NewRedis()
		})
	}
	return rdb
}
