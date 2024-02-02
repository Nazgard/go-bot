package proxy

import (
	"testing"
)

func TestProxyGet(t *testing.T) {
	resp, header := Get("https://google.com")
	if resp == "" {
		t.Errorf("empty response")
	}
	if header == nil {
		t.Errorf("nil header")
	}
}
