package channels

import "makarov.dev/bot/pkg/lostfilm"

var LostFilmInputChannel = make(chan lostfilm.RootElement)
