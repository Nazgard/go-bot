package log

import (
	"encoding/json"
	"github.com/logzio/logzio-go"
	"log"
	"makarov.dev/bot/internal/config"
	"strings"
)

type Entry struct {
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

var sender *logzio.LogzioSender

type Writer struct {
}

func (w *Writer) Write(p []byte) (n int, err error) {
	Debug(string(p))
	return len(p), nil
}

func Init() {
	cfg := config.GetConfig().Logzio
	l, err := logzio.New(
		cfg.Token,
		logzio.SetDebug(nil),
		logzio.SetUrl("https://listener-eu.logz.io:8071"),
		logzio.SetInMemoryQueue(true),
		logzio.SetinMemoryCapacity(1024*1024),
	)
	if err != nil {
		panic(err)
	}
	sender = l
}

func Debug(msg ...string) {
	log.Println(msg)
	if !config.GetConfig().Logzio.Enable {
		return
	}
	payload, err := json.Marshal(Entry{
		Message:  strings.Join(msg, " "),
		Severity: "DEBUG",
	})
	if err != nil {
		panic(err)
	}
	err = sender.Send(payload)
	if err != nil {
		log.Println(err)
	}
}

func Info(msg ...string) {
	log.Println(msg)
	if !config.GetConfig().Logzio.Enable {
		return
	}
	payload, err := json.Marshal(Entry{
		Message:  strings.Join(msg, " "),
		Severity: "INFO",
	})
	if err != nil {
		log.Println(err)
		return
	}
	err = sender.Send(payload)
	if err != nil {
		log.Println(err)
	}
}

func Error(msg ...string) {
	log.Println(msg)
	if !config.GetConfig().Logzio.Enable {
		return
	}
	payload, err := json.Marshal(Entry{
		Message:  strings.Join(msg, " "),
		Severity: "ERROR",
	})
	if err != nil {
		log.Println(err)
		return
	}
	err = sender.Send(payload)
	if err != nil {
		log.Println(err)
	}
}
