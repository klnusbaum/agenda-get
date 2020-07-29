package sites

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

var singleQuoteMatcher = regexp.MustCompile("'(.+?)'")

var Sites = []Site{
	simpleSite{
		"oakland",
		"https://www.oaklandca.gov/boards-commissions/planning-commission/meetings",
		func(doc *goquery.Document) (string, error) {
			link, found := doc.
				Find("#meetings").
				Find("tbody").
				Find("tr").
				First().
				Find("td").
				Eq(4).
				Find("a").
				Attr("href")
			if !found {
				return "", errors.New("couldn't find href attribute")
			}
			return link, nil
		},
	},
	simpleSite{
		"bakersfield",
		"https://bakersfield.novusagenda.com/AgendaPublic/?MeetingType=6",
		func(doc *goquery.Document) (string, error) {
			js, found := doc.
				Find("#ctl00_ContentPlaceHolder1_SearchAgendasMeetings_radGridMeetings_ctl00").
				Find("tbody").
				Find("tr").
				First().
				Find("td").
				Eq(4).
				Find("a").
				Attr("onclick")
			if !found {
				return "", errors.New("couldn't find onclick attribute")
			}

			matches := singleQuoteMatcher.FindStringSubmatch(js)
			if len(matches) < 2 {
				return "", fmt.Errorf("couldn't parse url from javascript: %s", js)
			}

			return "https://bakersfield.novusagenda.com/AgendaPublic/" + matches[1], nil
		},
	},
	simpleSite{
		"fresno",
		"https://fresno.legistar.com/DepartmentDetail.aspx?ID=24452&GUID=26F8DAF5-AC08-46BE-A9E4-EC0C6DDC0F66&Search=",
		func(doc *goquery.Document) (string, error) {
			link, found := doc.
				Find("#ctl00_ContentPlaceHolder1_gridCalendar_ctl00").
				Find("tbody").
				Find("tr").
				First().
				Find("td").
				Eq(5).
				Find("a").
				Attr("href")
			if !found {
				return "", errors.New("couldn't find href attribute")
			}

			return "https://fresno.legistar.com/" + link, nil
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
	finder  func(*goquery.Document) (string, error)
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

	agendaURL, err := s.finder(doc)
	if err != nil {
		return "", fmt.Errorf("finder: %s", err)
	}

	return agendaURL, nil
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
