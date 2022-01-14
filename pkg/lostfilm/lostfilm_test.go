package lostfilm

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type HttpClientMock struct {
}

func (c *HttpClientMock) Do(req *http.Request) (*http.Response, error) {
	var file *os.File
	if strings.HasPrefix(req.URL.Path, "/series") {
		file, _ = os.Open("./episode_page.thtml")
	}
	switch req.URL.Path {
	case "/new":
		file, _ = os.Open("./root_page.thtml")
	case "/v_search.php":
		file, _ = os.Open("./torrent_ref1.thtml")
	case "/v3/index.php":
		file, _ = os.Open("./torrent_ref2.thtml")
	case "/td.php":
		file, _ = os.Open("./Heels.S01E04.1080p.rus.LostFilm.TV.mkv.torrent")
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bufio.NewReader(file)),
	}, nil
}

func TestGetRoot(t *testing.T) {
	client := getClient()
	r, err := client.GetRoot()
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal(err)
	}
	if len(r) != 15 {
		t.Fatal("Incorrect len")
	}
	for _, e := range r {
		if e.Poster == "" {
			t.Fatal("Empty poster")
		}
	}
}

func TestGetEpisode(t *testing.T) {
	client := getClient()
	r, err := client.GetEpisode("/series/Heels/season_1/episode_4/")
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal(err)
	}
	if r.Id != 611001004 {
		t.Fatal("Incorrect id")
	}
}

func TestGetTorrentRef(t *testing.T) {
	client := getClient()
	r, err := client.GetTorrentRefs(611001004)
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal(err)
	}
	if len(r) != 3 {
		t.Fatal("Incorrect len")
	}
}

func TestGetTorrent(t *testing.T) {
	client := getClient()
	r, err := client.GetTorrent("http://n.tracktor.site/td.php?s=G1RFzE%2F%2FDtWo0CJNsFptuIQwyTICEAoQF8rR%2Fvg0ONBTuhzMHaTPZo372ohX6P99NIGWP5plNOqcVtAh4GPYn9SpAPjkW86gdiqAk6z29yWC%2Bcmqpabd95%2ByeiAb8Rg%2B")
	if err != nil {
		t.Fatal(err)
	}
	if r == nil {
		t.Fatal(err)
	}
}

func TestListing(t *testing.T) {
	ch := make(chan RootElement)
	client := getClient()

	go client.Listing(ch, 1*time.Minute)

	for i := range ch {
		if i.Page != "" {
			return
		}
	}
}

func getClient() Client {
	cfg := ClientConfig{
		HttpClient: &HttpClientMock{},
	}
	return Client{Config: cfg}
}
