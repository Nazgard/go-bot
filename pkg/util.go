package pkg

import (
	"net/http"
	"time"
)

var DefaultHttpClient = &http.Client{
	Timeout: 30 * time.Second,
}
