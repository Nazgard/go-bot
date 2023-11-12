package config

import (
	"github.com/mattn/go-mastodon"
	"sync"
)

var onceMastodonClient = sync.Once{}
var mastodonClient *mastodon.Client

func NewMastodonClient() *mastodon.Client {
	cfg := GetConfig().Mastodon
	mastodonClient = mastodon.NewClient(&mastodon.Config{
		Server:       cfg.Server,
		ClientID:     cfg.ClientKey,
		ClientSecret: cfg.ClientSecret,
		AccessToken:  cfg.AccessToken,
	})
	return mastodonClient
}

func GetMastodonClient() *mastodon.Client {
	if mastodonClient == nil {
		onceMastodonClient.Do(func() {
			mastodonClient = NewMastodonClient()
		})
	}
	return mastodonClient
}
