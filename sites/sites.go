package sites

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/PuerkitoBio/goquery"
)

var Sites = []Site{
	simpleSite{
		"oakland",
		"https://www.oaklandca.gov/boards-commissions/planning-commission/meetings",
		func(doc *goquery.Document) (string, bool) {
			return doc.
				Find("#meetings").
				Find("tbody").
				Find("tr").
				First().
				Find("td").
				Eq(4).
				Find("a").
				Attr("href")
		},
	},
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Site interface {
	Entity() string
	Get(ctx context.Context, client HTTPClient) (Agenda, error)
}

type simpleSite struct {
	entity  string
	baseURL string
	finder  func(*goquery.Document) (string, bool)
}

func (s simpleSite) Entity() string {
	return s.entity
}

func (s simpleSite) Get(ctx context.Context, client HTTPClient) (Agenda, error) {
	agendaURL, err := s.agendaURL(ctx, client)
	if err != nil {
		return Agenda{}, s.siteErr(err)
	}
	resp, err := get(ctx, client, agendaURL)
	if err != nil {
		return Agenda{}, s.siteErr(err)
	}
	return Agenda{
		Entity:  s.entity,
		Name:    path.Base(resp.Request.URL.Path),
		Content: resp.Body,
	}, nil
}

func (s simpleSite) agendaURL(ctx context.Context, client HTTPClient) (string, error) {
	resp, err := get(ctx, client, s.baseURL)
	if err != nil {
		return "", fmt.Errorf("get page: %s", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("baseURL: %s", err)
	}

	latestMeeting, found := s.finder(doc)
	if !found {
		return "", fmt.Errorf("no doc found")
	}

	return latestMeeting, nil
}

func (s simpleSite) siteErr(err error) error {
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
