package sites

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
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
	outExt  string
	finder  func(*goquery.Document, time.Time) (string, error)
}

type Agenda struct {
	Name    string
	Content io.ReadCloser
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
		Name:    s.entity + "." + s.outExt,
		Content: resp,
	}, nil
}

func (s SimpleSite) agendaURL(ctx context.Context, client HTTPClient, today time.Time) (string, error) {
	resp, err := get(ctx, client, s.baseURL)
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

	agendaURL, err := s.finder(doc, today)
	if err != nil {
		return "", NewFinderError(err, basePage, s.entity)
	}

	return agendaURL, nil
}

func (s SimpleSite) siteErr(err error) error {
	return fmt.Errorf("%s: %w\n", s.entity, err)
}

func get(ctx context.Context, client HTTPClient, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "Agenda-Get/3.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
	return resp.Body, nil
}
