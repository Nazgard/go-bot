package proxy

import (
	"golang.org/x/text/encoding/charmap"
	"io"
	"makarov.dev/bot/pkg"
	"net/http"
	"strings"
)

const (
	cp1251            = "windows-1251"
	utf8              = "utf-8"
	contentTypeHeader = "Content-Type"
)

// Get exec http get request to specific url
func Get(url string) (string, http.Header) {
	resp, err := pkg.DefaultHttpClient.Get(url)
	if err != nil {
		return err.Error(), nil
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err.Error(), nil
	}
	body := string(bytes)
	contentType := resp.Header.Get(contentTypeHeader)
	if strings.Contains(contentType, cp1251) {
		decoder := charmap.Windows1251.NewDecoder()
		body, err = decoder.String(body)
		if err != nil {
			return err.Error(), nil
		}
		body = strings.Replace(body, cp1251, utf8, -1)
	}
	resp.Header.Set(contentType, strings.Replace(contentType, cp1251, utf8, 1))
	return body, resp.Header
}
