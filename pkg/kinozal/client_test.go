package kinozal

import (
	"bufio"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

type HttpClientMock struct {
}

func (c *HttpClientMock) Do(req *http.Request) (*http.Response, error) {
	var file *os.File
	switch req.URL.Path {
	case "/browse.php":
		file, _ = os.Open("./main_page.html")
	case "/download.php":
		file, _ = os.Open("./[kinozal.tv]id1866821.torrent")
	case "/details.php":
		file, _ = os.Open("./details.html")
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bufio.NewReader(file)),
	}, nil
}

func TestClient_getRoot(t *testing.T) {
	type fields struct {
		Config ClientConfig
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: testing.CoverMode(),
			fields: fields{Config: ClientConfig{
				HttpClient:  &HttpClientMock{},
				MainPageUrl: "http://kinozal.tv",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				Config: tt.fields.Config,
			}
			if ids, err := c.GetRoot(); (err != nil) != tt.wantErr {
				if len(ids) != 50 {
					t.Errorf("getRoot() len(ids) = %v, want %v", len(ids), 50)
				}
				t.Errorf("getRoot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_getTorrent(t *testing.T) {
	file, err := os.Open("./[kinozal.tv]id1866821.torrent")
	if err != nil {
		panic(err)
	}
	type fields struct {
		Config ClientConfig
	}
	type args struct {
		id int64
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Element
		wantErr bool
	}{
		{
			name: testing.CoverMode(),
			fields: fields{Config: ClientConfig{
				HttpClient:  &HttpClientMock{},
				MainPageUrl: "http://kinozal.tv",
			}},
			wantErr: false,
			want:    Element{
				Name:    "Юрий Яковлев. Служу музам и только им! / 2008 / РУ / SATRip",
				Torrent: bytes,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				Config: tt.fields.Config,
			}
			got, err := c.GetElement(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTorrent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("getTorrent() got = %v, want %v", got, tt.want)
			}
			if len(got.Torrent) != len(tt.want.Torrent) {
				t.Errorf("getTorrent() got = %v, want %v", got, tt.want)
			}
		})
	}
}
