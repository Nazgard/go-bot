package config

import (
	"github.com/umputun/go-flags"
	"os"
)

type Config struct {
	LostFilm LostFilm `group:"LostFilm" env-namespace:"LOSTFILM"`
	Database Database `group:"Database" env-namespace:"DATABASE"`
	Web      Web      `group:"Web" env-namespace:"WEB"`
	Logzio   Logzio   `group:"Logzio" env-namespace:"LOGZIO"`
	Telegram Telegram `group:"Telegram" env-namespace:"TELEGRAM"`
	Twitch   Twitch   `group:"Twitch" env-namespace:"TWITCH"`
	Kinozal  Kinozal  `group:"Kinozal" env-namespace:"KINOZAL"`
}

type LostFilm struct {
	Domain     string `long:"lostfilm-domain" env:"DOMAIN" default:"https://www.lostfilm.win" description:"LostFilm domain"`
	CookieName string `long:"cookie-name" env:"COOKIE_NAME" required:"true" description:"LostFilm cookie name"`
	CookieVal  string `long:"cookie-val" env:"COOKIE_VAL" required:"true" description:"LostFilm cookie val"`
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
	Enable bool   `long:"logzio-enable" env:"ENABLE" description:"Logzio log send enabled"`
	Token  string `long:"logzio-token" env:"TOKEN" description:"Logzio token"`
}

type Telegram struct {
	Enable   bool   `long:"telegram-enable" env:"ENABLE" description:"Telegram integration is enabled"`
	BotToken string `long:"telegram-bot-token" env:"TOKEN" description:"Telegram bot token"`
	Debug    bool   `long:"debug" env:"DEBUG" description:"Telegram debug mode"`
}

type Twitch struct {
	Channels []string `long:"channel" env:"CHANNELS" env-delim:"," description:"Twitch channels to save messages"`
}

type Kinozal struct {
	Domain string `long:"kinozal-domain" env:"DOMAIN" default:"http://kinozal.tv" description:"Kinozal domain"`
	Cookie string `long:"kinozal-cookie" env:"COOKIE" required:"true" description:"Kinozal cookie"`
}

var config = &Config{}

func Init() {
	if _, err := flags.Parse(config); err != nil {
		os.Exit(1)
	}
}

func GetConfig() *Config {
	return config
}
