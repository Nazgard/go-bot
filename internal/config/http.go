package config

import (
	"net/http"
	"time"
)

const defaultTimeout = 30 * time.Second

var transport = &http.Transport{}

func CreateConfiguredHttpClient() *http.Client {
	return &http.Client{
		Timeout:   defaultTimeout,
		Transport: transport,
	}
}
