package sites

import (
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type SimpleSite struct {
	Entity  string
	BaseURL string
	OutExt  string
	Finder  func(*goquery.Document, time.Time) (string, error)
}

type Agenda struct {
	Name    string
	Content io.ReadCloser
}
