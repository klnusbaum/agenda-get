package sites

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Fetcher interface {
	Simple(ctx context.Context, site SimpleSite) (Agenda, error)
}

type DefaultFetcher struct {
	client HTTPClient
	today  time.Time
}

func NewDefaultFetcher(client HTTPClient, today time.Time) DefaultFetcher {
	return DefaultFetcher{
		client: client,
		today:  today,
	}
}

func (f DefaultFetcher) Simple(ctx context.Context, site SimpleSite) (Agenda, error) {
	agendaURL, err := f.agendaURL(ctx, site)
	if err != nil {
		return Agenda{}, fmt.Errorf("%s: %w\n", site.Entity, err)
	}
	resp, err := f.get(ctx, agendaURL)
	if err != nil {
		return Agenda{}, fmt.Errorf("%s: %w\n", site.Entity, err)
	}
	return Agenda{
		Name:    site.Entity + "." + site.OutExt,
		Content: resp,
	}, nil
}

func (f DefaultFetcher) agendaURL(ctx context.Context, site SimpleSite) (string, error) {
	resp, err := f.get(ctx, site.BaseURL)
	if err != nil {
		return "", fmt.Errorf("get page: %s", err)
	}
	defer resp.Close()

	basePage, err := ioutil.ReadAll(resp)
	if err != nil {
		return "", fmt.Errorf("readall: %s", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(basePage))
	if err != nil {
		return "", fmt.Errorf("baseURL: %s", err)
	}

	agendaURL, err := site.Finder(doc, f.today)
	if err != nil {
		return "", NewFinderError(err, basePage, site.Entity)
	}

	return agendaURL, nil
}

func (f DefaultFetcher) get(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Agenda-Get/3.0")
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
	return resp.Body, nil
}
