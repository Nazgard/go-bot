# go-bot

[![Go](https://github.com/Nazgard/go-bot/actions/workflows/go.yml/badge.svg)](https://github.com/Nazgard/go-bot/actions/workflows/go.yml)
![Go version](https://img.shields.io/github/go-mod/go-version/Nazgard/go-bot)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Личный бот на Go. Следит за новыми релизами на LostFilm и Kinozal, отдаёт их RSS-лентами,
публикует в Telegram-каналы и Mastodon, а также сохраняет сообщения из чатов Twitch.

## Возможности

- **LostFilm** — парсинг новых серий, скачивание торрентов, RSS-лента, публикация в Telegram.
- **Kinozal** — отслеживание релизов, RSS-лента, публикация в Telegram.
- **Telegram-бот** — команды:
  - `/lf list` (или `/lostfilm list`) — последние релизы LostFilm;
  - `/lf resend <id>` — переотправить релиз в канал;
  - `/dd [дата]` — сколько времени прошло с указанной даты.
- **Twitch** — сохранение сообщений из чатов выбранных каналов и цитат.
- **Mastodon** — кросспостинг обновлений.
- **HTTP API** на Gin со Swagger-документацией.
- Опциональные SOCKS5-прокси, кэш в Redis, логирование в Logz.io.

## HTTP API

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/lostfilm/rss` | RSS-лента LostFilm |
| GET | `/kinozal/rss` | RSS-лента Kinozal |
| GET | `/dl/:fileId` | Скачать сохранённый файл (торрент) |
| GET | `/twitch/messages` | Сообщения из Twitch |
| GET | `/twitch/tushqa` | Цитаты |
| GET | `/proxy` | Проксирование запроса |
| GET | `/swagger/index.html` | Swagger UI |
| GET | `/dev/pprof` | pprof (только при `DEBUG=true`) |

## Быстрый старт

Нужны Go 1.25+ и MongoDB.

```bash
go run ./cmd/app
```

Или через Docker:

```bash
docker build -t go-bot .
docker run --env-file .env -p 8080:8080 go-bot
```

## Конфигурация

Всё настраивается переменными окружения (или соответствующими флагами — см. `--help`).

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `DEBUG` | `false` | Режим отладки, включает pprof |
| `LOG_LEVEL` | `DEBUG` | Уровень логирования |
| `LOCALE` | `ru` | Локаль приложения |
| `DATABASE_URI` | `mongodb://localhost:27017/bot` | Строка подключения к MongoDB |
| `DATABASE_NAME` | `bot` | Имя базы |
| `WEB_ADDR` | `:8080` | Адрес веб-сервера |
| `WEB_MODE` | `release` | Режим Gin |
| `WEB_DOMAIN` | `http://localhost:8080` | Публичный адрес (для ссылок в RSS) |
| `LOSTFILM_ENABLE` | `false` | Включить LostFilm |
| `LOSTFILM_DOMAIN` | `https://www.lostfilm.pro` | Домен LostFilm |
| `LOSTFILM_COOKIE_NAME` / `LOSTFILM_COOKIE_VAL` | — | Cookie авторизации (обязательно) |
| `LOSTFILM_MAX_RETRIES` | `5` | Попыток скачать торрент |
| `KINOZAL_ENABLE` | `false` | Включить Kinozal |
| `KINOZAL_DOMAIN` | `http://kinozal.tv` | Домен Kinozal |
| `KINOZAL_COOKIE` | — | Cookie авторизации (обязательно) |
| `TELEGRAM_ENABLE` | `false` | Включить Telegram-бота |
| `TELEGRAM_TOKEN` | — | Токен бота |
| `TELEGRAM_DEBUG` | `false` | Отладка Telegram API |
| `TELEGRAM_LOSTFILM_UPDATE_CHANNEL` | — | ID канала для LostFilm |
| `TELEGRAM_KINOZAL_UPDATE_CHANNEL` | — | ID канала для Kinozal |
| `TWITCH_CHANNELS` | — | Каналы Twitch через запятую |
| `TWITCH_TUSHQA_USER_ID` | — | ID пользователей для цитат через запятую |
| `MASTODON_ENABLE` | `false` | Включить Mastodon |
| `MASTODON_SERVER`, `MASTODON_EMAIL`, `MASTODON_PASSWORD`, `MASTODON_CLIENT_KEY`, `MASTODON_CLIENT_SECRET`, `MASTODON_ACCESS_TOKEN` | — | Доступ к Mastodon |
| `REDIS_ENABLE` | `false` | Включить Redis |
| `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` | —, —, `0` | Подключение к Redis |
| `PROXY_ENABLE` | `false` | Включить SOCKS5-прокси |
| `PROXY_ADDR`, `PROXY_USER`, `PROXY_PASSWORD` | — | Параметры прокси |
| `LOGZIO_TOKEN` | — | Токен Logz.io (пусто — не отправлять) |
| `LOGZIO_HOST` | `https://listener-eu.logz.io:8071` | Адрес Logz.io |

## Структура

```
cmd/app/               точка входа
internal/background/   фоновые задачи (парсеры, Telegram, Twitch, healthcheck)
internal/config/       конфигурация, логгер, подключения к БД/Redis/Mastodon
internal/delivery/web/ HTTP-контроллеры
internal/integration/  бизнес-логика интеграций
pkg/                   клиенты LostFilm и Kinozal
docs/                  сгенерированный Swagger
```

## Разработка

```bash
go test ./...
swag init -g cmd/app/main.go   # обновить Swagger
```

## Лицензия

[MIT](LICENSE)
