package config

import (
	"context"
	"net"
	"net/http"

	"github.com/Nazgard/logruzio"
	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
	"github.com/umputun/go-flags"
	"golang.org/x/net/proxy"
	"makarov.dev/bot/pkg"
)

type Config struct {
	Debug    bool     `long:"Debug" env:"DEBUG" description:"Debug mode (pprof enabled)"`
	LogLevel string   `long:"Log level" env:"LOG_LEVEL" default:"DEBUG" description:"Log level"`
	LostFilm LostFilm `group:"LostFilm" env-namespace:"LOSTFILM"`
	Database Database `group:"Database" env-namespace:"DATABASE"`
	Web      Web      `group:"Web" env-namespace:"WEB"`
	Logzio   Logzio   `group:"Logzio" env-namespace:"LOGZIO"`
	Telegram Telegram `group:"Telegram" env-namespace:"TELEGRAM"`
	Twitch   Twitch   `group:"Twitch" env-namespace:"TWITCH"`
	Kinozal  Kinozal  `group:"Kinozal" env-namespace:"KINOZAL"`
	Proxy    Proxy    `group:"Proxy" env-namespace:"PROXY"`
}

type LostFilm struct {
	Enable     bool   `long:"lostfilm-enable" env:"ENABLE" description:"LostFilm integration toggle"`
	Domain     string `long:"lostfilm-domain" env:"DOMAIN" default:"https://www.lostfilm.pro" description:"LostFilm domain"`
	CookieName string `long:"cookie-name" env:"COOKIE_NAME" required:"true" description:"LostFilm cookie name"`
	CookieVal  string `long:"cookie-val" env:"COOKIE_VAL" required:"true" description:"LostFilm cookie val"`
	MaxRetries int    `long:"max-retries" env:"MAX_RETRIES" default:"5" required:"true" description:"LostFilm max tries for download torrent"`
}

type Database struct {
	DatabaseName string `long:"name" env:"NAME" default:"bot" description:"Database name"`
	Uri          string `long:"uri" env:"URI" default:"mongodb://localhost:27017/bot" description:"Database uri"`
}

type Web struct {
	Addr   string `long:"addr" env:"ADDR" default:":8080" description:"Web server address"`
	Mode   string `long:"mode" env:"MODE" default:"release" description:"Web server mode"`
	Domain string `long:"web-domain" env:"DOMAIN" default:"http://localhost:8080" description:"Web server domain"`
}

type Logzio struct {
	Host  string `long:"logzio-host" env:"HOST" default:"https://listener-eu.logz.io:8071" description:"Logzio token"`
	Token string `long:"logzio-token" env:"TOKEN" description:"Logzio token"`
}

type Telegram struct {
	Enable                bool   `long:"telegram-enable" env:"ENABLE" description:"Telegram integration is enabled"`
	BotToken              string `long:"telegram-bot-token" env:"TOKEN" description:"Telegram bot token"`
	Debug                 bool   `long:"debug" env:"DEBUG" description:"Telegram debug mode"`
	LostFilmUpdateChannel int64  `long:"telegram-lostfilm-update-channel" default:"-1001079947237" env:"LOSTFILM_UPDATE_CHANNEL" description:"Telegram channel for LostFilm updates"`
}

type Twitch struct {
	TushqaUserIds []string `long:"twitch-tushqa-user-id" env:"TUSHQA_USER_ID" env-delim:"," description:"Twitch Tushqa user ids"`
	Channels      []string `long:"channel" env:"CHANNELS" env-delim:"," description:"Twitch channels to save messages"`
}

type Kinozal struct {
	Enable bool   `long:"kinozal-enable" env:"ENABLE" description:"Kinozal integration toggle"`
	Domain string `long:"kinozal-domain" env:"DOMAIN" default:"http://kinozal.tv" description:"Kinozal domain"`
	Cookie string `long:"kinozal-cookie" env:"COOKIE" required:"true" description:"Kinozal cookie"`
}

type Proxy struct {
	Enable         bool   `long:"proxy-enable" env:"ENABLE" description:"Proxy toggle"`
	Socks5Addr     string `long:"proxy-socks5-addr" env:"ADDR" description:"Socks5 proxy address"`
	Socks5User     string `long:"proxy-socks5-user" env:"USER" description:"Socks5 proxy username"`
	Socks5Password string `long:"proxy-socks5-password" env:"PASSWORD" description:"Socks5 proxy password"`
}

var config = &Config{}

func Init(logger *log.Logger) {
	if _, err := flags.Parse(config); err != nil {
		logger.Fatal(err)
	}
	baseLogger = logger
	logLevel, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	baseLogger.SetLevel(logLevel)
	logger.SetFormatter(&nested.Formatter{})
	if !config.Debug {
		hook, err := logruzio.New(config.Logzio.Host, config.Logzio.Token, "Bot", log.Fields{})
		if err != nil {
			log.Fatal(err)
		}
		logger.AddHook(hook)
	}

	if config.Proxy.Enable {
		var auth proxy.Auth
		if config.Proxy.Socks5User != "" && config.Proxy.Socks5Password != "" {
			auth = proxy.Auth{
				User:     config.Proxy.Socks5User,
				Password: config.Proxy.Socks5Password,
			}
		}
		dealer, err := proxy.SOCKS5("tcp", config.Proxy.Socks5Addr, &auth, proxy.Direct)
		if err != nil {
			log.Errorf("Can't connect to the proxy: %s", err.Error())
		}

		dealContext := func(ctx context.Context, network, address string) (net.Conn, error) {
			return dealer.Dial(network, address)
		}

		tr := &http.Transport{DialContext: dealContext}
		pkg.DefaultHttpClient.Transport = tr
		logger.Infof("Proxy %s enabled", config.Proxy.Socks5Addr)
	}
}

func GetConfig() *Config {
	return config
}
