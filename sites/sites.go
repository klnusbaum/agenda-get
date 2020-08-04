package sites

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Site interface {
	Get(ctx context.Context, client HTTPClient, today time.Time) (Agenda, error)
}

type SimpleSite struct {
	entity  string
	baseURL string
	finder  func(*goquery.Document, time.Time) (string, error)
}

func (s SimpleSite) Get(ctx context.Context, client HTTPClient, today time.Time) (Agenda, error) {
	agendaURL, err := s.agendaURL(ctx, client, today)
	if err != nil {
		return Agenda{}, s.siteErr(err)
	}
	resp, err := get(ctx, client, agendaURL)
	if err != nil {
		return Agenda{}, s.siteErr(err)
	}
	return Agenda{
		Name:    s.entity,
		Content: resp.Body,
	}, nil
}

func (s SimpleSite) agendaURL(ctx context.Context, client HTTPClient, today time.Time) (string, error) {
	resp, err := get(ctx, client, s.baseURL)
	if err != nil {
		return "", fmt.Errorf("get page: %s", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("baseURL: %s", err)
	}

	agendaURL, err := s.finder(doc, today)
	if err != nil {
		return "", fmt.Errorf("finder: %s", err)
	}

	return agendaURL, nil
}

func (s SimpleSite) siteErr(err error) error {
	return fmt.Errorf("%s: %s\n", s.entity, err)
}

func get(ctx context.Context, client HTTPClient, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "Agenda-Get/3.0")
	return client.Do(req)
}
