package config

import (
	"context"
	"github.com/umputun/go-flags"
	"net"
	"net/http"
	"sync"

	"github.com/nleeper/goment"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
	"makarov.dev/bot/pkg"
)

type Config struct {
	Debug    bool           `long:"Debug" env:"DEBUG" description:"Debug mode (pprof enabled)"`
	LogLevel string         `long:"Log level" env:"LOG_LEVEL" default:"DEBUG" description:"Log level"`
	LostFilm LostFilmConfig `group:"LostFilm" env-namespace:"LOSTFILM"`
	Database DatabaseConfig `group:"Database" env-namespace:"DATABASE"`
	Web      WebConfig      `group:"Web" env-namespace:"WEB"`
	Logzio   LogzioConfig   `group:"Logzio" env-namespace:"LOGZIO"`
	Telegram TelegramConfig `group:"Telegram" env-namespace:"TELEGRAM"`
	Twitch   TwitchConfig   `group:"Twitch" env-namespace:"TWITCH"`
	Kinozal  KinozalConfig  `group:"Kinozal" env-namespace:"KINOZAL"`
	Proxy    ProxyConfig    `group:"Proxy" env-namespace:"PROXY"`
	Locale   string         `long:"Application localization" env:"LOCALE" description:"Application locale. Time print for example" default:"ru"`
	Redis    RedisConfig    `group:"Redis" env-namespace:"REDIS"`
	Mastodon MastodonConfig `group:"Mastodon" env-namespace:"MASTODON"`
}

type LostFilmConfig struct {
	Enable     bool   `long:"lostfilm-enable" env:"ENABLE" description:"LostFilm integration toggle"`
	Domain     string `long:"lostfilm-domain" env:"DOMAIN" default:"https://www.lostfilm.pro" description:"LostFilm domain"`
	CookieName string `long:"cookie-name" env:"COOKIE_NAME" required:"true" description:"LostFilm cookie name"`
	CookieVal  string `long:"cookie-val" env:"COOKIE_VAL" required:"true" description:"LostFilm cookie val"`
	MaxRetries int    `long:"max-retries" env:"MAX_RETRIES" default:"5" required:"true" description:"LostFilm max tries for download torrent"`
}

type DatabaseConfig struct {
	DatabaseName string `long:"name" env:"NAME" default:"bot" description:"Database name"`
	Uri          string `long:"uri" env:"URI" default:"mongodb://localhost:27017/bot" description:"Database uri"`
}

type WebConfig struct {
	Addr   string `long:"addr" env:"ADDR" default:":8080" description:"Web server address"`
	Mode   string `long:"mode" env:"MODE" default:"release" description:"Web server mode"`
	Domain string `long:"web-domain" env:"DOMAIN" default:"http://localhost:8080" description:"Web server domain"`
}

type LogzioConfig struct {
	Host  string `long:"logzio-host" env:"HOST" default:"https://listener-eu.logz.io:8071" description:"Logzio token"`
	Token string `long:"logzio-token" env:"TOKEN" description:"Logzio token"`
}

type TelegramConfig struct {
	Enable                bool   `long:"telegram-enable" env:"ENABLE" description:"Telegram integration is enabled"`
	BotToken              string `long:"telegram-bot-token" env:"TOKEN" description:"Telegram bot token"`
	Debug                 bool   `long:"debug" env:"DEBUG" description:"Telegram debug mode"`
	LostFilmUpdateChannel int64  `long:"telegram-lostfilm-update-channel" default:"-1001079947237" env:"LOSTFILM_UPDATE_CHANNEL" description:"Telegram channel for LostFilm updates"`
	KinozalUpdateChannel  int64  `long:"telegram-kinozal-update-channel" default:"-1001902326052" env:"KINOZAL_UPDATE_CHANNEL" description:"Telegram channel for Kinozal updates"`
}

type TwitchConfig struct {
	TushqaUserIds []string `long:"twitch-tushqa-user-id" env:"TUSHQA_USER_ID" env-delim:"," description:"Twitch Tushqa user ids"`
	Channels      []string `long:"channel" env:"CHANNELS" env-delim:"," description:"Twitch channels to save messages"`
}

type KinozalConfig struct {
	Enable bool   `long:"kinozal-enable" env:"ENABLE" description:"Kinozal integration toggle"`
	Domain string `long:"kinozal-domain" env:"DOMAIN" default:"http://kinozal.tv" description:"Kinozal domain"`
	Cookie string `long:"kinozal-cookie" env:"COOKIE" required:"true" description:"Kinozal cookie"`
}

type ProxyConfig struct {
	Enable         bool   `long:"proxy-enable" env:"ENABLE" description:"Proxy toggle"`
	Socks5Addr     string `long:"proxy-socks5-addr" env:"ADDR" description:"Socks5 proxy address"`
	Socks5User     string `long:"proxy-socks5-user" env:"USER" description:"Socks5 proxy username"`
	Socks5Password string `long:"proxy-socks5-password" env:"PASSWORD" description:"Socks5 proxy password"`
}

type RedisConfig struct {
	Enable   bool   `long:"redis-enable" env:"ENABLE" description:"Redis toggle"`
	Addr     string `long:"redis-addr" env:"ADDR" description:"Redis server address"`
	Password string `long:"redis-password" env:"PASSWORD" description:"Redis server password"`
	DB       int    `long:"redis-db" env:"DB" default:"0" description:"Redis server db"`
}

type MastodonConfig struct {
	Enable       bool   `long:"mastodon-enable" env:"ENABLE" description:"Mastodon integration toggle"`
	Server       string `long:"mastodon-server" env:"SERVER" description:"Mastodon server addr"`
	Email        string `long:"mastodon-email" env:"EMAIL" description:"Mastodon user email"`
	Password     string `long:"mastodon-password" env:"PASSWORD" description:"Mastodon user password"`
	ClientKey    string `long:"mastodon-client-key" env:"CLIENT_KEY" description:"Mastodon client key"`
	ClientSecret string `long:"mastodon-client-secret" env:"CLIENT_SECRET" description:"Mastodon client secret"`
	AccessToken  string `long:"mastodon-access-token" env:"ACCESS_TOKEN" description:"Mastodon access token"`
}

var config *Config
var initOnce sync.Once
var configOnce sync.Once

func Init(logger *log.Logger) {
	initOnce.Do(func() {
		config = &Config{}
		if _, err := flags.Parse(config); err != nil {
			logger.Fatal(err)
		}
		initLogger(logger)
		initProxy()
		initMoment()
	})
}

func GetConfig() *Config {
	if config == nil {
		configOnce.Do(func() {
			if config != nil {
				return
			}
			Init(log.New())
		})
	}
	return config
}

func initProxy() {
	if !config.Proxy.Enable {
		return
	}
	var auth proxy.Auth
	if config.Proxy.Socks5User != "" && config.Proxy.Socks5Password != "" {
		auth = proxy.Auth{
			User:     config.Proxy.Socks5User,
			Password: config.Proxy.Socks5Password,
		}
	}
	dealer, err := proxy.SOCKS5("tcp", config.Proxy.Socks5Addr, &auth, proxy.Direct)
	if err != nil {
		baseLogger.Errorf("Can't connect to the proxy: %s", err.Error())
	}

	dealContext := func(ctx context.Context, network, address string) (net.Conn, error) {
		return dealer.Dial(network, address)
	}

	tr := &http.Transport{DialContext: dealContext}
	pkg.DefaultHttpClient.Transport = tr
	baseLogger.Infof("Proxy %s enabled", config.Proxy.Socks5Addr)
}

func initMoment() {
	goment.SetLocale(config.Locale)
}
