package sites

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/PuerkitoBio/goquery"
)

var Sites = []Site{
	simpleSite{
		"oakland",
		"https://www.oaklandca.gov/boards-commissions/planning-commission/meetings",
	},
}

type Site interface {
	Get(ctx context.Context, outDir string) error
}

type simpleSite struct {
	city    string
	baseURL string
}

func (s simpleSite) Get(ctx context.Context, outDir string) error {
	docURL, err := s.docURL(ctx)
	if err != nil {
		return s.siteErr(err)
	}
	if err := s.getDoc(ctx, docURL, outDir); err != nil {
		return s.siteErr(err)
	}
	return nil
}

func (s simpleSite) docURL(ctx context.Context) (string, error) {
	resp, err := get(ctx, s.baseURL)
	if err != nil {
		return "", fmt.Errorf("get page: %s", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("baseURL: %s", err)
	}

	latestMeeting, found := doc.
		Find("#meetings").
		Find("tbody").
		Find("tr").
		First().
		Find("td").
		Eq(4).
		Find("a").
		Attr("href")

	if !found {
		html, _ := goquery.OuterHtml(doc.Selection)
		return "", fmt.Errorf("no doc found: %s", html)
	}

	return latestMeeting, nil
}

func (s simpleSite) getDoc(ctx context.Context, docURL, outDir string) error {
	resp, err := get(ctx, docURL)
	if err != nil {
		return fmt.Errorf("get agenda %s", err)
	}

	defer resp.Body.Close()
	output, err := os.Create(path.Join(outDir, s.city))
	if err != nil {
		return fmt.Errorf("create output: %s", err)
	}

	if _, err := io.Copy(output, resp.Body); err != nil {
		return fmt.Errorf("write output: %s", err)
	}

	return nil
}

func (s simpleSite) siteErr(err error) error {
	return fmt.Errorf("%s: %s\n", s.city, err)
}

func get(ctx context.Context, url string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "Agenda-Get/3.0")
	return client.Do(req)
}
